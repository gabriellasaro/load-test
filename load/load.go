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

func (lt *DataTest) startLogForLoop(loop int) (*logByLoop, error) {
	if lt.LogFolder.TrimSpace().IsEmpty() {
		return nil, nil
	}

	logLoop := newLogByLoop(lt.logFolder(), loop)
	logLoop.newLogHistory()

	return logLoop, nil
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

		logLoop, err := load.startLogForLoop(loop)
		if err != nil {
			return err
		}

		for w := 1; w <= load.workersPerLoop(); w++ {
			var cycle Cycle
			if err := json.Unmarshal(content, &cycle); err != nil {
				return err
			}

			go func(worker int) {
				defer wgLoop.Done()

				err := cycle.execute(variables, loop, worker, logLoop)
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
		logLoop.waitHistory()

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
