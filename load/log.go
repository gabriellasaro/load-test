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
	"github.com/gabriellasaro/load-test/logwriter"
	"os"
)

type logByLoop struct {
	loop    int
	folder  string
	history *logwriter.LogWriter
}

func newLogByLoop(logFolder string, loop int) *logByLoop {
	return &logByLoop{
		loop:   loop,
		folder: logFolder,
	}
}

func (lw *logByLoop) logDisabled() bool {
	return lw == nil
}

func (lw *logByLoop) newLogHistory() {
	if !lw.logDisabled() {
		lw.history = logwriter.NewLogWriter(lw.folder + fmt.Sprintf("/%d.loop.txt", lw.loop))
		lw.history.Writer()
		lw.history.Send(fmt.Sprintf("LOOP [%d]\n", lw.loop))
	}
}

func (lw *logByLoop) sendDataToHistory(data string) {
	if !lw.logDisabled() {
		lw.history.Send(data)
	}
}

func (lw *logByLoop) waitHistory() {
	if !lw.logDisabled() {
		lw.history.Wait()
	}
}

func startLogFolder(folder string) error {
	if err := os.Mkdir(folder, 0775); err != nil {
		return err
	}

	return nil
}
