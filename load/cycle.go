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
	"errors"
	"fmt"
)

type Cycle struct {
	Steps []*Step `json:"cycle"`
}

func (c *Cycle) existsCycles() error {
	if len(c.Steps) == 0 {
		return errors.New("no cycle provided")
	}

	return nil
}

func (c *Cycle) execute(variables []*Variable, loop, worker int, logLoop *logByLoop) error {
	if err := c.existsCycles(); err != nil {
		return err
	}

	logCycle := fmt.Sprintf("----------------\n\nWORKER [%d] | STEPS TO RUN: %d [0-%d]\n", worker, len(c.Steps), len(c.Steps)-1)

	for i, step := range c.Steps {
		if err := c.Steps[i].preload(i); err != nil {
			logCycle += step.responseDataToLog(step.index, err)
			logLoop.sendDataToHistory(logCycle)

			return err
		}

		err := step.execute(variables, &c.Steps)
		logCycle += step.responseDataToLog(step.index, err)
		if err != nil {
			logLoop.sendDataToHistory(logCycle)

			return err
		}

		step.addDuration(loop)
	}

	logLoop.sendDataToHistory(logCycle)

	return nil
}
