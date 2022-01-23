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
	"strings"
)

type Condition struct {
	equalityOperator string
	arguments        [2]string
	values           [2]string
}

func (c *Condition) applyCondition() bool {
	switch c.equalityOperator {
	case "==":
		return c.values[0] == c.values[1]
	case "!=":
		return c.values[0] != c.values[1]
	default:
		return false
	}
}

func validateEqualityOperator(op string) (operator string, err error) {
	switch op {
	case "==", "!=":
		operator = op
	default:
		err = fmt.Errorf("operator (%s) is not valid", op)
	}

	return
}

func (s *Step) getArgsOfCondition() []string {
	args := make([]string, 0)
	argsRaw := strings.Split(s.ConditionRaw.TrimSpace().String(), " ")

	var (
		rightArgument string
		argumentFound int
	)

	for _, argRaw := range argsRaw {
		if argumentFound == 2 {
			rightArgument += " " + argRaw
			continue
		}

		arg := strings.TrimSpace(argRaw)
		if arg != "" {
			args = append(args, strings.TrimSpace(arg))
			argumentFound++
		}
	}

	if rightArgument != "" {
		args = append(args, strings.TrimSpace(rightArgument))
	}

	return args
}

func (s *Step) preloadIf() error {
	if s.ConditionRaw == nil {
		return nil
	}

	if s.index == 0 {
		return errors.New("cannot use if statement in cycle[0]")
	}

	cond := s.getArgsOfCondition()
	if len(cond) != 3 {
		return fmt.Errorf("the condition (%s) must be composed of 3 parameters", s.ConditionRaw)
	}

	s.condition = new(Condition)

	operador, err := validateEqualityOperator(cond[0])
	if err != nil {
		return err
	}

	s.condition.equalityOperator = operador
	s.condition.arguments[0] = cond[1]
	s.condition.arguments[1] = cond[2]

	return nil
}

func (s *Step) setValuesForCondition(variables []*Variable, cycle *[]*Step) error {
	value0, err := s.applyVariables(variables, cycle, s.condition.arguments[0])
	if err != nil {
		return err
	}
	s.condition.values[0] = value0

	value1, err := s.applyVariables(variables, cycle, s.condition.arguments[1])
	if err != nil {
		return err
	}
	s.condition.values[1] = value1

	return nil
}

func (s *Step) executeIf(variables []*Variable, cycle *[]*Step) error {
	if s.condition == nil {
		return nil
	}

	if err := s.setValuesForCondition(variables, cycle); err != nil {
		return err
	}

	if apply := s.condition.applyCondition(); !apply {
		return fmt.Errorf("(%s %s %s) -> false", s.condition.values[0], s.condition.equalityOperator, s.condition.values[1])
	}

	return nil
}
