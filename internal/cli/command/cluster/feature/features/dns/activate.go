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
	"net/http"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type ActivateManager struct {
	baseManager
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func (ActivateManager) BuildRequestInteractively(banzaiCLI cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {
	builtSpec, err := assembleFeatureRequest(banzaiCLI, nil)
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

// assembleFeatureRequest assembles the request for activate and update the ExternalDNS feature
func assembleFeatureRequest(banzaiCli cli.Cli, rawSpec interface{}) (map[string]interface{}, error) {

	currentDnsFeatureSpec := DNSFeatureSpec{
		ExternalDNS: ExternalDNS{
			Provider: &Provider{},
		},
	}
	if rawSpec != nil {
		// update feature case
		if err := mapstructure.Decode(rawSpec, &currentDnsFeatureSpec); err != nil {
			return nil, errors.WrapIf(err, "failed to decode feature DNSFeatureSpec")
		}
	}

	// select the provider
	providerInfo, err := selectProvider(*currentDnsFeatureSpec.ExternalDNS.Provider)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider data")
	}

	if rawSpec == nil {
		// activate flow
		currentDnsFeatureSpec, err = getFeatureSpecDefaults(banzaiCli, providerInfo);
		if err != nil {
			return nil, errors.WrapIf(err, "failed to get dns feature defaults")
		}
	}

	// read secret
	providerInfo, err = decorateProviderSecret(banzaiCli, providerInfo)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider secret")
	}

	// read options
	providerInfo, err = decorateProviderOptions(banzaiCli, providerInfo)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider options")
	}

	externalDNS := currentDnsFeatureSpec.ExternalDNS
	externalDNS.Provider = &providerInfo

	externalDNS, err = readExternalDNS(externalDNS)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read external dns data")
	}

	clusterDomain, err := readClusterDomain(currentDnsFeatureSpec.ClusterDomain)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read cluster domain")
	}

	var jsonSpec map[string]interface{}
	if err := mapstructure.Decode(
		&DNSFeatureSpec{
			ExternalDNS:   externalDNS,
			ClusterDomain: clusterDomain,
		},
		&jsonSpec);
		err != nil {
		return nil, errors.WrapIf(err, "failed to assemble DNSFeatureSpec")
	}

	return jsonSpec, nil
}

func readClusterDomain(currentDomain string) (string, error) {
	qs := []*survey.Question{
		{
			Name: "ClusterDomain",
			Prompt: &survey.Input{
				Message: "Please specify the cluster domain",
				Default: currentDomain,
				Help:    "cluster domain",
			},
		},
	}

	var clusterDomain string
	if err := survey.Ask(qs, &clusterDomain); err != nil {
		return "", errors.WrapIf(err, "failed to read cluster domain")
	}

	return clusterDomain, nil
}

// decorateProviderSecret decorates the selected provider with secret information
func decorateProviderSecret(banzaiCLI cli.Cli, selectedProvider Provider) (Provider, error) {
	providerWithSecret := selectedProvider // just for naming

	// collects provider specific questions
	questions := make([]*survey.Question, 0)

	secretsMap, err := getSecretsForProvider(banzaiCLI, selectedProvider.Name)
	if err != nil {
		return Provider{}, errors.WrapIf(err, "failed to retrieve secrets for provider")
	}

	defaultSecret := NameForID(secretsMap, selectedProvider.SecretID)
	if defaultSecret == "" {
		// if no secrets is set so far, the first secret is  the default
		defaultSecret = Names(secretsMap)[0]
	}

	secretIDQuestion := survey.Question{
		Name: "SecretID",
		Prompt: &survey.Select{
			Message: "Please select the secret to access the DNS provider",
			Options: Names(secretsMap),
			Default: defaultSecret,
		},
		Validate:  survey.Required,
		Transform: nameToIDTransformer(secretsMap),
	}

	switch selectedProvider.Name {
	case dnsBanzaiCloud:
		// no need for secrets
	case dnsRoute53:
		questions = append(questions, &secretIDQuestion)
	case dnsGoogle:
		questions = append(questions, &secretIDQuestion)
	case dnsAzure:
		questions = append(questions, &secretIDQuestion)
	}

	if err := survey.Ask(questions, &providerWithSecret); err != nil {
		return Provider{}, errors.WrapIf(err, "request assembly failed")
	}

	return providerWithSecret, nil
}

