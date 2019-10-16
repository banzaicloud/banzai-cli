// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// Modified version of https://github.com/moul/http2curl by Manfred Touron

// CurlCommand contains exec.Command compatible slice + helpers
type CurlCommand []string

// append appends a string to the CurlCommand
func (c *CurlCommand) append(newSlice ...string) {
	*c = append(*c, newSlice...)
}

// String returns a ready to copy/paste command
func (c *CurlCommand) String() string {
	return strings.Join(*c, " ")
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func GetCurlCommand(req *http.Request, insecureSkipVerify bool) (*CurlCommand, error) {
	command := CurlCommand{}

	command.append("curl")

	if insecureSkipVerify {
		command.append("-k")
	}

	command.append("-X", bashEscape(req.Method))

	keys := make([]string, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	command.append(bashEscape(req.URL.String()))

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		bodyEscaped := bashEscape(string(body))
		command.append("-d", bodyEscaped)
	}

	return &command, nil
}
