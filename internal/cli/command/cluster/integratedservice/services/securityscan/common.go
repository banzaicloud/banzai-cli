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

package securityscan

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
)

const (
	serviceName = "securityscan"
)

var (
	policyBundles = utils.IdToNameMap{
		"2c53a13c-1765-11e8-82ef-23527761d060": "Default bundle",
		"a81d4e45-6021-4b42-a217-a6554015d431": "DenyAll",
		"0cd4785e-71fa-4273-8ea5-3b15f515cca4": "RejectHigh",
		"bdb91dcc-62ca-49a2-a497-ee8a3bb7ec9f": "RejectCritical",
		"377c130d-0af7-45d4-adf9-cd72878993e2": "BlockRoot",
		"97b33e2c-3b57-4a3f-a12b-a8c0daa472a0": "AllowAll",
	}
)

//ServiceSpec security scan cluster integratedservice specific specification
type ServiceSpec struct {
	CustomAnchore    anchoreSpec       `json:"customAnchore" mapstructure:"customAnchore"`
	Policy           policySpec        `json:"policy" mapstructure:"policy"`
	ReleaseWhiteList []releaseSpec     `json:"releaseWhiteList,omitempty" mapstructure:"releaseWhiteList"`
	WebhookConfig    webHookConfigSpec `json:"webhookConfig" mapstructure:"webhookConfig"`
}

type baseManager struct{}

func (baseManager) GetName() string {
	return serviceName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

// Validate validates the input security scan specification.
func (s ServiceSpec) Validate() error {

	var validationErrors error

	if s.CustomAnchore.Enabled {
		validationErrors = s.CustomAnchore.Validate()
	}

	if !s.Policy.CustomPolicy.Enabled && s.Policy.PolicyID == "" {
		validationErrors = errors.Combine(validationErrors, errors.New("policyId is required"))
	}

	for _, releaseSpec := range s.ReleaseWhiteList {
		validationErrors = errors.Combine(validationErrors, releaseSpec.Validate())
	}

	validationErrors = errors.Combine(validationErrors, s.WebhookConfig.Validate())

	return validationErrors
}

type anchoreSpec struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Url      string `json:"url" mapstructure:"url"`
	SecretID string `json:"secretId" mapstructure:"secretId"`
}

func (a anchoreSpec) Validate() error {

	if a.Enabled {
		if a.Url == "" && a.SecretID == "" {
			return errors.New("both anchore url and secretId are required")
		}
	}

	return nil
}

type policySpec struct {
	PolicyID     string           `json:"policyId,omitempty" mapstructure:"policyId"`
	CustomPolicy customPolicySpec `json:"customPolicy,omitempty" mapstructure:"customPolicy"`
}

type customPolicySpec struct {
	Enabled bool                   `json:"enabled" mapstructure:"enabled"`
	Policy  map[string]interface{} `json:"policy" mapstructure:"policy"`
}

type releaseSpec struct {
	Name   string `json:"name" mapstructure:"name"`
	Reason string `json:"reason" mapstructure:"reason"`
	Regexp string `json:"regexp,omitempty" mapstructure:"regexp"`
}

func (r releaseSpec) Validate() error {
	if r.Name == "" || r.Reason == "" {
		return errors.NewPlain("both name and reason must be specified")
	}

	return nil
}

type webHookConfigSpec struct {
	Enabled    bool     `json:"enabled" mapstructure:"enabled"`
	Selector   string   `json:"selector" mapstructure:"selector"`
	Namespaces []string `json:"namespaces" mapstructure:"namespaces"`
}

func (w webHookConfigSpec) Validate() error {
	if w.Enabled {
		if w.Selector == "" || len(w.Namespaces) < 1 {
			return errors.NewPlain("selector and namespaces must be filled")
		}
	}

	return nil
}

// specAssembler component for common spec assembling operations
// designed mainly to handle activation and update spec assembly
type specAssembler struct {
	banzaiCLI cli.Cli
}

