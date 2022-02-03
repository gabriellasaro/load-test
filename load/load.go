/*
Copyright 2022 Gabriel Lasaro.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gabriellasaro/load-test/logwriter"
	"github.com/gabriellasaro/load-test/metrics"
	"github.com/gabriellasaro/load-test/types"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type DataTest struct {
	Loops     int       `json:"loops"`
	Parallel  int       `json:"parallel"`
	LogFolder types.Str `json:"log"`
	history   *logwriter.LogWriter
	Variables []Variable `json:"variables"`
}

type Cycle struct {
	log   *LogByWorker
	Steps []*Step `json:"cycle"`
}

func (lt *DataTest) totalLoops() int {
	if lt.Loops <= 0 {
		return 1
	}

	return lt.Loops
}

func (lt *DataTest) workersPerLoop() int {
	if lt.Parallel <= 0 {
		return 1
	}

	return lt.Parallel
}

func (lt *DataTest) logFolder() string {
	return lt.LogFolder.TrimSpace().String()
}

func (lt *DataTest) preload() error {
	if lt.Loops < 0 {
		return errors.New("\"loops\" must be greater than zero")
	}

	if lt.Parallel < 0 {
		return errors.New("\"parallel\" must be greater than zero")
	}

	if err := lt.validateVariables(); err != nil {
		return err
	}

	return nil
}

func (lt *DataTest) newLogHistory() {
	if !lt.logDisabled() {
		lt.history = logwriter.NewLogWriter(path.Join(lt.logFolder(), "history.txt"))
		lt.history.Writer()
	}
}

func (lt *DataTest) sendDataToHistory(data string, print bool) {
	if print {
		fmt.Println(data)
	}

	if !lt.logDisabled() {
		lt.history.Send(data)
	}
}

func (lt *DataTest) waitHistory() {
	if !lt.logDisabled() {
		lt.history.Wait()
	}
}

func (lt *DataTest) logDisabled() bool {
	return lt.LogFolder.TrimSpace().IsEmpty()
}

func (lt *DataTest) startLog() error {
	if lt.logDisabled() {
		return nil
	}

	if err := startLogFolder(lt.logFolder()); err != nil {
		return err
	}

	return nil
}

func (lt *DataTest) startLogForLoop(loop int) error {
	if lt.LogFolder.TrimSpace().IsEmpty() {
		return nil
	}

	if err := startLogFolder(path.Join(lt.logFolder(), fmt.Sprintf("%d", loop))); err != nil {
		return err
	}

	return nil
}

func (lt *DataTest) showAveragesOfSteps() {
	lt.sendDataToHistory(
		"\nAVERAGES OF STEPS",
		true,
	)

	for _, at := range durationMetrics.AveragesOfSteps() {
		lt.sendDataToHistory(
			fmt.Sprintf("\tSTEP [%s]: %s", at.Index(), at.Average()),
			true,
		)
	}
}

func (lt *DataTest) showAveragesOfLoopSteps() {
	lt.sendDataToHistory(
		"\nAVERAGES OF LOOP STEPS",
		true,
	)

	for _, at := range durationMetrics.AveragesOfLoopSteps() {
		lt.sendDataToHistory(
			fmt.Sprintf("\tLOOP [%s] | STEP [%s]: %s", at.Loop(), at.Index(), at.Average()),
			true,
		)
	}
}

func (c *Cycle) existsCycles() error {
	if len(c.Steps) == 0 {
		return errors.New("no cycle provided")
	}

	return nil
}

func (c *Cycle) execute(variables []*Variable, loop int) error {
	if err := c.existsCycles(); err != nil {
		return err
	}

	c.log.newLogHistory()
	c.log.sendDataToHistory(fmt.Sprintf("STEPS TO RUN: %d [0-%d]\n", len(c.Steps), len(c.Steps)-1))

	for i, step := range c.Steps {
		if err := c.Steps[i].preload(i, c.log); err != nil {
			return err
		}

		err := step.execute(variables, &c.Steps)
		step.saveResponseDataToLog(step.index, err)
		if err != nil {
			return err
		}

		step.addDuration(loop)
	}

	c.log.waitHistory()

	return nil
}

func (c *Cycle) startLogByWorker(destinationFolder types.Str, loop, worker int) {
	if !destinationFolder.TrimSpace().IsEmpty() {
		c.log = newLogByWorker(path.Join(destinationFolder.TrimSpace().String(), fmt.Sprintf("%d/%d", loop, worker)), loop, worker)

		if err := startLogFolder(c.log.folder); err != nil {
			log.Fatalln(err)
		}
	}
}

var durationMetrics = metrics.NewMetrics()

func Run(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var load DataTest
	if err := json.Unmarshal(content, &load); err != nil {
		return err
	}

	if err := load.preload(); err != nil {
		return err
	}

	variables := load.getVariablesForReplace()

	if err := load.startLog(); err != nil {
		return err
	}

	load.newLogHistory()
	load.sendDataToHistory("HISTORY\n", false)

	loop := 1
	for {
		var wgLoop sync.WaitGroup

		wgLoop.Add(load.workersPerLoop())

		load.sendDataToHistory(fmt.Sprintf("\nLOOP %d\n", loop), true)

		if err := load.startLogForLoop(loop); err != nil {
			return err
		}

		for w := 1; w <= load.workersPerLoop(); w++ {
			var cycle Cycle
			if err := json.Unmarshal(content, &cycle); err != nil {
				return err
			}

			cycle.startLogByWorker(load.LogFolder, loop, w)

			go func(worker int) {
				defer wgLoop.Done()

				err := cycle.execute(variables, loop)
				logTime := time.Now().Format("01-02-2006 15:04:05")

				if err != nil {
					load.sendDataToHistory(
						fmt.Sprintf("%s: LOOP: %d | WORKER: %d | ERROR: %q", logTime, loop, worker, err),
						true,
					)
				} else {
					load.sendDataToHistory(
						fmt.Sprintf("%s: LOOP: %d | WORKER: %d | SUCCESS", logTime, loop, worker),
						true,
					)
				}
			}(w)
		}

		wgLoop.Wait()

		if loop == load.totalLoops() {
			break
		}

		loop += 1
	}

	load.showAveragesOfLoopSteps()
	load.showAveragesOfSteps()
	load.waitHistory()

	return nil
}
