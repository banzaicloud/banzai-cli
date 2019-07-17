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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

type createOptions struct {
	file string
	controlPlaneInstallerOptions
}

// NewUpCommand creates a new cobra.Command for `banzai controlplane up`.
func NewUpCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"c"},
		Short:   "Create a controlplane",
		Long:    "Create controlplane based on json stdin or interactive session in the current Kubernetes context. The current working directory will be used for storing the applied configuration and deployment status.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUp(options, banzaiCli)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", valuesDefault, "Input control plane descriptor file")

	bindInstallerFlags(flags, &options.controlPlaneInstallerOptions)

	return cmd
}

func runUp(options createOptions, banzaiCli cli.Cli) error {
	var out map[string]interface{}

	filename := options.file

	if isInteractive() {
		var content string

		for {
			if filename == "" {
				_ = survey.AskOne(
					&survey.Input{
						Message: "Load a JSON or YAML file:",
						Default: valuesDefault,
						Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML control plane descriptor. Leave empty to cancel.",
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
				if err := utils.Unmarshal(raw, &out); err != nil {
					return emperror.Wrap(err, "failed to parse control plane descriptor")
				}

				break
			}
		}

		if bytes, err := json.MarshalIndent(out, "", "  "); err != nil {
			return emperror.Wrapf(err, "failed to marshal descriptor")
		} else {
			content = string(bytes)
			_, _ = fmt.Fprintf(os.Stderr, "The current state of the descriptor:\n\n%s\n", content)
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
			return errors.New("controlplane creation cancelled")
		}
	} else { // non-interactive
		filename, raw, err := utils.ReadFileOrStdin(filename)
		if err != nil {
			return emperror.WrapWith(err, "failed to read", "filename", filename)
		}

		if err := utils.Unmarshal(raw, &out); err != nil {
			return emperror.Wrap(err, "failed to parse controlplane descriptor")
		}
	}

	kindCluster := isKINDClusterRequested(out)
	if kindCluster {
		err := ensureKINDCluster(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "failed to create KIND cluster")
		}
	}

	// create temp dir for the files to attach
	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		return emperror.Wrap(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(dir)

	// write values to temp file
	values, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return emperror.Wrap(err, "failed to masrshal values file")
	}

	valuesName, err := filepath.Abs(filepath.Join(dir, "values"))
	if err != nil {
		return emperror.Wrap(err, "failed to construct values file name")
	}

	if err := ioutil.WriteFile(valuesName, values, 0600); err != nil {
		return emperror.Wrapf(err, "failed to write temporary file %q", valuesName)
	}

	if err := ioutil.WriteFile(filename, values, 0600); err != nil {
		return emperror.Wrapf(err, "failed to write values.yaml file %q", filename)
	}

	kubeconfigName, err := filepath.Abs(filepath.Join(dir, "kubeconfig"))
	if err != nil {
		return emperror.Wrap(err, "failed to construct kubeconfig file name")
	}

	if err := copyKubeconfig(banzaiCli, kubeconfigName, kindCluster); err != nil {
		return emperror.Wrap(err, "failed to copy Kubeconfig")
	}

	tfdir, err := filepath.Abs("./.tfstate")
	if err != nil {
		return emperror.Wrap(err, "failed to construct tfstate directory path")
	}

	log.Info("controlplane is being created")
	return emperror.Wrap(runInternal("apply", valuesName, kubeconfigName, tfdir, options.controlPlaneInstallerOptions), "controlplane creation failed")
}

func isInteractive() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd()) {
		return !viper.GetBool("formatting.no-interactive")
	}
	return viper.GetBool("formatting.force-interactive")
}

func runInternal(command, valuesFile, kubeconfigFile, tfdir string, installerOptions controlPlaneInstallerOptions) error {

	infoCmd := exec.Command("docker", "info", "-f", "{{or (eq .OperatingSystem \"Docker Desktop\") (eq .OperatingSystem \"Docker for Mac\")}}")

	infoOuput, err := infoCmd.Output()
	if err != nil {
		return err
	}

	isDockerForMac := strings.Trim(string(infoOuput), "\n")

	if installerOptions.pullInstaller {
		if err := installerOptions.pullDockerImage(); err != nil {
			return emperror.Wrap(err, "failed to pull cp-installer")
		}
	}

	args := []string{
		"run", "-it", "--rm",
		"-v", fmt.Sprintf("%s:/root/.kube/config", kubeconfigFile),
		"-v", fmt.Sprintf("%s:/tfstate", tfdir),
		"-e", fmt.Sprintf("IS_DOCKER_FOR_MAC=%s", isDockerForMac),
		"--entrypoint", "/terraform/entrypoint.sh",
	}

	if valuesFile != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/terraform/values.yaml", valuesFile))
	}

	args = append(args,
		fmt.Sprintf("banzaicloud/cp-installer:%s", installerOptions.installerTag),
		command,
		"-state=/tfstate/terraform.tfstate", // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271
		"-parallelism=1")

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		println("\nPipeline is ready, now you can login with: \x1b[1mbanzai login\x1b[0m")
	}

	return err
}
