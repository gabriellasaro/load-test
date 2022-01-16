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
}

func (c *Step) applyVariables(variables []*Variable, cycle *[]*Step, data string) (string, error) {
	data = types.ReplaceKeyByValue(variables, data)

	env, err := getEnvironmentVariables(data)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(env, data)

	paths, err := getPathVariables(c.index, data, cycle)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(paths, data)

	resp, err := getResponseVariables(c.index, cycle, data)
	if err != nil {
		return "", err
	}
	data = types.ReplaceKeyByValue(resp, data)

	return data, nil
}

func (c *Step) getURL(variables []*Variable, cycle *[]*Step) (string, error) {
	return c.applyVariables(variables, cycle, c.URL.TrimSpace().String())
}

func (c *Step) getMethod() string {
	return c.Method.TrimSpace().ToUpper().String()
}

func (c *Step) getContentType() string {
	return c.ContentType.ToUpper().String()
}

func (c *Step) preloadBody() error {
	body := c.Body.TrimSpace().String()
	if len(body) > 0 {
		c.preloadedBody = body
	} else if c.BodyJSON != nil {
		content, err := json.Marshal(c.BodyJSON)
		if err != nil {
			return err
		}

		c.preloadedBody = string(content)
	} else if len(c.BodyLoadFile) > 0 {
		content, err := os.ReadFile(c.BodyLoadFile)
		if err != nil {
			return err
		}

		c.preloadedBody = string(content)
	}

	return nil
}

func (c *Step) getBodyReader(variables []*Variable, cycles *[]*Step) (io.Reader, error) {
	body, err := c.applyVariables(variables, cycles, c.preloadedBody)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(body), nil
}

func (c *Step) preload(index int) error {
	c.index = index

	if err := c.preloadIf(); err != nil {
		return err
	}

	if len(c.URL) == 0 {
		return fmt.Errorf("cycle[%d].url cannot be empty", index)
	}

	if len(c.getMethod()) == 0 {
		c.Method = "GET"
	}

	if c.getMethod() != "GET" && c.BodyJSON == nil && c.getContentType() == "" {
		return errors.New("for your request type it is necessary to inform the content_type")
	} else if c.getContentType() == "" && c.BodyJSON != nil {
		c.ContentType = "application/json"
	}

	err := c.preloadBody()
	if err != nil {
		return err
	}

	return nil
}

func (c *Step) execute(variables []*Variable, cycles *[]*Step) error {
	if err := c.executeIf(variables, cycles); err != nil {
		return fmt.Errorf("condition (%s) is not satisfied: %s", c.ConditionRaw, err.Error())
	}

	url, err := c.getURL(variables, cycles)
	if err != nil {
		return err
	}

	body, err := c.getBodyReader(variables, cycles)
	if err != nil {
		return err
	}

	timeStart := time.Now()

	req, err := http.NewRequest(c.getMethod(), url, body)
	if err != nil {
		return fmt.Errorf("cycle[%d]: %s", c.index, err.Error())
	}

	if c.getContentType() != "" {
		req.Header.Set("Content-Type", c.getContentType())
	}

	for _, item := range c.Header {
		req.Header.Set(item.Key(), item.Value())
	}

	client := &http.Client{}
	if c.Timeout != nil {
		client.Timeout = time.Second * (*c.Timeout)
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
	c.response = responseCycle

	return nil
}

func getStepByIndex(cycle *[]*Step, index int) (*Step, error) {
	for _, step := range *cycle {
		if step.index == index {
			return step, nil
		}
	}

	return nil, fmt.Errorf("cycle[%d] not found", index)
}
