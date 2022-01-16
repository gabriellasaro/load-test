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

package main

import (
	"fmt"
	"github.com/gabriellasaro/load-test/load"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Informe o arquivo para ser executado: %s filename.json", os.Args[0])
	} else {
		for _, filename := range os.Args[1:] {
			err := load.Run(filename)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
