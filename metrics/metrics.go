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

package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type AverageTime struct {
	loop    string
	index   string
	average time.Duration
}

func (at *AverageTime) Loop() string {
	return at.loop
}

func (at *AverageTime) Index() string {
	return at.index
}

func (at *AverageTime) Average() time.Duration {
	return at.average
}

type Metrics struct {
	mutex             sync.Mutex
	durationStepsLoop map[string]int64
	durationSteps     map[string]int64
	totalStepsLoop    map[string]int64
	totalSteps        map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		durationStepsLoop: make(map[string]int64),
		durationSteps:     make(map[string]int64),
		totalStepsLoop:    make(map[string]int64),
		totalSteps:        make(map[string]int64),
	}
}

func (m *Metrics) keyLoopAndIndex(worker, index int) string {
	return fmt.Sprintf("%d-%d", worker, index)
}

func (m *Metrics) keyIndex(index int) string {
	return fmt.Sprintf("%d", index)
}

func (m *Metrics) AddDuration(loop, index int, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.durationStepsLoop[m.keyLoopAndIndex(loop, index)] += value
	m.durationSteps[m.keyIndex(index)] += value

	m.totalStepsLoop[m.keyLoopAndIndex(loop, index)]++
	m.totalSteps[m.keyIndex(index)]++
}

func (m *Metrics) averagesOfLoopSteps(key string) time.Duration {
	return m.average(
		key,
		m.durationStepsLoop,
		m.totalStepsLoop,
	)
}

func (m *Metrics) averagesOfStep(key string) time.Duration {
	return m.average(
		key,
		m.durationSteps,
		m.totalSteps,
	)
}

func (m *Metrics) average(key string, values, totals map[string]int64) time.Duration {
	return time.Duration(float64(values[key])/float64(totals[key])) * time.Millisecond
}

func (m *Metrics) AveragesOfLoopSteps() []AverageTime {
	averages := make([]AverageTime, 0)

	for key := range m.durationStepsLoop {
		parts := strings.Split(key, "-")

		averages = append(averages, AverageTime{
			loop:    parts[0],
			index:   parts[1],
			average: m.averagesOfLoopSteps(key),
		})
	}

	return averages
}

func (m *Metrics) AveragesOfSteps() []AverageTime {
	averages := make([]AverageTime, 0)

	for key := range m.durationSteps {
		averages = append(averages, AverageTime{
			index:   key,
			average: m.averagesOfStep(key),
		})
	}

	return averages
}
