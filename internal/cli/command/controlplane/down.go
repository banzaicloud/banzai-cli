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
	"errors"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
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

	var env map[string]string
	switch values["provider"] {
	case providerEc2:
		_, creds, err := input.GetAmazonCredentials()
		if err != nil {
			return emperror.Wrap(err, "failed to get AWS credentials")
		}
		env = creds
	}

	log.Info("controlplane is being destroyed")
	err := runInternal("destroy", *options.cpContext, env)
	if err != nil {
		return emperror.Wrap(err, "control plane destroy failed")
	}

	switch values["provider"] {
	case providerKind:
		err = deleteKINDCluster(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "KIND cluster destroy failed")
		}
	}

	return nil
}
