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

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	providerK8s  = "k8s"
	providerEc2  = "ec2"
	providerKind = "kind"
)

type initOptions struct {
	file     string
	provider string
	*cpContext
}

func newInitOptions(cmd *cobra.Command, banzaiCli cli.Cli) *initOptions {
	cp := NewContext(cmd, banzaiCli)
	options := initOptions{cpContext: cp}

	flags := cmd.Flags()
	flags.StringVarP(&options.file, "file", "f", "", "Input Banzai Cloud Pipeline instance descriptor file")
	flags.StringVar(&options.provider, "provider", "", "Provider of the infrastructure for the deployment (k8s|kind|ec2)")
	return &options
}

const initLongDescription = `

Depending on the --provider selection, the installer will work in the current Kubernetes context (k8s), deploy a KIND (Kubernetes in Docker) cluster to the local machine (kind), or deploy a PKE cluster in Amazon EC2 (ec2).

The directory specified with --workspace, set in the installer.workspace key of the config, or $BANZAI_INSTALLER_WORKSPACE (default: ~/.banzai/pipeline/default) will be used for storing the applied configuration and deployment status.

The command requires docker to be accessible in the system and able to run containers.

The input file will be copied to the workspace during initialization. Further changes can be done there before re-running the command (without --file).`

// NewInitCommand creates a new cobra.Command for `banzai pipeline init`.
func NewInitCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration for Banzai Cloud Pipeline",
		Long:  `Prepare a workspace for the deployment of an instance of Banzai Cloud Pipeline based on a values file or an interactive session.` + initLongDescription,
		Args:  cobra.NoArgs,
		PostRun: func(*cobra.Command, []string) {
			log.Info("Successfully initialized workspace. You can run now `banzai up` to deploy Pipeline.")
		},
	}

	options := newInitOptions(cmd, banzaiCli)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runInit(*options, banzaiCli)
	}

	return cmd
}

func runInit(options initOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	if options.valuesExists() {
		return errors.Errorf("workspace is already initialized in %q", options.workspace)
	}

	out := make(map[string]interface{})

	if !banzaiCli.Interactive() || options.file != "" {
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return emperror.Wrapf(err, "failed to read file %q", filename)
		}

		err = utils.Unmarshal(raw, &out)
		if err != nil {
			return emperror.Wrap(err, "failed to parse descriptor")
		}

		if provider, ok := out["provider"].(string); ok && provider != "" {
			options.provider = provider
		}
	}

	k8sContext, k8sConfig, err := input.GetCurrentKubecontext()
	if err != nil {
		k8sContext = ""
		log.Debugf("won't use local kubernetes context: %v", err)
	}

	if options.provider == "" {
		if !banzaiCli.Interactive() {
			return errors.New("please select provider (--provider or provider field of values file)")
		}

		choices := []string{"Create single-node cluster in Amazon EC2", "Create KIND (Kubernetes in Docker) cluster locally"}
		if k8sContext != "" {
			choices = append(choices, fmt.Sprintf("Use %q Kubernetes context", k8sContext))
		}

		var provider string
		if err := survey.AskOne(&survey.Select{Message: "Select provider:", Options: choices}, &provider, nil); err != nil {
			return err
		}

		switch {
		case provider == choices[0]:
			options.provider = providerEc2

		case provider == choices[1]:
			options.provider = providerKind

		case provider == choices[2]:
			options.provider = providerK8s

		}
	}

	out["provider"] = options.provider
	switch options.provider {
	case providerEc2:
		id, _, err := input.GetAmazonCredentials()
		if err != nil {
			log.Info("Please set your AWS credentials using aws-cli. See https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration")
			return emperror.Wrap(err, "failed to use local AWS credentials")
		}
		var confirmed bool
		_ = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Do you want to use the following AWS access key: %s?", id)}, &confirmed, nil)
		if !confirmed {
			return errors.New("cancelled")
		}
	case providerK8s:
		if err := options.writeKubeconfig([]byte(k8sConfig)); err != nil {
			return err
		}
	}

	return options.writeValues(out)
}
