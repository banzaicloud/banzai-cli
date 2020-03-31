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

// The code is taken from the kind project and slightly modified.

package spinner

import (
	"fmt"
	"os"
)

// Status is used to track ongoing status in a CLI, with a nice loading spinner
// when attached to a terminal
type Status struct {
	spinner *Spinner
	status  string
	// for controlling coloring etc
	successFormat string
	failureFormat string
}

// NewStatus returns a new status object
func NewStatus() *Status {
	s := &Status{
		spinner:       NewSpinner(os.Stdout),
		successFormat: " \x1b[32m✓\x1b[0m %s\n",
		failureFormat: " \x1b[31m✗\x1b[0m %s\n",
	}
	return s
}

// Start starts a new phase of the status, if attached to a terminal
// there will be a loading spinner with this status
func (s *Status) Start(status string) {
	s.End(true)
	// set new status
	s.status = status
	if s.spinner != nil {
		s.spinner.SetSuffix(fmt.Sprintf(" %s ", s.status))
		s.spinner.Start()
	} else {
		fmt.Printf(" • %s  ...\n", s.status)
	}
}

// End completes the current status, ending any previous spinning and
// marking the status as success or failure
func (s *Status) End(success bool) {
	if s.status == "" {
		return
	}

	if s.spinner != nil {
		s.spinner.Stop()
		fmt.Fprint(s.spinner.writer, "\r")
	}
	if success {
		fmt.Printf(s.successFormat, s.status)
	} else {
		fmt.Printf(s.failureFormat, s.status)
	}

	s.status = ""
}
