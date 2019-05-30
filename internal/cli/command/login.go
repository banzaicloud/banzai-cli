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

package command

import (
	"errors"
	"fmt"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

type loginOptions struct {
	token string
}

// NewLoginCommand returns a cobra command for logging in.
func NewLoginCommand(banzaiCli cli.Cli) *cobra.Command {
	options := loginOptions{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Configure and log in to a Banzai Cloud context",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runLogin(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.token, "token", "t", "", "Pipeline token to save")

	return cmd
}

func runLogin(banzaiCli cli.Cli, options loginOptions) error {
	endpoint := viper.GetString("pipeline.basepath")
	token := options.token

	if banzaiCli.Interactive() {
		_ = survey.AskOne(
			&survey.Input{
				Message: "Pipeline endpoint:",
				Help:    "The API endpoint to use for accessing Pipeline",
				Default: endpoint,
			},
			&endpoint, survey.Required)

		if token == "" {
			_ = survey.AskOne(
				&survey.Input{
					Message: "Pipeline token:",
					Help:    fmt.Sprintf("Please copy your Pipeline access token from the token field of %s/api/v1/token", endpoint),
				},
				&token, survey.Required)
		}
	}

	if token != "" {
		viper.Set("pipeline.token", token)
		viper.Set("pipeline.basepath", endpoint)

		var orgID int32

		if banzaiCli.Interactive() {
			orgID = input.AskOrganization(banzaiCli)
		}

		banzaiCli.Context().SetOrganizationID(orgID)
	} else {
		// nolint: stylecheck
		return errors.New("Password login is not implemented yet. Please either set a pipeline token aquired from https://beta.banzaicloud.io/pipeline/api/v1/token in the environment variable PIPELINE_TOKEN or as pipeline.token in ~/.banzai/config.yaml. You can also use the `banzai login -t $TOKEN` command.")
	}

	return nil
}
