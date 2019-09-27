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
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/mitchellh/mapstructure"
)

const (
	featureName = "securityscan"
)

//SecurityScanFeatureSpec security scan cluster feature specific specification
type SecurityScanFeatureSpec struct {
	CustomAnchore    anchoreSpec       `json:"customAnchore" mapstructure:"customAnchore"`
	Policy           policySpec        `json:"policy" mapstructure:"policy"`
	ReleaseWhiteList []releaseSpec     `json:"releaseWhiteList,omitempty" mapstructure:"releaseWhiteList"`
	WebhookConfig    webHookConfigSpec `json:"webhookConfig" mapstructure:"webhookConfig"`
}

type baseManager struct{}

func (baseManager) GetName() string {
	return featureName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

// Validate validates the input security scan specification.
func (s SecurityScanFeatureSpec) Validate() error {

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

func (sa *specAssembler) askForAnchoreConfig(currentAnchoreSpec *anchoreSpec) (*anchoreSpec, error) {

	if currentAnchoreSpec == nil {
		currentAnchoreSpec = new(anchoreSpec)
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

func (sa *specAssembler) askForSecret(currentSecretID string) (string, error) {
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

func (sa *specAssembler) askForPolicy(currentPolicySpec *policySpec) (*policySpec, error) {
	if currentPolicySpec == nil {
		currentPolicySpec = new(policySpec)
	}

	type policy struct {
		name string
		id   string
	}
	// todo add all supported policies here
	policies := []policy{{"Default bundle", "1"}}

	options := make([]string, len(policies))
	currentPolicyName := policies[0].name

	for i, s := range policies {
		options[i] = s.name
		if s.id == currentPolicySpec.PolicyID {
			currentPolicyName = s.name
		}
	}

	var policyName string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a policy for the Anchor Engine:",
			Options: options,
			Default: currentPolicyName,
		},
		&policyName,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select policy")
	}

	var selected policy
	for _, s := range policies {
		if s.name == policyName {
			selected = s
		}
	}

	return &policySpec{
		PolicyID: selected.id,
	}, nil
}

func (sa *specAssembler) askForWebHookConfig(currentWebHookSpec *webHookConfigSpec) (*webHookConfigSpec, error) {
	if currentWebHookSpec == nil {
		currentWebHookSpec = new(webHookConfigSpec)
	}
	var enable bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Configure the security scan webhook?",
		},
		&enable,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !enable {
		return &webHookConfigSpec{
			Enabled: false,
		}, nil
	}

	if currentWebHookSpec.Selector == "" {
		// select the default selector
		currentWebHookSpec.Selector = "exclude"
	}

	var selector string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please choose the selector for the webhook configuration:",
			Options: []string{"exclude", "include"},
			Default: currentWebHookSpec.Selector,
		},
		&selector,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to select policy")
	}

	var namespaces string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the comma separated list of namespaces:",
			Default: strings.Join(currentWebHookSpec.Namespaces, ","),
		},
		&namespaces,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read namespaces")
	}

	return &webHookConfigSpec{
		Enabled:    true,
		Selector:   selector,
		Namespaces: strings.Split(namespaces, ","),
	}, nil
}

func (sa *specAssembler) securityScanSpecAsMap(spec *SecurityScanFeatureSpec) (map[string]interface{}, error) {
	// fill the structure of the config - make filling up the values easier
	if spec == nil {
		spec = &SecurityScanFeatureSpec{
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