func (sa specAssembler) isServiceEnabled(ctx context.Context) error {
	capabilities, r, err := sa.banzaiCLI.Client().PipelineApi.ListCapabilities(ctx)
	if err := utils.CheckCallResults(r, err); err != nil {
		return errors.WrapIf(err, "failed to retrieve capabilities")
	}

	rawSecurityscanCapability, ok := capabilities["features"]["securityScan"]
	if !ok {
		return errors.New("no securityscan capabilities found")
	}

	var securityScanCapability = struct {
		Enabled bool `json:"enabled" mapstructure:"enabled"`
		Managed bool `json:"managed" mapstructure:"managed"`
	}{}

	if err := mapstructure.Decode(rawSecurityscanCapability, &securityScanCapability); err != nil {
		return errors.WrapIf(err, "failed to parse securityscan capabilities")
	}

	// todo change this implementation when adding support for non-managed anchore
	if !securityScanCapability.Enabled || !securityScanCapability.Managed {
		return errors.New("security scan is not enabled")
	}

	return nil
}

func (sa specAssembler) askForAnchoreConfig(currentAnchoreSpec *anchoreSpec) (*anchoreSpec, error) {

	if currentAnchoreSpec == nil {
		currentAnchoreSpec = &anchoreSpec{}
	}

	var customAnchore bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Configure a custom anchore instance? ",
		},
		&customAnchore,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !customAnchore {
		return &anchoreSpec{
			Enabled: false,
		}, nil
	}

	// custom anchore config
	var anchoreURL string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the custom anchore URL:",
			Default: currentAnchoreSpec.Url,
		},
		&anchoreURL,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read custom Anchore URL")
	}

	secretID, err := sa.askForSecret(currentAnchoreSpec.SecretID)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read secret for accessing custom Anchore ")
	}

	return &anchoreSpec{
		Enabled:  true,
		Url:      anchoreURL,
		SecretID: secretID,
	}, nil
}

