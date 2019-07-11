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

package login

import (
	"fmt"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

const defaultLoginFlow = "login with browser"

type loginOptions struct {
	token    string
	endpoint string
	orgName  string
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
	flags.StringVarP(&options.endpoint, "endpoint", "e", "", "Pipeline API endpoint to save")
	flags.StringVarP(&options.orgName, "organization", "", "", "Name of the organization to select as default")

	return cmd
}

func runLogin(banzaiCli cli.Cli, options loginOptions) error {
	endpoint := viper.GetString("pipeline.basepath")

	if options.endpoint != "" {
		endpoint = options.endpoint
	} else if banzaiCli.Interactive() {
		_ = survey.AskOne(
			&survey.Input{
				Message: "Pipeline endpoint:",
				Help:    "The API endpoint to use for accessing Pipeline",
				Default: endpoint,
			},
			&endpoint, survey.Required)
	}

	if endpoint == "" {
		return errors.New("Please set Pipeline endpoint with --endpoint, or run the command interactively.")
	}

	token := options.token
	if token == "" {
		if banzaiCli.Interactive() {
			_ = survey.AskOne(
				&survey.Input{
					Message: "Pipeline token:",
					Default: defaultLoginFlow,
					Help:    fmt.Sprintf("Login through a browser flow or copy your Pipeline access token from the token field of %s/api/v1/token", endpoint),
				},
				&token, nil)
		} else {
			return errors.New("Please set Pipeline token with --token, or run the command interactively.")
		}
	}

	if token == "" || token == defaultLoginFlow {
		var err error
		token, err = runServer(banzaiCli, endpoint)
		if err != nil {
			return err
		}
	}

	viper.Set("pipeline.token", token)
	viper.Set("pipeline.basepath", endpoint)

	var orgID int32
	if options.orgName != "" {
		orgs, err := input.GetOrganizations(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "could not get organizations")
		}

		var orgFound bool
		orgID, orgFound = orgs[options.orgName]
		if !orgFound {
			return errors.Errorf("organization %q doesn't exist", options.orgName)
		}
	} else if banzaiCli.Interactive() {
		orgID = input.AskOrganization(banzaiCli)
	}

	banzaiCli.Context().SetOrganizationID(orgID)

	return nil
}
