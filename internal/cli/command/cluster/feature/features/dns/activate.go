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
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
)

type ActivateManager struct{}

func (m *ActivateManager) GetName() string {
	return featureName
}

func (m *ActivateManager) BuildRequestInteractively(banzaiCLI cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {
	var request pipeline.ActivateClusterFeatureRequest

	comp, err := askDnsComponent(dnsAuto)
	if err != nil {
		return nil, errors.WrapIf(err, "error during choosing DNS component")
	}

	switch comp {
	case dnsAuto:
		request.Spec = buildAutoDNSFeatureRequest()
	case dnsCustom:
		customDNS, err := buildCustomDNSFeatureRequest(banzaiCLI, defaults{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to build custom DNS feature request")
		}
		request.Spec = customDNS
	}

	return &request, nil
}

func (m *ActivateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func buildCustomDNSFeatureRequest(banzaiCli cli.Cli, defaults defaults) (map[string]interface{}, error) {
	domainFilters, err := askDomainFilter(defaults.domainFilters)
	if err != nil {
		return nil, err
	}

	clusterDomain, err := askDomain(defaults.clusterDomain)
	if err != nil {
		return nil, err
	}

	provider, err := askDnsProvider(defaults)
	if err != nil {
		return nil, err
	}

	providerSpec, err := askDnsProviderSpecificOptions(banzaiCli, provider, defaults)
	if err != nil {
		return nil, err
	}

	return obj{
		"customDns": obj{
			"enabled":       true,
			"domainFilters": domainFilters,
			"clusterDomain": clusterDomain,
			"provider":      providerSpec,
		},
	}, nil
}

func askDomainFilter(defaultValues []string) ([]string, error) {
	var domainFilter string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please provide a domain filter to match domains against",
			Default: strings.Join(defaultValues, ","),
			Help:    "To add multiple domains separate with commna (,) character. Like: foo.com, bar.com",
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

func askDomain(defaultValue string) (string, error) {
	var clusterDomain string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please specify the cluster's domain:",
			Default: defaultValue,
		},
		&clusterDomain,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}

	return clusterDomain, nil
}

func askDnsProvider(defaults defaults) (string, error) {
	options := make([]string, 0, len(providers))
	for _, p := range providers {
		options = append(options, p.Name)
	}

	var defaultProvider struct {
		Name       string
		SecretType string
	}
	if len(defaults.provider.name) != 0 {
		defaultProvider = providers[defaults.provider.name]
	}

	var provider string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a DNS provider:",
			Options: options,
			Default: defaultProvider.Name,
		},
		&provider,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	for id, p := range providers {
		if p.Name == provider {
			return id, nil
		}
	}
	return "", errors.Errorf("unsupported provider %q", provider)
}

func askSecret(banzaiCli cli.Cli, provider string, defaultID string) (string, error) {

	log.Debugf("load %s secrets", provider)

	orgID := banzaiCli.Context().OrganizationID()
	secretType := providers[provider].SecretType
	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(secretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO (colin): add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there are no secrets with type %q", secretType))
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

	secretID, err := askSecret(banzaiCli, provider, defaults.provider.secretId)
	if err != nil {
		return nil, errors.WrapIf(err, fmt.Sprintf("failed to get secret for %q provider", provider))
	}

	r := activateCustomRequest{
		Name:     provider,
		SecretID: secretID,
	}

	switch provider {
	case dnsRoute53:
	case dnsGoogle:
		project, err := askGoogleProject(banzaiCli, secretID, orgID, defaults.provider.options["project"])
		if err != nil {
			return nil, err
		}
		r.Options = providerOptions{
			Project: project,
		}
	case dnsAzure:
		resourceGroup, err := input.AskResourceGroup(banzaiCli, orgID, secretID, defaults.provider.options["resourceGroup"])
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

func buildAutoDNSFeatureRequest() map[string]interface{} {
	return obj{
		"autoDns": obj{
			"enabled": true,
		},
	}
}

func askDnsComponent(defaultValue string) (string, error) {
	var comp string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a DNS component to activate:",
			Options: []string{dnsAuto, dnsCustom},
			Default: defaultValue,
		},
		&comp,
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
