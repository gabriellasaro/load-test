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

import "strings"

type KeyValue interface {
	Key() string
	Value() string
}

func ReplaceKeyByValue[T KeyValue](variables []T, data string) string {
	if len(variables) == 0 {
		return data
	}

	for _, v := range variables {
		data = strings.ReplaceAll(data, v.Key(), v.Value())
	}

	return data
}
