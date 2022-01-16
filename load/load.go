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
	"log"
	"os"
	"sync"
)

type DataTest struct {
	Loops     int        `json:"loops"`
	Parallel  int        `json:"parallel"`
	Variables []Variable `json:"variables"`
}

type Cycle struct {
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

func (c *Cycle) existsCycles() error {
	if len(c.Steps) == 0 {
		return errors.New("no cycle provided")
	}

	return nil
}

func (c *Cycle) execute(variables []*Variable) error {
	if err := c.existsCycles(); err != nil {
		return err
	}

	for i, step := range c.Steps {
		if err := c.Steps[i].preload(i); err != nil {
			return err
		}

		if err := step.execute(variables, &c.Steps); err != nil {
			return err
		}
	}

	return nil
}

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

	var wg sync.WaitGroup

	loop := 1
	for {
		wg.Add(load.workersPerLoop())

		for w := 1; w <= load.workersPerLoop(); w++ {
			var cycle Cycle
			if err := json.Unmarshal(content, &cycle); err != nil {
				return err
			}

			go func(worker int) {
				defer wg.Done()
				err := cycle.execute(variables)
				if err != nil {
					log.Printf("GROUP: %d | WORKER: %d | ERROR: %q", loop, worker, err)
				} else {
					log.Printf("GROUP: %d | WORKER: %d | SUCCESS", loop, worker)
				}
			}(w)
		}

		wg.Wait()

		if loop == load.totalLoops() {
			break
		}

		loop += 1
	}

	return nil
}
