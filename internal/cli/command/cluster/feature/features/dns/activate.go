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
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

type ActivateManager struct {
	baseManager
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func (ActivateManager) BuildRequestInteractively(banzaiCLI cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {
	builtSpec, err := buildExternalDNSFeatureRequest(banzaiCLI, defaults{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to build external DNS feature request")
	}

	return &pipeline.ActivateClusterFeatureRequest{
		Spec: builtSpec,
	}, nil
}

func (ActivateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}

func buildExternalDNSFeatureRequest(banzaiCli cli.Cli, defaults defaults) (map[string]interface{}, error) {
	provider, err := askDnsProvider(defaults)
	if err != nil {
		return nil, err
	}

	providerSpec, err := askDnsProviderSpecificOptions(banzaiCli, provider, defaults)
	if err != nil {
		return nil, err
	}

	domainFilters, err := askDomainFilter(defaults.domainFilters)
	if err != nil {
		return nil, err
	}

	policy, err := askForPolicy(defaults.policy)
	if err != nil {
		return nil, err
	}

	sources, err := askForSources(defaults.sources)
	if err != nil {
		return nil, err
	}

	txtOwner, err := askForTxtOwner(defaults.txtOwner)
	if err != nil {
		return nil, err
	}

	clusterDomain, err := askForClusterDomain(defaults.clusterDomain)
	if err != nil {
		return nil, err
	}

	return obj{
		"externalDns": obj{
			"provider":      providerSpec,
			"domainFilters": domainFilters,
			"policy":        policy,
			"sources":       sources,
			"txtOwnerId":    txtOwner,
		},
		"clusterDomain": clusterDomain,
	}, nil
}

func askDomainFilter(defaultValues []string) ([]string, error) {
	var domainFilter string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please provide a domain filter to match domains against",
			Default: strings.Join(defaultValues, ","),
			Help:    "To add multiple domains separate with commna (,) character. Example: foo.com, bar.com",
		},
		&domainFilter,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	filterItems := strings.Split(domainFilter, ",")
	filters := make([]string, 0, len(filterItems))
	for _, s := range filterItems {
		filters = append(filters, strings.TrimSpace(s))
	}

	return filters, nil
}

func askForClusterDomain(defaultClusterDomain string) (string, error) {
	var clusterDomain string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please specify the cluster's domain:",
			Default: defaultClusterDomain,
		},
		&clusterDomain,
	); err != nil {
		return "", errors.WrapIf(err, "failed to read cluster domain")
	}

	return clusterDomain, nil
}

func askDnsProvider(defaults defaults) (string, error) {
	options := make([]string, 0, len(providerMeta))
	for _, p := range providerMeta {
		options = append(options, p.Name)
	}

	var selectedProvider string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a DNS provider:",
			Options: options,
			Default: defaults.provider.name,
		},
		&selectedProvider,
	); err != nil {
		return "", errors.WrapIf(err, "faieled to select dns provider")
	}

	for id, p := range providerMeta {
		if p.Name == selectedProvider {
			return id, nil
		}
	}
	return "", errors.Errorf("unsupported provider %q", selectedProvider)
}

func askSecret(banzaiCli cli.Cli, provider string, defaultID string) (string, error) {

	orgID := banzaiCli.Context().OrganizationID()
	secretType := providerMeta[provider].SecretType
	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID,
		&pipeline.GetSecretsOpts{Type_: optional.NewString(secretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO (colin): add create secret option
	if len(secrets) == 0 {
		return "", errors.Errorf("there are no secrets with type %q", secretType)
	}

	var defaultName interface{}
	options := make([]string, len(secrets))
	for i, s := range secrets {
		options[i] = s.Name
		if s.Id == defaultID {
			defaultName = s.Name
		}
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a secret for accessing the provider:",
			Options: options,
			Default: defaultName,
		},
		&secretName,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve secrets")
	}

	for _, s := range secrets {
		if s.Name == secretName {
			return s.Id, nil
		}
	}

	return "", errors.Errorf("no secret with name %q", secretName)
}

func askDnsProviderSpecificOptions(banzaiCli cli.Cli, provider string, defaults defaults) (interface{}, error) {
	orgID := banzaiCli.Context().OrganizationID()

	var (
		secretID string
		options  providerOptions
		err      error
	)

	switch provider {
	case dnsBanzaiCloud:
	case dnsRoute53:
		secretID, err = askSecret(banzaiCli, provider, defaults.provider.secretId)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to get secret for %q provider", provider)
		}
	case dnsGoogle:
		secretID, err = askSecret(banzaiCli, provider, defaults.provider.secretId)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to get secret for %q provider", provider)
		}

		project, err := askGoogleProject(banzaiCli, secretID, orgID, defaults.provider.options["project"])
		if err != nil {
			return nil, err
		}

		options.Project = project

	case dnsAzure:
		secretID, err = askSecret(banzaiCli, provider, defaults.provider.secretId)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to get secret for %q provider", provider)
		}

		resourceGroup, err := input.AskResourceGroup(banzaiCli, orgID, secretID, defaults.provider.options["resourceGroup"])
		if err != nil {
			return nil, err
		}

		options.ResourceGroup = resourceGroup

	default:
		return nil, errors.Errorf("unsupported provider: %q", provider)
	}

	return activateCustomRequest{
		Name:     provider,
		SecretID: secretID,
		Options:  options,
	}, nil
}

func askGoogleProject(banzaiCli cli.Cli, secretID string, orgID int32, defaultID string) (string, error) {
	projects, _, err := banzaiCli.Client().ProjectsApi.GetProjects(context.Background(), orgID, secretID)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve google projects")
	}

	var defaultName interface{}
	options := make([]string, len(projects.Projects))
	for i, p := range projects.Projects {
		options[i] = p.Name
		if p.ProjectId == defaultID {
			defaultName = p.Name
		}
	}

	var projectName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a project:",
			Options: options,
			Default: defaultName,
		},
		&projectName,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve projects")
	}

	for _, p := range projects.Projects {
		if p.Name == projectName {
			return p.ProjectId, nil
		}
	}

	return "", errors.Errorf("unknown project name %q", projectName)
}

func askForPolicy(defaultPolicyValue string) (string, error) {
	var policy string

	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select the policy for the provider:",
			Options: []string{"sync", "upsert-only"},
			Default: defaultPolicyValue,
		},
		&policy,
	); err != nil {
		return "", errors.WrapIf(err, "failed to select policy")
	}

	return policy, nil
}

func askForSources(defaultSources []string) ([]string, error) {
	var sources []string

	if err := survey.AskOne(
		&survey.MultiSelect{
			Message: "Please select resource types to monitor:",
			Options: []string{"ingress", "service"},
			Default: defaultSources,
		},
		&sources,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select sources")
	}

	return sources, nil
}

func askForTxtOwner(defaultTxtOwner string) (string, error) {
	var txtOwner string

	if err := survey.AskOne(
		&survey.Input{
			Message: "Please specify the TXT owner id for the external dns instance:",
			Default: defaultTxtOwner,
			Help:    "When using the TXT registry, a name that identifies this instance of ExternalDNS",
		},
		&txtOwner,
	); err != nil {
		return "", errors.WrapIf(err, "failed to read in the txt owner")
	}

	return txtOwner, nil
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
