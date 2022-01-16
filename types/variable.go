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

package types

import (
	"fmt"
	"regexp"
	"strconv"
)

var regexVarEnv = regexp.MustCompile(`{%ENV:([^ ]+):ENDENV%}`)

type Variable struct {
	rawVariable string
	variable    string
	value       string
}

type IndexVariable struct {
	Variable
	index int
}

func (v *Variable) Key() string {
	return v.rawVariable
}

func (v *Variable) Value() string {
	return v.value
}

func (v *IndexVariable) Path() string {
	return v.variable
}

func (v *IndexVariable) Index() int {
	return v.index
}

func (v *IndexVariable) SetValue(value string) {
	v.value = value
}

func GetVariables(re *regexp.Regexp, data string, getValue func(string) (string, error)) ([]*Variable, error) {
	vars := make([]*Variable, 0)

	matches := re.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		if len(match) != 2 {
			return nil, fmt.Errorf("a possible error occurred while fetching the variable: %v", match)
		}

		value, err := getValue(match[1])
		if err != nil {
			return nil, err
		}

		vars = append(vars, &Variable{
			rawVariable: match[0],
			variable:    match[1],
			value:       value,
		})
	}

	return vars, nil
}

func GetMatchesForIndexVariables(re *regexp.Regexp, data string) ([]*IndexVariable, error) {
	paths := make([]*IndexVariable, 0)

	matches := re.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		if len(match) != 3 {
			return nil, fmt.Errorf("a possible error occurred while fetching the variable: %v", match)
		}

		path := new(IndexVariable)

		index, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("an error occurred while extracting the variable index: %s", match)
		}

		path.rawVariable = match[0]
		path.index = index
		path.variable = match[2]

		paths = append(paths, path)
	}

	return paths, nil
}
