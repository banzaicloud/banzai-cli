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
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

type destroyOptions struct {
	*cpContext
}

// NewDownCommand creates a new cobra.Command for `banzai controlplane down`.
func NewDownCommand(banzaiCli cli.Cli) *cobra.Command {
	options := destroyOptions{}

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy the controlplane",
		Long:  "Destroy a controlplane based on json stdin or interactive session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runDestroy(options, banzaiCli)
		},
	}

	options.cpContext = NewContext(cmd, banzaiCli)

	return cmd
}

func runDestroy(options destroyOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	var values map[string]interface{}
	if err := options.readValues(&values); err != nil {
		return err
	}

	if banzaiCli.Interactive() {
		var destroy bool
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to DESTROY the controlplane now?",
				Default: false,
			},
			&destroy,
		)

		if !destroy {
			return errors.New("controlplane destroy cancelled")
		}
	}

	awsAccessKeyID, env, err := getImageMetadata(options.cpContext, values, false)
	if err != nil {
		return err
	}

	// TODO: check if there are any clusters are created with the pipeline instance

	log.Info("controlplane is being destroyed")
	switch values["provider"] {
	case providerEc2:
		if valuesConfig, ok := values["providerConfig"]; ok {
			if valuesConfig, ok := valuesConfig.(map[string]interface{}); ok {
				if ak := valuesConfig["accessKey"]; ak != "" {
					if ak != awsAccessKeyID {
						return errors.Errorf("Current AWS access key %q differs from the one used earlier: %q", ak, awsAccessKeyID)
					}
				}
			}
		}

		err := deleteEC2Cluster(options.cpContext, env)
		if err != nil {
			return errors.WrapIf(err, "EC2 cluster destroy failed")
		}

		if err := options.deleteKubeconfig(); err != nil {
			return errors.WrapIf(err, "failed to remove Kubeconfig")
		}

	case providerKind:
		if err := runTerraform("destroy", options.cpContext, env); err != nil {
			return errors.WrapIf(err, "control plane destroy failed")
		}

		if err := deleteKINDCluster(banzaiCli); err != nil {
			return errors.WrapIf(err, "KIND cluster destroy failed")
		}

		if err := options.deleteKubeconfig(); err != nil {
			return errors.WrapIf(err, "failed to remove Kubeconfig")
		}

	case providerCustom:
		if pc, ok := values["providerConfig"]; ok {
			pc := cast.ToStringMap(pc)
			if _, ok := pc["accessKey"]; ok {
				if ak := pc["accessKey"]; ak != "" {
					if ak != awsAccessKeyID {
						return errors.Errorf("Current AWS access key %q differs from the one used earlier: %q", ak, awsAccessKeyID)
					}
				}
			}
		}

		if err := deleteCustomCluster(options.cpContext, env); err != nil {
			return errors.WrapIf(err, "Custom Kubernetes cluster destroy failed")
		}

	default:
		err := runTerraform("destroy", options.cpContext, env)
		if err != nil {
			return errors.WrapIf(err, "control plane destroy failed")
		}
	}

	if err := options.deleteTfstate(); err != nil {
		return errors.WrapIf(err, "failed to remove state file")
	}

	return nil
}
