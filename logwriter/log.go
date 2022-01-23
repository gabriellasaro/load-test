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

package logwriter

import (
	"log"
	"os"
	"sync"
)

type LogWriter struct {
	wg       *sync.WaitGroup
	write    chan string
	filename string
}

func NewLogWriter(filename string) *LogWriter {
	return &LogWriter{
		wg:       &sync.WaitGroup{},
		write:    make(chan string),
		filename: filename,
	}
}

func (w *LogWriter) Writer() {
	w.wg.Add(1)
	go w.writer()
}

func (w *LogWriter) writer() {
	defer w.wg.Done()

	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	for line := range w.write {
		if _, err := f.WriteString(line + "\n"); err != nil {
			_ = f.Close()
			log.Fatal(err)
		}
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func (w *LogWriter) Send(data string) {
	w.write <- data
}

func (w *LogWriter) Wait() {
	close(w.write)
	w.wg.Wait()
}