func (sa specAssembler) askForSecret(currentSecretID string) (string, error) {
	const (
		PasswordSecretType = "password"
	)

	orgID := sa.banzaiCLI.Context().OrganizationID()

	secrets, _, err := sa.banzaiCLI.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(PasswordSecretType)})
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve secrets")
	}

	// TODO add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there are no secrets with type %q", PasswordSecretType))
	}

	options := make([]string, len(secrets))
	currentSecretName := secrets[0].Name // default!
	for i, s := range secrets {
		options[i] = s.Name
		if s.Id == currentSecretID {
			currentSecretName = s.Name
		}
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a secret to access the custom Anchore instance:",
			Options: options,
			Default: currentSecretName,
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

func (sa specAssembler) getNamespaces(ctx context.Context, orgID int32, clusterID int32) ([]string, error) {
	nsResponse, response, err := sa.banzaiCLI.Client().ClustersApi.ListNamespaces(ctx, orgID, clusterID)
	if err := utils.CheckCallResults(response, err); err != nil {
		return nil, errors.WrapIf(err, "failed to retrieve policies")
	}

	// filter out system namespaces
	filtered := make([]string, 0, len(nsResponse.Namespaces))
	for _, ns := range nsResponse.Namespaces {
		if ns.Name == "kube-system" || ns.Name == "pipeline-system" {
			continue
		}

		filtered = append(filtered, ns.Name)
	}

	return filtered, nil
}

// policies are statically stored, the selection is made from a "wired" list
func (sa *specAssembler) askForPolicy(policySpecIn policySpec) (policySpec, error) {

	defaultPolicyBundle := utils.NameForID(policyBundles, policySpecIn.PolicyID)
	if defaultPolicyBundle == "" {
		defaultPolicyBundle = "Default bundle"
	}

	qs := []*survey.Question{
		{
			Name: "PolicyID",
			Prompt: &survey.Select{
				Message: "please select the policy bundle",
				Options: utils.Names(policyBundles),
				Default: defaultPolicyBundle,
			},
			Transform: utils.NameToIDTransformer(policyBundles),
		},
	}

	if err := survey.Ask(qs, &policySpecIn); err != nil {
		return policySpec{}, errors.WrapIf(err, "failed to read cluster domain")
	}

	return policySpecIn, nil
}

func (sa specAssembler) askForWebHookConfig(ctx context.Context, orgID int32, clusterID int32, webhookSpecIn webHookConfigSpec) (webHookConfigSpec, error) {

	qs := []*survey.Question{
		{
			Name: "Enabled",
			Prompt: &survey.Confirm{
				Message: "enable the security scan webhook",
				Default: webhookSpecIn.Enabled,
			},
		},
	}

	if err := survey.Ask(qs, &webhookSpecIn); err != nil {
		return webHookConfigSpec{}, errors.WrapIf(err, "failed to read webhook configuration value")
	}

	if !webhookSpecIn.Enabled {
		return webHookConfigSpec{
			Enabled: false,
		}, nil
	}

	// set the default selector
	if webhookSpecIn.Selector == "" {
		webhookSpecIn.Selector = "include"
	}

	// ignore the error
	namespaceOptions, _ := sa.getNamespaces(ctx, orgID, clusterID)
	if len(namespaceOptions) == 0 {
		// couldn't retrieve namespaces / namespaces are not available
		namespaceOptions = append(namespaceOptions, "default")
	}

	// append the allStar
	namespaceOptions = append([]string{"*"}, namespaceOptions...)

	defaultNamespaces := webhookSpecIn.Namespaces
	// empty the namespaces field
	webhookSpecIn.Namespaces = make([]string, 0, len(defaultNamespaces))

	if len(defaultNamespaces) == 0 {
		defaultNamespaces = []string{"*"}
	}

	// questions to fill the remaining parts of the webhook configuration
	qs = []*survey.Question{
		{
			Name: "Selector",
			Prompt: &survey.Select{
				Message: "choose the selector for namespaces:",
				Options: []string{"include", "exclude"},
				Default: webhookSpecIn.Selector,
				Help:    "The selector defines whether the selected namespaces are included or excluded from security scans",
			},
		},
		{
			Name: "Namespaces",
			Prompt: &survey.MultiSelect{
				Message: "select the namespaces the selector applies to:",
				Options: namespaceOptions,
				Default: defaultNamespaces,
				Help:    "selected namespaces will be included or excluded form security scans",
			},
			Validate: func(selection interface{}) error {
				selected := selection.([]core.OptionAnswer)
				for _, ns := range selected {
					if ns.Value == "*" && len(selected) > 1 {
						return errors.New("all namespaces (*) is selected; please deselect it or deselect the other namespaces")
					}
				}
				return nil
			},
		},
	}

	if err := survey.Ask(qs, &webhookSpecIn); err != nil {
		return webHookConfigSpec{}, errors.WrapIf(err, "failed to read webhook configuration value")
	}

	return webhookSpecIn, nil
}

func (sa *specAssembler) securityScanSpecAsMap(spec *ServiceSpec) (map[string]interface{}, error) {
	// fill the structure of the config - make filling up the values easier
	if spec == nil {
		spec = &ServiceSpec{
			CustomAnchore:    anchoreSpec{},
			Policy:           policySpec{},
			ReleaseWhiteList: nil,
			WebhookConfig:    webHookConfigSpec{},
		}
	}

	var specMap map[string]interface{}
	if err := mapstructure.Decode(spec, &specMap); err != nil {
		return nil, err
	}

	return specMap, nil
}

func (sa *specAssembler) askForWhiteLists() ([]releaseSpec, error) {

	addMore := true
	releaseWhiteList := make([]releaseSpec, 0)

	for addMore {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Add a release whitelist item to the security scan? ",
			},
			&addMore,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}

		if !addMore {
			continue
		}

		item, err := sa.askForWhiteListItem()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to read release whitelist item")
		}
		releaseWhiteList = append(releaseWhiteList, *item)
	}

	return releaseWhiteList, nil
}

func (sa *specAssembler) askForWhiteListItem() (*releaseSpec, error) {

	var releaseName string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the name of the release whitelist item:",
		},
		&releaseName,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the name of the release whitelist item")
	}

	var reason string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the reason of the release whitelist item:",
		},
		&reason,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the reason of the release whitelist item")
	}

	var regexp string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the regexp for the release whitelist item:",
		},
		&regexp,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the regexp of the release whitelist item")
	}

	return &releaseSpec{
		Name:   releaseName,
		Reason: reason,
		Regexp: regexp,
	}, nil
}

func (sa specAssembler) assembleServiceSpec(ctx context.Context, orgID int32, clusterID int32, serviceSpecIn ServiceSpec) (ServiceSpec, error) {

	policy, err := sa.askForPolicy(serviceSpecIn.Policy)
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assembele policy data")
	}

	webhookConfig, err := sa.askForWebHookConfig(ctx, orgID, clusterID, serviceSpecIn.WebhookConfig)
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assembele webhook data")

	}

	return ServiceSpec{
		Policy:        policy,
		WebhookConfig: webhookConfig,
	}, nil
}
