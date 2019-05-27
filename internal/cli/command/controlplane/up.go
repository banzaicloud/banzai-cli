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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

type createOptions struct {
	file string
}

// NewUpCommand creates a new cobra.Command for `banzai clontrolplane up`.
func NewUpCommand() *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"c"},
		Short:   "Create a controlplane",
		Long:    "Create controlplane based on json stdin or interactive session",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUp(options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "values.yaml", "Control Plane descriptor file")

	return cmd
}

func runUp(options createOptions) {
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

		for {
			if bytes, err := json.MarshalIndent(out, "", "  "); err != nil {
				log.Errorf("failed to marshal descriptor: %v", err)
				log.Debugf("descriptor: %#v", out)
			} else {
				content = string(bytes)
				_, _ = fmt.Fprintf(os.Stderr, "The current state of the descriptor:\n\n%s\n", content)
			}

			var open bool
			_ = survey.AskOne(&survey.Confirm{Message: "Do you want to edit the controlplane descriptor in your text editor?"}, &open, nil)
			if !open {
				break
			}

			_ = survey.AskOne(&survey.Editor{Message: "controlplane descriptor:", Default: content, HideDefault: true, AppendDefault: true}, &content, nil)
			if err := json.Unmarshal([]byte(content), &out); err != nil {
				log.Errorf("can't parse descriptor: %v", err)
			}
		}

		var create bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to CREATE the controlplane now?",
				Default: true,
			},
			&create,
			nil,
		)

		if !create {
			log.Fatal("controlplane creation cancelled")
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

	log.Info("controlplane is being created")

	if err := runInternal("apply", filename); err != nil {
		log.Fatalf("controlplane creation failed: %v", err)
	}
}

func isInteractive() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
		return !viper.GetBool("formatting.no-interactive")
	}
	return viper.GetBool("formatting.force-interactive")
}

func runInternal(command, valuesFile string) error {

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	valuesFile, err = filepath.Abs(valuesFile)
	if err != nil {
		return err
	}

	infoCmd := exec.Command("docker", "info", "-f", "{{eq .OperatingSystem \"Docker Desktop\"}}")

	infoOuput, err := infoCmd.Output()
	if err != nil {
		return err
	}

	isDockerForMac := strings.Trim(string(infoOuput), "\n")

	args := []string{
		"run", "-it", "--rm",
		"-v", fmt.Sprintf("%s:/root/.kube/config", kubeconfig),
		"-v", fmt.Sprintf("%s/.tfstate:/tfstate", pwd),
		"-v", fmt.Sprintf("%s:/terraform/values.yaml", valuesFile),
		"-e", fmt.Sprintf("IS_DOCKER_FOR_MAC=%s", isDockerForMac),
		"--entrypoint", "/terraform/entrypoint.sh",
		"banzaicloud/cp-installer:latest",
		command,
		"-state=/tfstate/terraform.tfstate", // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271
		"-parallelism=1",
	}

	log.Infof("docker %v", args)

	cmd := exec.Command("docker", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
