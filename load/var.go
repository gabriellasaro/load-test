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
	"fmt"
	"strings"
)

type Variable struct {
	Variable string `json:"key"`
	Data     string `json:"value"`
	F        []string
}

func (v *Variable) Key() string {
	return v.Variable
}

func (v *Variable) Value() string {
	return v.Data
}

func (lt *DataTest) validateVariables() error {
	for i, v := range lt.Variables {
		if strings.TrimSpace(v.Key()) == "" {
			return fmt.Errorf("variables[%d].key does not have a valid value", i)
		}
	}

	return nil
}

func (lt *DataTest) getVariablesForReplace() []*Variable {
	variablesForReplace := make([]*Variable, len(lt.Variables))

	for i, v := range lt.Variables {
		variablesForReplace[i] = &Variable{
			Variable: "{%VAR:" + v.Key() + ":ENDVAR%}",
			Data:     v.Value(),
		}
	}

	return variablesForReplace
}
