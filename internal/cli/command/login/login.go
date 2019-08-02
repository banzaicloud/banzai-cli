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
	"context"
	"fmt"
	"os"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/dgrijalva/jwt-go"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

const defaultLoginFlow = "login with browser"

type loginOptions struct {
	token      string
	endpoint   string
	orgName    string
	permanent  bool
	skipVerify bool
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
	flags.BoolVarP(&options.permanent, "permanent", "", false, "Create permanent token (interactive login flow only)")
	flags.BoolVar(&options.skipVerify, "skip-verify", false, "Skip certificate verification and pin fingerprint")

	return cmd
}

func Login(banzaiCli cli.Cli, endpoint, orgName string, permanent bool, skipVerify bool) error {
	return runLogin(banzaiCli, loginOptions{endpoint: endpoint, orgName: orgName, permanent: permanent, skipVerify: skipVerify})
}

func runLogin(banzaiCli cli.Cli, options loginOptions) error {
	endpoint := viper.GetString("pipeline.basepath")

	if options.endpoint != "" {
		endpoint = options.endpoint
	} else if banzaiCli.Interactive() {
		err := survey.AskOne(
			&survey.Input{
				Message: "Pipeline endpoint:",
				Help:    "The API endpoint to use for accessing Pipeline",
				Default: endpoint,
			},
			&endpoint, survey.Required)
		if err != nil {
			return emperror.Wrap(err, "no endpoint selected")
		}
	}

	if endpoint == "" {
		return errors.New("Please set Pipeline endpoint with --endpoint, or run the command interactively.")
	}

	log.Debugf("checking if endpoint is available: %q", endpoint)
	fingerprint, x509Err, err := cli.CheckPipelineEndpoint(endpoint)
	if err != nil {
		return err
	}

	if x509Err != nil {
		if !options.skipVerify && banzaiCli.Interactive() {
			_ = survey.AskOne(
				&survey.Confirm{
					Message: fmt.Sprintf("Failed to verify server certificate: %v. Do you want to connect anyway?", x509Err),
					Help:    fmt.Sprintf("The following certificate fingerprint will be pinned: %v", fingerprint),
				},
				&options.skipVerify, nil)
		}
		if options.skipVerify {
			log.Warnf("Could not verify server certificate: %v. Pinning certificate fingerprint %s.", x509Err, fingerprint)
		} else {
			return emperror.Wrap(x509Err, "could not verify server certificate")
		}
	} else {
		fingerprint = "" // don't pin valid certificates. TODO: make this an option
	}

	token := options.token
	if token == "" {
		if banzaiCli.Interactive() {
			err := survey.AskOne(
				&survey.Input{
					Message: "Pipeline token:",
					Default: defaultLoginFlow,
					Help:    fmt.Sprintf("Login through a browser flow or copy your Pipeline access token from the token field of %s/api/v1/token", endpoint),
				},
				&token, nil)
			if err != nil {
				return emperror.Wrap(err, "no token selected")
			}
		} else {
			return errors.New("Please set Pipeline token with --token, or run the command interactively.")
		}
	}

	banzaiCli.Context().SetFingerprint(fingerprint)
	sessionToken := false
	if token == "" || token == defaultLoginFlow {
		var err error
		token, err = runServer(banzaiCli, endpoint)
		if err != nil {
			return err
		}

		sessionToken = true
	}

	viper.Set("pipeline.basepath", endpoint)
	banzaiCli.Context().SetToken(token)

	if !options.permanent && banzaiCli.Interactive() {
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Create permanent token?",
				Help:    "Create a permanent token instead of saving the temporary one generated automatically.",
			},
			&options.permanent, nil)
	}

	if options.permanent {
		err := createPermanentToken(banzaiCli)
		if err != nil {
			return emperror.Wrap(err, "failed to create permanent token")
		}

		if sessionToken {
			if err := deleteToken(banzaiCli, token); err != nil {
				return emperror.Wrap(err, "failed to delete session token")
			}
		}
	}

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

func createPermanentToken(banzaiCli cli.Cli) error {

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	req := pipeline.TokenCreateRequest{Name: fmt.Sprintf("cli-%s-%s", hostname, user)}
	permToken, _, err := banzaiCli.Client().AuthApi.CreateToken(context.Background(), req)
	if err != nil {
		return err
	}

	banzaiCli.Context().SetToken(permToken.Token)
	return nil
}

func deleteToken(banzaiCli cli.Cli, secret string) error {
	token, _ := jwt.Parse(secret, nil)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("can't parse old token")
	}

	id, ok := claims["jti"].(string)
	if !ok {
		return errors.New("can't find token id in secret")
	}

	_, err := banzaiCli.Client().AuthApi.DeleteToken(context.Background(), id)
	return err
}
