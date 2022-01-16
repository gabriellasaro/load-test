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
	"github.com/gabriellasaro/load-test/types"
	"regexp"
	"strconv"
	"time"
)

var regexVarResp = regexp.MustCompile(`{%RESP\[([0-9]+)\]:([A-Z_]+):ENDRESP%}`)

type ResponseCycle struct {
	StatusCode int
	Body       []byte
	Duration   time.Duration
}

func (r *ResponseCycle) bodyToInterface() interface{} {
	if r.Body == nil {
		return nil
	}

	var body map[string]interface{}

	if err := json.Unmarshal(r.Body, &body); err != nil {
		var body []map[string]interface{}

		if err := json.Unmarshal(r.Body, &body); err != nil {
			return err
		}

		return body
	}

	return body
}

func (r *ResponseCycle) getValueInResponseByPath(path []string) (string, error) {
	value, err := getValueInResponse(r.bodyToInterface(), path)
	if err != nil {
		return "", fmt.Errorf("%s: %q", err.Error(), path)
	}

	return value, nil
}

func getValueInResponse(body interface{}, path []string) (string, error) {
	if body, ok := body.(map[string]interface{}); ok {
		if value, found := body[path[0]]; found {
			if len(path) == 1 {
				return fmt.Sprintf("%v", value), nil
			}

			return getValueInResponse(body[path[0]], path[1:])
		}
	}

	if len(path) >= 2 {
		if body, ok := body.([]interface{}); ok {
			key, err := strconv.Atoi(path[0])
			if err != nil {
				return "", err
			}

			if key > (len(body) - 1) {
				return "", fmt.Errorf("the index [%d] provided does not exist", key)
			}

			if body, ok := body[key].(map[string]interface{}); ok {
				if value, found := body[path[1]]; found {
					if len(path) == 2 {
						return fmt.Sprintf("%v", value), nil
					}

					if body, ok := value.(map[string]interface{}); ok {
						return getValueInResponse(body, path[2:])
					}
				}
			}
		}
	}

	return "", errors.New("the path was not found")
}

func (r *ResponseCycle) getValueInResponseVariable(key string) (string, error) {
	switch key {
	case "STATUS_CODE":
		return fmt.Sprintf("%d", r.StatusCode), nil
	default:
		return "", fmt.Errorf("the variable is not valid: %s", key)
	}
}

func getResponseVariables(currentCycle int, cycle *[]*Step, data string) ([]*types.IndexVariable, error) {
	indexVars, err := types.GetMatchesForIndexVariables(regexVarResp, data)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(indexVars); i++ {
		if indexVars[i].Index() >= currentCycle {
			return nil, fmt.Errorf("cannot use a variable that does not yet exist: %s", indexVars[i].Key())
		}

		step, err := getStepByIndex(cycle, indexVars[i].Index())
		if err != nil {
			return nil, err
		}

		value, err := step.response.getValueInResponseVariable(indexVars[i].Path())
		if err != nil {
			return nil, err
		}

		indexVars[i].SetValue(value)
	}

	return indexVars, nil
}
