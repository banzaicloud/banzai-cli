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
	"context"
	"encoding/json"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type ActivateManager struct{}

func (m *ActivateManager) GetName() string {
	return featureName
}

func (m *ActivateManager) BuildRequestInteractively(banzaiCLI cli.Cli) (*pipeline.ActivateClusterFeatureRequest, error) {
	var request pipeline.ActivateClusterFeatureRequest

	vaultType, err := askVaultComponent(vaultCustom)
	if err != nil {
		return nil, errors.WrapIf(err, "error during choosing Vault type")
	}

	switch vaultType {
	case vaultCustom:
		customSpec, err := buildCustomVaultFeatureRequest(banzaiCLI, defaults{})
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

func buildCustomVaultFeatureRequest(banzaiCLI cli.Cli, defaults defaults) (map[string]interface{}, error) {
	address, err := askVaultAddress(defaults.address)
	if err != nil {
		return nil, err
	}

	secretID, err := askVaultSecret(banzaiCLI, defaults.secretID)
	if err != nil {
		return nil, err
	}

	policy, err := askPolicy(defaults.policy)
	if err != nil {
		return nil, err
	}

	return obj{
		"customVault": obj{
			"enabled":  true,
			"address":  address,
			"secretId": secretID,
			"policy":   policy,
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

func askVaultSecret(banzaiCLI cli.Cli, defaultValue string) (string, error) {
	orgID := banzaiCLI.Context().OrganizationID()
	secrets, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(
		context.Background(),
		orgID,
		&pipeline.GetSecretsOpts{
			Type_: optional.NewString(vaultSecretType),
		},
	)
	if err != nil {
		return "", errors.WrapIf(err, "failed to get Vault secret(s)")
	}

	if len(secrets) == 0 {
		// TODO (colin): add option to create new Vault secret
		return "", errors.New("there's no Vault secrets, create a new one")
	}

	const skip = "skip"

	var secretName string
	var defaultSecretName = skip
	secretOptions := make([]string, len(secrets)+1)
	secretIds := make(map[string]string, len(secrets))
	secretOptions[0] = skip
	for i, s := range secrets {
		secretOptions[i+1] = s.Name
		secretIds[s.Name] = s.Id
		if s.Id == defaultValue {
			defaultSecretName = s.Name
		}
	}
	err = survey.AskOne(
		&survey.Select{
			Message: "Provide secret to access Vault:",
			Options: secretOptions,
			Default: defaultSecretName,
		},
		&secretName,
	)
	if err != nil {
		return "", errors.WrapIf(err, "failed to select secret")
	}

	if secretName == skip {
		return "", nil
	}

	return secretIds[secretName], nil
}

func askPolicy(defaultValue string) (string, error) {
	var policy string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please provide policy: ",
			Default: defaultValue,
		},
		&policy,
	); err != nil {
		return "", errors.WrapIf(err, "failure during survey")
	}
	return policy, nil
}
