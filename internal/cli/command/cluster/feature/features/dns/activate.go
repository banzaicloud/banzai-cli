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

package dns

import (
	"context"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}

	cmd := &cobra.Command{
		Use:     "activate",
		Aliases: []string{"add", "enable", "install", "on"},
		Short:   "Activate the DNS feature of a cluster",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runActivate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "activate DNS cluster feature for")

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type activateOptions struct {
	clustercontext.Context
	filePath string
}

func runActivate(banzaiCli cli.Cli, options activateOptions, _ []string) error {
	var req pipeline.ActivateClusterFeatureRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if err := buildActivateReqInteractively(banzaiCli, options, &req); err != nil {
			return errors.WrapIf(err, "failed to build activate request interactively")
		}
	} else {
		if err := readActivateReqFromFileOrStdin(options.filePath, &req); err != nil {
			return errors.WrapIf(err, "failed to read DNS cluster feature specification")
		}
	}

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()
	resp, err := banzaiCli.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError("activate DNS cluster feature", err, resp.Request)
		log.Fatalf("could not activate DNS cluster feature: %v", err)
	}

	return nil
}

func readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func buildActivateReqInteractively(banzaiCli cli.Cli, _ activateOptions, req *pipeline.ActivateClusterFeatureRequest) error {

	var comp string
	if err := survey.AskOne(&survey.Select{Message: "Please select a DNS component to activate:", Options: []string{"Auto DNS", "Custom DNS"}, Default: "Auto DNS"}, &comp, nil); err != nil {
		return errors.Wrap(err, "failure during survey")
	}

	switch comp {
	case "Auto DNS":
		req.Spec = obj{
			"autoDns": obj{
				"enabled": true,
			},
		}
	case "Custom DNS":
		var domainFilters []string
		for {
			var domainFilter string
			if err := survey.AskOne(&survey.Input{Message: "Please provide a domain filter to match domains against", Default: "*"}, &domainFilter, nil); err != nil {
				return errors.Wrap(err, "failure during survey")
			}
			domainFilters = append(domainFilters, domainFilter)

			var another bool
			if err := survey.AskOne(&survey.Confirm{Message: "Would you like to add another domain filter?"}, &another, nil); err != nil {
				return errors.Wrap(err, "failure during survey")
			}
			if !another {
				break
			}
		}

		var clusterDomain string
		if err := survey.AskOne(&survey.Input{Message: "Please specify the cluster's domain:"}, &clusterDomain, nil); err != nil {
			return errors.Wrap(err, "failure during survey")
		}

		var provider string
		{
			options := make([]string, 0, len(providers))
			for _, p := range providers {
				options = append(options, p.Name)
			}
			if err := survey.AskOne(&survey.Select{Message: "Please select a DNS provider:", Options: options}, &provider, nil); err != nil {
				return errors.Wrap(err, "failure during survey")
			}
			for id, p := range providers {
				if p.Name == provider {
					provider = id
					break
				}
			}
		}

		// TODO: provider specific options

		orgID := banzaiCli.Context().OrganizationID()
		secretType := providers[provider].SecretType
		secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(secretType)})
		if err != nil {
			return errors.Wrap(err, "failed to retrieve secrets")
		}

		// TODO: share secret selection/creation flow
		var secretID string
		{
			options := make([]string, 1+len(secrets))
			options[0] = "Create new secret"
			for i, s := range secrets {
				options[1+i] = s.Name
			}
			var secretName string
			if err := survey.AskOne(&survey.Select{Message: "Please select a secret for accessing the provider:", Options: options}, &secretName, nil); err != nil {
				return errors.Wrap(err, "failed to retrieve secrets")
			}
			if secretName == options[0] {
				// TODO: create new secret
			} else {
				for _, s := range secrets {
					if s.Name == secretName {
						secretID = s.Id
						break
					}
				}
			}
		}

		req.Spec = obj{
			"customDns": obj{
				"domainFilters": domainFilters,
				"clusterDomain": clusterDomain,
				"provider": obj{
					"name":     provider,
					"secretId": secretID,
				},
			},
		}
	}

	var edit bool
	if err := survey.AskOne(&survey.Confirm{Message: "Do you want to edit the cluster feature activation request in your text editor?"}, &edit); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if !edit {
		return nil
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}
	if err := survey.AskOne(
		&survey.Editor{Default: string(content), HideDefault: true, AppendDefault: true},
		&content,
		survey.WithValidator(validateActivateClusterFeatureRequest)); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if err := json.Unmarshal(content, req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func validateActivateClusterFeatureRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}
