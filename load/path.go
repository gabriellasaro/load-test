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
	"github.com/gabriellasaro/load-test/types"
	"regexp"
	"strings"
)

var regexVarPath = regexp.MustCompile(`{%PATH\[([0-9]+)\]:([^{%}]+):ENDPATH%}`)

func setValue(indexVar *types.IndexVariable, currentCycle int, cycle *[]*Step) error {
	if indexVar.Index() >= currentCycle {
		return fmt.Errorf("cannot use a variable that does not yet exist: %s", indexVar.Key())
	}

	step, err := getStepByIndex(cycle, indexVar.Index())
	if err != nil {
		return err
	}

	p := strings.Split(indexVar.Path(), ".")

	value, err := step.response.getValueInResponseByPath(p)
	if err != nil {
		return err
	}

	indexVar.SetValue(value)

	return nil
}

func getPathVariables(currentCycle int, sourceVariable string, cycle *[]*Step) ([]*types.IndexVariable, error) {
	paths, err := types.GetMatchesForIndexVariables(regexVarPath, sourceVariable)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(paths); i++ {
		if err := setValue(paths[i], currentCycle, cycle); err != nil {
			return nil, err
		}
	}

	return paths, nil
}
