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
	"log"
	"os"
	"path"
	"strings"
)

type LogByWorker struct {
	loop    int
	worker  int
	folder  string
	history *logwriter.LogWriter
}

func newLogByWorker(destinationFolder string, loop, worker int) *LogByWorker {
	return &LogByWorker{
		loop:   loop,
		worker: worker,
		folder: destinationFolder,
	}
}

func (lw *LogByWorker) logDisabled() bool {
	return lw == nil
}

func (lw *LogByWorker) pathToBodyFolder(index int) string {
	return fmt.Sprintf("body/%d", index)
}

func (lw *LogByWorker) createFolderForStep(name string) {
	if err := os.MkdirAll(path.Join(lw.folder, name), 0775); err != nil {
		log.Fatalln(err)
	}
}

func (lw *LogByWorker) saveBody(index int, data []byte, contentType string) {
	if !lw.logDisabled() {
		ext := ".txt"
		if strings.Contains(contentType, "json") {
			ext = ".json"
		}

		lw.createFolderForStep(lw.pathToBodyFolder(index))

		if err := os.WriteFile(path.Join(lw.folder, lw.pathToBodyFolder(index), "response-body"+ext), data, 0666); err != nil {
			log.Fatalln(err)
		}
	}
}

func (lw *LogByWorker) newLogHistory() {
	if !lw.logDisabled() {
		lw.history = logwriter.NewLogWriter(path.Join(lw.folder, "history.txt"))
		lw.history.Writer()
	}
}

func (lw *LogByWorker) sendDataToHistory(data string) {
	if !lw.logDisabled() {
		lw.history.Send(data)
	}
}

func (lw *LogByWorker) waitHistory() {
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
