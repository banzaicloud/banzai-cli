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
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			&endpoint, survey.WithValidator(survey.Required))
		if err != nil {
			return errors.WrapIf(err, "no endpoint selected")
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
				&options.skipVerify)
		}
		if options.skipVerify {
			log.Warnf("Could not verify server certificate: %v. Pinning certificate fingerprint %s.", x509Err, fingerprint)
		} else {
			return errors.WrapIf(x509Err, "could not verify server certificate")
		}
	} else {
		fingerprint = "" // don't pin valid certificates. TODO: make this an option
	}

	token := options.token
	if token == "" {
		if banzaiCli.Interactive() {
			var browserLogin bool
			err := survey.AskOne(
				&survey.Confirm{
					Message: "Login using web browser?",
					Help:    "The easiest login flow will open a browser window where you can log in using your credentials. Alternatively you can log in by a token.",
					Default: true,
				},
				&browserLogin)

			if err != nil {
				return err
			}

			if !browserLogin {
				err := survey.AskOne(
					&survey.Input{
						Message: "Pipeline token:",
						Default: defaultLoginFlow,
						Help:    "Create a token on the Settings page of the web UI (https://banzaicloud.com/docs/pipeline/security/authentication/personal-access-tokens/).",
					},
					&token, survey.WithValidator(survey.Required))
				if err != nil {
					return errors.WrapIf(err, "no token selected")
				}
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

	expiringToken, err := isExpiringToken(token)
	if err != nil {
		return errors.WrapIf(err, "failed to create permanent token")
	}

	if !options.permanent && banzaiCli.Interactive() && expiringToken {
		_ = survey.AskOne(
			&survey.Confirm{
				Message: "Create permanent token?",
				Help:    "Create a permanent token instead of saving the temporary one generated automatically.",
			},
			&options.permanent)
	}

	if options.permanent && expiringToken {
		err := createPermanentToken(banzaiCli)
		if err != nil {
			return errors.WrapIf(err, "failed to create permanent token")
		}

		if sessionToken {
			if err := deleteToken(banzaiCli, token); err != nil {
				return errors.WrapIf(err, "failed to delete session token")
			}
		}
	}

	var orgID int32
	if options.orgName != "" {
		orgs, err := input.GetOrganizations(banzaiCli)
		if err != nil {
			return errors.WrapIf(err, "could not get organizations")
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
	claims := jwt.StandardClaims{}
	_, _ = jwt.ParseWithClaims(secret, &claims, nil)

	if err := tokenNotExpired(claims); err != nil {
		return errors.Wrap(err, "old token is invalid")
	}

	_, err := banzaiCli.Client().AuthApi.DeleteToken(context.Background(), claims.Id)
	return err
}

func isExpiringToken(secret string) (bool, error) {
	claims := jwt.StandardClaims{}
	_, _ = jwt.ParseWithClaims(secret, &claims, nil)

	if err := tokenNotExpired(claims); err != nil {
		return false, errors.Wrap(err, "old token is invalid")
	}

	return claims.ExpiresAt != 0, nil
}

func tokenNotExpired(c jwt.StandardClaims) error {
	now := time.Now().Unix()

	if c.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		return fmt.Errorf("token is expired by %v", delta)
	}

	return nil
}