func decorateProviderOptions(banzaiCLI cli.Cli, selectedProvider Provider) (Provider, error) {
	type providerOptions struct {
		Project       string `json:"project,omitempty"`
		ResourceGroup string `json:"resourceGroup,omitempty"`
	}

	providerWithOptions := selectedProvider
	// collects provider specific questions
	questions := make([]*survey.Question, 0)

	switch selectedProvider.Name {
	case dnsBanzaiCloud:
		// no need for secrets

	case dnsRoute53:
	case dnsGoogle:
		projectsMap, err := getGoogleProjectsMap(banzaiCLI, providerWithOptions)
		if err != nil {
			return Provider{}, errors.WrapIf(err, "failed to get google projects")
		}
		defaultProject := NameForID(projectsMap, selectedProvider.Options["project"].(string))
		if defaultProject == "" {
			// the default is the first project
			defaultProject = Names(projectsMap)[0]
		}

		questions = append(questions,
			&survey.Question{
				Name: "",
				Prompt: &survey.Select{
					Message: "Please select the google project",
					Options: Names(projectsMap),
					Default: defaultProject,
				},
				Validate:  survey.Required,
				Transform: nameToIDTransformer(projectsMap),
			},
		)

	case dnsAzure:
		resourceGroups, err := getAzureResourceGroupMap(banzaiCLI, providerWithOptions)
		if err != nil {
			return Provider{}, errors.WrapIf(err, "failed to get azure resourceGroups")
		}

		questions = append(questions,
			&survey.Question{
				Name: "",
				Prompt: &survey.Select{
					Message: "Please select the google project",
					Options: resourceGroups,
					Default: selectedProvider.Options["resourcegroup"].(string),
				},
				Validate: survey.Required,
			},
		)
	default:
		// do nothing ?
	}

	options := providerOptions{}
	if err := survey.Ask(questions, &options); err != nil {
		return Provider{}, errors.WrapIf(err, "request assembly failed")
	}

	if err := mapstructure.Decode(&options, &providerWithOptions.Options); err != nil {
		return Provider{}, errors.WrapIf(err, "failed to assemble provider options")
	}

	return providerWithOptions, nil
}

//getGoogleProjectsMap retrieves google projects
func getGoogleProjectsMap(banzaiCLI cli.Cli, provider Provider) (idNameMap, error) {

	projects, _, err := banzaiCLI.Client().ProjectsApi.GetProjects(
		context.Background(),
		banzaiCLI.Context().OrganizationID(),
		provider.SecretID) // it's assumed, that the secret id is already filled
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve google projects")
	}

	projectMap := make(idNameMap, 0)
	for _, p := range projects.Projects {
		projectMap[p.ProjectId] = p.Name
	}

	return projectMap, nil
}

//getGoogleProjectsMap retrieves google projects
func getAzureResourceGroupMap(banzaiCLI cli.Cli, provider Provider) ([]string, error) {

	resourceGroups, _, err := banzaiCLI.Client().InfoApi.GetResourceGroups(
		context.Background(),
		banzaiCLI.Context().OrganizationID(),
		provider.SecretID)
	if err != nil {
		return nil, errors.WrapIf(utils.ConvertError(err), "can't list resource groups")
	}

	return resourceGroups, nil
}

func providerTransformer(ans interface{}) interface{} {
	selected := ans.(core.OptionAnswer).Value
	for provider, meta := range providerMeta {
		if meta.Name == selected {
			return core.OptionAnswer{
				Value: provider,
			}
		}
	}
	return core.OptionAnswer{}
}

func selectProvider(providerIn Provider) (Provider, error) {
	var defaultProviderName string
	if providerIn.Name != "" {
		defaultProviderName = providerMeta[providerIn.Name].Name
	}

	if defaultProviderName == "" {
		// select the first from the list
		for _, v := range providerMeta {
			defaultProviderName = v.Name
			break
		}

	}

	providerOptions := make([]string, 0, len(providerMeta))
	for _, p := range providerMeta {
		providerOptions = append(providerOptions, p.Name)
	}

	providerQuestions := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Select{
				Message: "Please select the DNS provider",
				Options: providerOptions,
				Default: defaultProviderName,
			},
			Validate:  survey.Required,
			Transform: providerTransformer,
		},
	}

	if err := survey.Ask(providerQuestions, &providerIn); err != nil {
		return Provider{}, errors.WrapIf(err, "request assembly failed")
	}

	return providerIn, nil
}

