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
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type createOptions struct {
	init bool
	initOptions
}

// NewUpCommand creates a new cobra.Command for `banzai pipeline up`.
func NewUpCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"c"},
		Short:   "Deploy Banzai Cloud Pipeline",
		Long:    `Deploy or upgrade an instance of Banzai Cloud Pipeline based on a values file in the workspace, or initialize the workspace from an input file or an interactive session.` + initLongDescription,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUp(options, banzaiCli)
		},
	}

	options.initOptions = newInitOptions(cmd, banzaiCli)

	flags := cmd.Flags()
	flags.BoolVarP(&options.init, "init", "i", false, "Initialize workspace")

	return cmd
}

func runUp(options createOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	if !options.valuesExists() {
		if !options.init && banzaiCli.Interactive() {
			if err := survey.AskOne(
				&survey.Confirm{
					Message: "The workspace is not initialized. Do you want to initialize it now?",
					Default: true,
				},
				&options.init,
				nil,
			); err != nil {
				options.init = false
			}
		}
		if options.init {
			if err := runInit(options.initOptions, banzaiCli); err != nil {
				return err
			}
		} else {
			return errors.New("workspace is uninitialized")
		}
	} else {
		log.Debugf("using existing workspace %q", options.workspace)
	}

	var values map[string]interface{}
	if err := options.readValues(values); err != nil {
		return err
	}

	if !options.kubeconfigExists() {
		switch values["provider"] {
		case providerKind:
			err := ensureKINDCluster(banzaiCli, options.cpContext)
			if err != nil {
				return emperror.Wrap(err, "failed to create KIND cluster")
			}

		case providerEc2:
			err := ensureEC2Cluster(banzaiCli, options.cpContext)
			if err != nil {
				return emperror.Wrap(err, "failed to create EC2 cluster")
			}
		default:
			return errors.New("could not find Kubeconfig in workspace")
		}
	}

	log.Info("Deploying Banzai Cloud Pipeline to Kubernetes cluster...")
	return emperror.Wrap(runInternal("apply", options.cpContext, nil), "controlplane creation failed")
}

func runInternal(command string, options cpContext, env map[string]string) error {

	infoCmd := exec.Command("docker", "info", "-f", "{{or (eq .OperatingSystem \"Docker Desktop\") (eq .OperatingSystem \"Docker for Mac\")}}")

	infoOuput, err := infoCmd.Output()
	if err != nil {
		return err
	}

	isDockerForMac := strings.Trim(string(infoOuput), "\n")

	if options.pullInstaller {
		if err := options.pullDockerImage(); err != nil {
			return emperror.Wrap(err, "failed to pull cp-installer")
		}
	}

	args := []string{
		"run", "-it", "--rm",
		"-v", fmt.Sprintf("%s:/root", options.workspace),
		"-e", fmt.Sprintf("IS_DOCKER_FOR_MAC=%s", isDockerForMac),
		"--entrypoint", "/terraform/entrypoint.sh",
	}

	envs := os.Environ()
	for key, value := range env {
		args = append(args, "-e", key)
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args,
		fmt.Sprintf("banzaicloud/cp-installer:%s", options.installerTag),
		command,
		"-state=/tfstate/terraform.tfstate", // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271
		"-parallelism=1")

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		println("\nPipeline is ready, now you can login with: \x1b[1mbanzai login\x1b[0m")
		// TODO add --endpoint to example command
	}

	return err
}
