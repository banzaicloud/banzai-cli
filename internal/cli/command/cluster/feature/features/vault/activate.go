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

package vault

import (
	"encoding/json"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type ActivateManager struct{}

func (m *ActivateManager) GetName() string {
	return featureName
}

func (m *ActivateManager) BuildRequestInteractively(_ cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {
	var request pipeline.ActivateClusterFeatureRequest

	vaultType, err := askVaultComponent(vaultCustom)
	if err != nil {
		return nil, errors.WrapIf(err, "error during choosing Vault type")
	}

	switch vaultType {
	case vaultCustom:
		customSpec, err := buildCustomVaultFeatureRequest(defaults{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to build custom Vault feature request")
		}
		request.Spec = customSpec
	case vaultCP:
	default:
		return nil, errors.New("not supported type of Vault component")
	}

	settings, err := buildSettingsFeatureRequest(
		defaults{
			namespaces:      []string{"*"},
			serviceAccounts: []string{"*"},
		},
	)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to build settings to feature activate request")
	}

	if request.Spec == nil {
		request.Spec = make(map[string]interface{}, 0)
	}

	request.Spec["settings"] = settings

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

func buildCustomVaultFeatureRequest(defaults defaults) (map[string]interface{}, error) {
	address, err := askVaultAddress(defaults.address)
	if err != nil {
		return nil, err
	}

	token, err := askVaultToken(defaults.token)
	if err != nil {
		return nil, err
	}

	return obj{
		"customVault": obj{
			"enabled": true,
			"address": address,
			"token":   token,
		},
	}, nil
}

func buildSettingsFeatureRequest(defaults defaults) (map[string]interface{}, error) {
	namespaces, err := askNamespaces(defaults.namespaces)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get namespaces")
	}

	serviceAccounts, err := askServiceAccounts(defaults.serviceAccounts)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get service accounts")
	}

	return obj{
		"namespaces":      namespaces,
		"serviceAccounts": serviceAccounts,
	}, nil
}

func askVaultComponent(defaultValue string) (string, error) {
	var comp string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Please select a Vault component to activate:",
			Options: []string{vaultCustom, vaultCP},
			Default: defaultValue,
		},
		&comp,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	return comp, nil
}

func askNamespaces(defaultValues []string) ([]string, error) {
	// todo (colin) use multiselect in v2
	var response string
	if err := survey.AskOne(
		&survey.Input{
			Message: "List of namespaces allowed to access:",
			Default: strings.Join(defaultValues, ","),
			Help:    "Use '*' to select all. To add multiple namespaces separate with comma (,) character. Like: default, pipeline-system",
		},
		&response,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	items := strings.Split(response, ",")
	namespaces := make([]string, 0, len(items))
	for _, s := range items {
		namespaces = append(namespaces, strings.TrimSpace(s))
	}

	return namespaces, nil
}

func askServiceAccounts(defaultValues []string) ([]string, error) {
	// todo (colin) use multiselect in v2
	var response string
	if err := survey.AskOne(
		&survey.Input{
			Message: "List of service account names able to access:",
			Default: strings.Join(defaultValues, ","),
			Help:    "Use '*' to select all. To add multiple service account names separate with commna (,) character",
		},
		&response,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	items := strings.Split(response, ",")
	serviceAccounts := make([]string, 0, len(items))
	for _, s := range items {
		serviceAccounts = append(serviceAccounts, strings.TrimSpace(s))
	}

	return serviceAccounts, nil
}

func askVaultAddress(defaultValue string) (string, error) {
	var address string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please provide a Vault address",
			Default: defaultValue,
		},
		&address,
		survey.WithValidator(survey.Required),
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	return address, nil
}

func askVaultToken(defaultValue string) (string, error) {
	var token string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please provide a Vault token",
			Default: defaultValue,
		},
		&token,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	return token, nil
}
