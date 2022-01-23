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
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Step struct {
	ConditionRaw  *types.Str `json:"if"`
	condition     *Condition
	URL           types.Str      `json:"url"`
	ContentType   types.Str      `json:"content_type"`
	Method        types.Str      `json:"method"`
	Header        []Variable     `json:"header"`
	Timeout       *time.Duration `json:"timeout"`
	BodyJSON      interface{}    `json:"body_json"`
	Body          types.Str      `json:"body"`
	BodyLoadFile  string         `json:"body_load_file"`
	index         int
	preloadedBody string
	response      *ResponseCycle
	log           *LogByWorker
}

func (s *Step) applyVariables(variables []*Variable, cycle *[]*Step, data string) (string, error) {
	data = types.ReplaceKeyByValue(variables, data)

	env, err := getEnvironmentVariables(data)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(env, data)

	paths, err := getPathVariables(s.index, data, cycle)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(paths, data)

	resp, err := getResponseVariables(s.index, cycle, data)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(resp, data)

	return data, nil
}

func (s *Step) getURL(variables []*Variable, cycle *[]*Step) (string, error) {
	return s.applyVariables(variables, cycle, s.URL.TrimSpace().String())
}

func (s *Step) getMethod() string {
	return s.Method.TrimSpace().ToUpper().String()
}

func (s *Step) getContentType() string {
	return s.ContentType.ToUpper().String()
}

func (s *Step) preloadBody() error {
	body := s.Body.TrimSpace().String()
	if len(body) > 0 {
		s.preloadedBody = body
	} else if s.BodyJSON != nil {
		content, err := json.Marshal(s.BodyJSON)
		if err != nil {
			return err
		}

		s.preloadedBody = string(content)
	} else if len(s.BodyLoadFile) > 0 {
		content, err := os.ReadFile(s.BodyLoadFile)
		if err != nil {
			return err
		}

		s.preloadedBody = string(content)
	}

	return nil
}

func (s *Step) getBodyReader(variables []*Variable, cycles *[]*Step) (io.Reader, error) {
	body, err := s.applyVariables(variables, cycles, s.preloadedBody)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(body), nil
}

func (s *Step) preload(index int, log *LogByWorker) error {
	s.index = index

	if err := s.preloadIf(); err != nil {
		return err
	}

	if len(s.URL) == 0 {
		return fmt.Errorf("cycle[%d].url cannot be empty", index)
	}

	if len(s.getMethod()) == 0 {
		s.Method = "GET"
	}

	if s.getMethod() != "GET" && s.BodyJSON == nil && s.getContentType() == "" {
		return errors.New("for your request type it is necessary to inform the content_type")
	} else if s.getContentType() == "" && s.BodyJSON != nil {
		s.ContentType = "application/json"
	}

	err := s.preloadBody()
	if err != nil {
		return err
	}

	s.setLog(log)

	return nil
}

func (s *Step) execute(variables []*Variable, cycles *[]*Step) error {
	if err := s.executeIf(variables, cycles); err != nil {
		return fmt.Errorf("condition (%s) is not satisfied: %s", s.ConditionRaw, err.Error())
	}

	url, err := s.getURL(variables, cycles)
	if err != nil {
		return err
	}

	body, err := s.getBodyReader(variables, cycles)
	if err != nil {
		return err
	}

	timeStart := time.Now()

	req, err := http.NewRequest(s.getMethod(), url, body)
	if err != nil {
		return fmt.Errorf("cycle[%d]: %s", s.index, err.Error())
	}

	if s.getContentType() != "" {
		req.Header.Set("Content-Type", s.getContentType())
	}

	for _, item := range s.Header {
		req.Header.Set(item.Key(), item.Value())
	}

	client := &http.Client{}
	if s.Timeout != nil {
		client.Timeout = time.Second * (*s.Timeout)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseCycle := new(ResponseCycle)
	responseCycle.StatusCode = resp.StatusCode

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseCycle.Body = responseBody
	responseCycle.Duration = time.Since(timeStart)
	s.response = responseCycle

	s.log.saveBody(s.index, responseBody, resp.Header.Get("content-type"))

	return nil
}

func (s *Step) setLog(log *LogByWorker) {
	if log != nil {
		s.log = log
	}
}

func getStepByIndex(cycle *[]*Step, index int) (*Step, error) {
	for _, step := range *cycle {
		if step.index == index {
			return step, nil
		}
	}

	return nil, fmt.Errorf("cycle[%d] not found", index)
}
