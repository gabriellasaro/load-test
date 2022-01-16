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
	"os"
	"regexp"
)

var regexVarEnv = regexp.MustCompile(`{%ENV:([^ ]+):ENDENV%}`)

func lookupEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable (%s) not found", key)
	}

	return value, nil
}

func getEnvironmentVariables(data string) ([]*types.Variable, error) {
	return types.GetVariables(regexVarEnv, data, lookupEnv)
}
