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

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
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
				Default: true,
			},
			&destroy,
			nil,
		)

		if !destroy {
			return errors.New("controlplane destroy cancelled")
		}
	}

	// TODO: check if there are any clusters are created with the pipeline instance

	log.Info("controlplane is being destroyed")
	var env map[string]string
	switch values["provider"] {
	case providerEc2:
		id, creds, err := input.GetAmazonCredentials()
		if err != nil {
			return emperror.Wrap(err, "failed to get AWS credentials")
		}

		if valuesConfig, ok := values["providerConfig"]; ok {
			if valuesConfig, ok := valuesConfig.(map[string]interface{}); ok {
				if ak := valuesConfig["accessKey"]; ak != "" {
					if ak != id {
						return errors.Errorf("Current AWS access key %q differs from the one used earlier: %q", ak, id)
					}
				}
			}
		}
		env = creds

		err = deleteEC2Cluster(banzaiCli, *options.cpContext, env)

		if err != nil {
			return emperror.Wrap(err, "EC2 cluster destroy failed")
		}

		return nil

	case providerKind:
		err := deleteKINDCluster(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "KIND cluster destroy failed")
		}
	default:
		err := runInternal("destroy", *options.cpContext, env)
		if err != nil {
			return emperror.Wrap(err, "control plane destroy failed")
		}
		return nil
	}

	err := os.Remove(fmt.Sprintf("%s/terraform.tfstate", options.workspace))

	if err != nil {
		return emperror.Wrap(err, "remove state file failed")
	}

	return nil
}