func readExternalDNS(extDnsIn ExternalDNS) (ExternalDNS, error) {

	providerQuestions := []*survey.Question{
		{
			Name: "DomainFilters",
			Prompt: &survey.Input{
				Message: "Please provide domain filters to match domains against",
				Default: strings.Join(extDnsIn.DomainFilters, ","),
				Help:    "To add multiple domains separate with commna (,) character. Example: foo.com,bar.com",
			},
		},
		{
			Name: "Policy",
			Prompt: &survey.Select{
				Message: "Please select the policy for the provider:",
				Options: []string{"sync", "upsert-only"},
				Default: extDnsIn.Policy,
			},
		},
		{
			Name: "Sources",
			Prompt: &survey.MultiSelect{
				Message: "Please select resource types to monitor:",
				Options: []string{"ingress", "service"},
				Default: extDnsIn.Sources,
			},
		},
		{
			Name: "TxtOwnerId",
			Prompt: &survey.Input{
				Message: "Please specify the TXT record owner id:",
				Default: extDnsIn.TxtOwnerId,
				Help:    "When using the TXT registry, a name that identifies this instance of ExternalDNS",
			},
		},
	}

	if err := survey.Ask(providerQuestions, &extDnsIn); err != nil {
		return ExternalDNS{}, errors.WrapIf(err, "request assembly failed")
	}

	return extDnsIn, nil
}

// getSecretsForProvider retrieves the available secrets for the provider as a map (secretID -> secretName)
func getSecretsForProvider(banzaiCLI cli.Cli, dnsProvider string) (idNameMap, error) {
	secretMap := make(idNameMap, 0)

	secrets, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(
		context.Background(),
		banzaiCLI.Context().OrganizationID(),
		&pipeline.GetSecretsOpts{Type_: optional.NewString(providerMeta[dnsProvider].SecretType)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secrets")
	}

	if len(secrets) == 0 {
		return nil, errors.Errorf("there are no available secrets for the provider %q", dnsProvider)
	}

	for _, s := range secrets {
		secretMap[s.Id] = s.Name
	}

	return secretMap, nil
}

// getFeatureSpecDefaults fills the spec with provider specific defaults (activate flow only)
func getFeatureSpecDefaults(banzaiCLI cli.Cli, provider Provider) (DNSFeatureSpec, error) {
	switch provider.Name {
	case dnsBanzaiCloud:
		caps, r, err := banzaiCLI.Client().PipelineApi.ListCapabilities(context.Background(), )
		if err != nil || r.StatusCode != http.StatusOK {
			return DNSFeatureSpec{}, errors.WrapIf(err, "failed to retrieve capabilities")
		}

		rawDnsCaps, ok := caps["features"]["dns"]
		if !ok {
			return DNSFeatureSpec{}, errors.New("no DNS capabilities found")
		}

		var dnsCapability = struct {
			Enabled    bool   `json:"enabled" mapstructure:"enabled"`
			BaseDomain string `json:"baseDomain" mapstructure:"baseDomain"`
		}{}

		if err := mapstructure.Decode(rawDnsCaps, &dnsCapability); err != nil {
			return DNSFeatureSpec{}, errors.WrapIf(err, "failed to parse DNS capabilities")
		}

		retSpec := DNSFeatureSpec{
			ExternalDNS: ExternalDNS{
				Policy:     "upsert-only",
				Sources:    []string{"ingress", "service"},
				TxtOwnerId: "",
				Provider: &Provider{
					Name: dnsBanzaiCloud,
				},
			},
		}

		if dnsCapability.Enabled {
			retSpec.ClusterDomain = dnsCapability.BaseDomain
		}
		return retSpec, nil

	default:
		// fill the provider specifics if required
	}

	return DNSFeatureSpec{
		ExternalDNS: ExternalDNS{
			Policy:     "upsert-only",
			Sources:    []string{"ingress", "service"},
			TxtOwnerId: "",
			Provider:   &provider,
		},
	}, nil
}
