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
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	serviceutils "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type ActivateManager struct {
	baseManager
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func (ActivateManager) BuildRequestInteractively(banzaiCli cli.Cli, clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {

	defaultSpec := ServiceSpec{
		ExternalDNS: ExternalDNS{
			Provider: &Provider{
				Name: dnsBanzaiCloud,
			},
		},
	}

	builtSpec, err := assembleServiceRequest(banzaiCli, clusterCtx, defaultSpec, NewActionContext(actionNew))
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to build external DNS service request")
	}

	return pipeline.ActivateIntegratedServiceRequest{
		Spec: builtSpec,
	}, nil
}

func (ActivateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateIntegratedServiceRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
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

	defaultSecret := serviceutils.NameForID(secretsMap, selectedProvider.SecretID)
	if defaultSecret == "" {
		// if no secrets is set so far, the first secret is  the default
		defaultSecret = serviceutils.Names(secretsMap)[0]
	}

	secretIDQuestion := survey.Question{
		Name: "SecretID",
		Prompt: &survey.Select{
			Message: "Please select the secret to access the DNS provider",
			Options: serviceutils.Names(secretsMap),
			Default: defaultSecret,
		},
		Validate:  survey.Required,
		Transform: serviceutils.NameToIDTransformer(secretsMap),
	}

	switch selectedProvider.Name {
	case dnsBanzaiCloud:
		// no need for secrets - remove if set (update flow)
		providerWithSecret.SecretID = ""
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

	// helper struct for requesting provider specific integratedservice-options
	type providerOptions struct {
		Project       string `json:"project,omitempty" mapstructure:"project"`
		ResourceGroup string `json:"resourceGroup,omitempty" mapstructure:"resourceGroup"`
		Region        string `json:"region,omitempty" mapstructure:"region"`
		BatchSize     int    `json:"batchSize,omitempty" mapstructure:"batchSize"`
	}

	var currentProviderOpts providerOptions
	if err := mapstructure.Decode(selectedProvider.Options, &currentProviderOpts); err != nil {
		return Provider{}, errors.WrapIf(err, "failed to decode provider options")
	}

	providerWithOptions := selectedProvider
	// collects provider specific questions
	questions := make([]*survey.Question, 0)

	switch selectedProvider.Name {
	case dnsBanzaiCloud:
		// no need for secrets

	case dnsRoute53:
		regions, r, err := banzaiCLI.CloudinfoClient().RegionsApi.GetRegions(context.Background(), "amazon", "eks")
		if err := serviceutils.CheckCallResults(r, err); err != nil {
			return Provider{}, errors.Wrap(err, "failed to get regions")
		}

		regOptions := make(serviceutils.IdToNameMap, len(regions))
		for _, reg := range regions {
			regOptions[reg.Id] = reg.Name
		}
		questions = append(questions,
			&survey.Question{
				Name: "Region",
				Prompt: &survey.Select{
					Message: "Please select the Amazon region:",
					Options: serviceutils.Names(regOptions),
					Default: serviceutils.NameForID(regOptions, currentProviderOpts.Region),
				},
				Transform: serviceutils.NameToIDTransformer(regOptions),
			},
			&survey.Question{
				Name: "BatchSize",
				Prompt: &survey.Input{
					Message: "Please provide the batch size",
					Default: strconv.Itoa(currentProviderOpts.BatchSize),
				},
			},
		)
	case dnsGoogle:
		projectsMap, err := getGoogleProjectsMap(banzaiCLI, providerWithOptions)
		if err != nil {
			return Provider{}, errors.WrapIf(err, "failed to get google projects")
		}
		defaultProject := serviceutils.NameForID(projectsMap, selectedProvider.Options["project"].(string))
		if defaultProject == "" {
			// the default is the first project
			defaultProject = serviceutils.Names(projectsMap)[0]
		}

		questions = append(questions,
			&survey.Question{
				Name: "",
				Prompt: &survey.Select{
					Message: "Please select the google project",
					Options: serviceutils.Names(projectsMap),
					Default: defaultProject,
				},
				Validate:  survey.Required,
				Transform: serviceutils.NameToIDTransformer(projectsMap),
			},
		)

	case dnsAzure:
		resourceGroups, err := getAzureResourceGroupMap(banzaiCLI, providerWithOptions)
		if err != nil {
			return Provider{}, errors.WrapIf(err, "failed to get azure resourceGroups")
		}

		questions = append(questions,
			&survey.Question{
				Name: "ResourceGroup",
				Prompt: &survey.Select{
					Message: "Please select the Azure Resource Group",
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
func getGoogleProjectsMap(banzaiCLI cli.Cli, provider Provider) (serviceutils.IdToNameMap, error) {

	projects, _, err := banzaiCLI.Client().GoogleApi.ListProjects(
		context.Background(),
		banzaiCLI.Context().OrganizationID(),
		provider.SecretID) // it's assumed, that the secret id is already filled
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve google projects")
	}

	projectMap := make(serviceutils.IdToNameMap, 0)
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
	// the banzaicloud-dns is the default!
	defaultProviderName := providerMeta[dnsBanzaiCloud].Name

	if providerIn.Name != "" {
		pn, ok := providerMeta[providerIn.Name]
		if ok {
			defaultProviderName = pn.Name
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

func readExternalDNS(extDnsIn ExternalDNS, actionCtx actionContext) (ExternalDNS, error) {
	defaultDomainFilters := strings.Join(extDnsIn.DomainFilters, ",")
	providerQuestions := make([]*survey.Question, 0)

	if actionCtx.providerName != dnsBanzaiCloud {
		providerQuestions = append(providerQuestions, &survey.Question{
			Name: "DomainFilters",
			Prompt: &survey.Input{
				Message: "Please provide domain filters to match domains against:",
				Default: defaultDomainFilters,
				Help:    "Comma separated list of domains that are expected to trigger the external dns. Example: foo.com,bar.com",
			},
		})
	}

	questions := []*survey.Question{
		{
			Name: "Policy",
			Prompt: &survey.Select{
				Message: "Please select the policy for the provider:",
				Options: policies,
				Default: extDnsIn.Policy,
			},
		},
		{
			Name: "Sources",
			Prompt: &survey.MultiSelect{
				Message: "Please select resource types to monitor:",
				Options: sources,
				Default: extDnsIn.Sources,
			},
		},
		{
			Name: "TxtOwnerId",
			Prompt: &survey.Input{
				Message: "Please specify the TXT record owner id:",
				Default: extDnsIn.TxtOwnerId,
				Help:    "When using the TXT registry, a name that identifies this instance of ExternalDNS. Autogenerated value if left emty",
			},
		},
	}

	providerQuestions = append(providerQuestions, questions...)
	if err := survey.Ask(providerQuestions, &extDnsIn); err != nil {
		return ExternalDNS{}, errors.WrapIf(err, "request assembly failed")
	}

	return extDnsIn, nil
}

// getSecretsForProvider retrieves the available secrets for the provider as a map (secretID -> secretName)
func getSecretsForProvider(banzaiCLI cli.Cli, dnsProvider string) (serviceutils.IdToNameMap, error) {
	secretMap := make(serviceutils.IdToNameMap, 0)

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

// getServiceSpecDefaults fills the spec with provider specific defaults (activate flow only)
func getServiceSpecDefaults(banzaiCLI cli.Cli, clusterCtx clustercontext.Context, specIn ServiceSpec, actionCtx actionContext) (ServiceSpec, error) {
	switch actionCtx.providerName {
	case dnsBanzaiCloud:
		caps, r, err := banzaiCLI.Client().PipelineApi.ListCapabilities(context.Background())
		if err := serviceutils.CheckCallResults(r, err); err != nil {
			return ServiceSpec{}, errors.WrapIf(err, "failed to retrieve capabilities")
		}

		rawDnsCaps, ok := caps["features"]["dns"]
		if !ok {
			return ServiceSpec{}, errors.New("no DNS capabilities found")
		}

		var dnsCapability = struct {
			Enabled    bool   `json:"enabled" mapstructure:"enabled"`
			BaseDomain string `json:"baseDomain" mapstructure:"baseDomain"`
		}{}

		if err := mapstructure.Decode(rawDnsCaps, &dnsCapability); err != nil {
			return ServiceSpec{}, errors.WrapIf(err, "failed to parse DNS capabilities")
		}

		org, r, err := banzaiCLI.Client().OrganizationsApi.GetOrg(context.Background(), banzaiCLI.Context().OrganizationID())
		if err := serviceutils.CheckCallResults(r, err); err != nil {
			return ServiceSpec{}, errors.WrapIf(err, "failed to retrieves organizaton")
		}

		clusterDomain := fmt.Sprintf("%s.%s.%s", clusterCtx.ClusterName(), org.Name, dnsCapability.BaseDomain)

		if actionCtx.IsUpdate() {
			// cleanup in case of update from another provider
			retSpec := specIn
			retSpec.ExternalDNS.Provider = &Provider{
				Name: dnsBanzaiCloud,
			}
			retSpec.ClusterDomain = clusterDomain
			retSpec.ExternalDNS.DomainFilters = []string{clusterDomain}
			return retSpec, nil
		}

		// activate flow, plain new defaults
		return ServiceSpec{
			ExternalDNS: ExternalDNS{
				Policy:     policySync,
				Sources:    sources,
				TxtOwnerId: "",
				Provider: &Provider{
					Name: dnsBanzaiCloud,
				},
				// defaults to the clusterdomain
				DomainFilters: []string{clusterDomain},
			},
			ClusterDomain: clusterDomain,
		}, nil

	default:
		// fill the provider specifics if required
	}

	// update non banzai dns provider
	if actionCtx.IsUpdate() {
		return specIn, nil
	}

	// new non banzai dns provider
	return ServiceSpec{
		ExternalDNS: ExternalDNS{
			Policy:     policySync,
			Sources:    sources,
			TxtOwnerId: "",
		},
	}, nil
}
