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

package controlplane

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

// NewDownCommand creates a new cobra.Command for `banzai clontrolplane down`.
func NewDownCommand() *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy the controlplane",
		Long:  "Destroy a controlplane based on json stdin or interactive session",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runDestroy(options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Control Plane descriptor file")

	return cmd
}

func runDestroy(options createOptions) {
	var out map[string]interface{}

	filename := options.file

	if isInteractive() {
		var content string

		for {
			if filename == "" {
				_ = survey.AskOne(
					&survey.Input{
						Message: "Load a JSON or YAML file:",
						Default: "values.yaml",
						Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML Control Plane creation descriptor. Leave empty to cancel.",
					},
					&filename,
					nil,
				)
				if filename == "skip" || filename == "" {
					break
				}
			}

			if raw, err := ioutil.ReadFile(filename); err != nil {

				log.Errorf("failed to read file %q: %v", filename, err)

				filename = "" // reset fileName so that we can ask for one

				continue
			} else {
				if err := unmarshal(raw, &out); err != nil {
					log.Fatalf("failed to parse control plane descriptor: %v", err)
				}

				break
			}
		}

		if bytes, err := json.MarshalIndent(out, "", "  "); err != nil {
			log.Debugf("descriptor: %#v", out)
			log.Fatalf("failed to marshal descriptor: %v", err)
		} else {
			content = string(bytes)
			_, _ = fmt.Fprintf(os.Stderr, "The current state of the descriptor:\n\n%s\n", content)
		}

		var destroy bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to DESTROY the controlplane now?",
				Default: true,
			},
			&destroy,
			nil,
		)

		if !destroy {
			log.Fatal("controlplane destroy cancelled")
		}
	} else { // non-interactive
		var raw []byte
		var err error

		if filename != "" {
			raw, err = ioutil.ReadFile(filename)
		} else {
			raw, err = ioutil.ReadAll(os.Stdin)
			filename = "stdin"
		}

		if err != nil {
			log.Fatalf("failed to read %s: %v", filename, err)
		}

		if err := unmarshal(raw, &out); err != nil {
			log.Fatalf("failed to parse controlplane descriptor: %v", err)
		}
	}

	log.Info("controlplane is being destroy")

	if err := runInternal("destroy", filename); err != nil {
		log.Fatalf("controlplane destroy failed: %v", err)
	}
}
