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
	"fmt"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}

	cmd := &cobra.Command{
		Use:           "activate",
		Aliases:       []string{"add", "enable", "install", "on"},
		Short:         "Activate the DNS feature of a cluster",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	_, err := banzaiCli.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError("activate DNS cluster feature", err, req)
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
	comp, err := askDnsComponent()
	if err != nil {
		return errors.WrapIf(err, "error during choosing DNS component")
	}

	switch comp {
	case dnsAuto:
		buildAutoDNSFeatureRequest(req)
	case dnsCustom:
		if err := buildCustomDNSFeatureRequest(banzaiCli, req); err != nil {
			return err
		}
	}

	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Do you want to edit the cluster feature activation request in your text editor?",
		},
		&edit,
		nil,
	); err != nil {
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
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true,
		},
		&content,
		survey.WithValidator(validateActivateClusterFeatureRequest),
	); err != nil {
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

// todo activate??
func buildCustomDNSFeatureRequest(banzaiCli cli.Cli, req *pipeline.ActivateClusterFeatureRequest) error {
	domainFilters, err := askDomainFilter()
	if err != nil {
		return err
	}

	clusterDomain, err := askDomain()
	if err != nil {
		return err
	}

	provider, err := askDnsProvider()
	if err != nil {
		return err
	}

	secretID, err := askSecret(banzaiCli, provider)
	if err != nil {
		return err
	}

	providerSpec, err := askDnsProviderSpecificOptions(banzaiCli, provider, secretID)
	if err != nil {
		return err
	}

	req.Spec = obj{
		"customDns": obj{
			"enabled":       true,
			"domainFilters": domainFilters,
			"clusterDomain": clusterDomain,
			"provider":      providerSpec,
		},
	}

	return nil
}

func askDomainFilter() ([]string, error) {
	var domainFilters []string
	for {
		var domainFilter string
		if err := survey.AskOne(
			&survey.Input{
				Message: "Please provide a domain filter to match domains against",
				Default: "*",
			},
			&domainFilter,
			nil,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}
		domainFilters = append(domainFilters, domainFilter)

		var another bool
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Would you like to add another domain filter?",
			},
			&another,
			nil,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}
		if !another {
			break
		}
	}

	return domainFilters, nil
}

func askDomain() (string, error) {
	var clusterDomain string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please specify the cluster's domain:",
		},
		&clusterDomain,
		nil,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}

	return clusterDomain, nil
}

func askDnsProvider() (string, error) {
	var provider string
	{
		options := make([]string, 0, len(providers))
		for _, p := range providers {
			options = append(options, p.Name)
		}
		if err := survey.AskOne(
			&survey.Select{
				Message: "Please select a DNS provider:",
				Options: options,
			},
			&provider,
			nil,
		); err != nil {
			return "", errors.WrapIf(err, "failure during survey")
		}
		for id, p := range providers {
			if p.Name == provider {
				provider = id
				break
			}
		}
	}
	return provider, nil
}

func askSecret(banzaiCli cli.Cli, provider string) (string, error) {

	log.Debugf("load %s secrets", provider)

	orgID := banzaiCli.Context().OrganizationID()
	secretType := providers[provider].SecretType
	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(secretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO (colin): add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there's no secrets with '%s' type", secretType))
	}

	var secretID string
	options := make([]string, len(secrets))
	for i, s := range secrets {
		options[i] = s.Name
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a secret for accessing the provider:",
			Options: options,
		},
		&secretName,
		nil,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve secrets")
	}

	for _, s := range secrets {
		if s.Name == secretName {
			secretID = s.Id
			break
		}
	}

	return secretID, nil
}

func askDnsProviderSpecificOptions(banzaiCli cli.Cli, provider string, secretID string) (interface{}, error) {
	orgID := banzaiCli.Context().OrganizationID()

	r := activateCustomRequest{
		Name:     provider,
		SecretID: secretID,
	}

	switch provider {
	case dnsRoute53:
	case dnsGoogle:
		project, err := askGoogleProject(banzaiCli, secretID, orgID)
		if err != nil {
			return nil, err
		}
		r.Options = providerOptions{
			Project: project,
		}
	case dnsAzure:
		resourceGroup, err := input.AskResourceGroup(banzaiCli, orgID, secretID, "")
		if err != nil {
			return nil, err
		}
		r.Options = providerOptions{
			ResourceGroup: resourceGroup,
		}
	default:
		return nil, &NotSupportedProviderError{
			provider: provider,
		}
	}

	return r, nil
}

type NotSupportedProviderError struct {
	provider string
}

func (e *NotSupportedProviderError) Error() string {
	return fmt.Sprintf("not supported provider: %s", e.provider)
}

func askGoogleProject(banzaiCli cli.Cli, secretID string, orgID int32) (string, error) {
	projects, _, err := banzaiCli.Client().ProjectsApi.GetProjects(context.Background(), orgID, secretID)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve google projects")
	}

	options := make([]string, len(projects.Projects))
	for i, p := range projects.Projects {
		options[i] = p.Name
	}

	var projectName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a project:",
			Options: options,
		},
		&projectName,
		nil,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve projects")
	}

	var projectID string
	for _, p := range projects.Projects {
		if p.Name == projectName {
			projectID = p.ProjectId
			break
		}
	}

	return projectID, nil
}

func buildAutoDNSFeatureRequest(req *pipeline.ActivateClusterFeatureRequest) {
	req.Spec = obj{
		"autoDns": obj{
			"enabled": true,
		},
	}
}

func askDnsComponent() (string, error) {
	var comp string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a DNS component to activate:",
			Options: []string{dnsAuto, dnsCustom},
			Default: dnsAuto,
		},
		&comp,
		nil,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	return comp, nil
}

type activateCustomRequest struct {
	Name     string          `json:"name"`
	SecretID string          `json:"secretId"`
	Options  providerOptions `json:"options,omitempty"`
}

type providerOptions struct {
	Project       string `json:"project,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
}
