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
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type Manager struct {
	banzaiCLI cli.Cli
}

func NewManager(banzaiCLI cli.Cli) Manager {
	return Manager{
		banzaiCLI: banzaiCLI,
	}
}

func (Manager) ReadableName() string {
	return "Vault"
}

func (Manager) ServiceName() string {
	return "vault"
}

func (m Manager) BuildActivateRequestInteractively(clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {
	var request pipeline.ActivateIntegratedServiceRequest

	vaultType, err := askVaultComponent(vaultCustom)
	if err != nil {
		return request, errors.WrapIf(err, "error during choosing Vault type")
	}

	switch vaultType {
	case vaultCustom:
		customSpec, err := buildCustomVaultServiceRequest(m.banzaiCLI, defaults{})
		if err != nil {
			return request, errors.Wrap(err, "failed to build custom Vault integratedservice request")
		}
		request.Spec = customSpec
	case vaultCP:
	default:
		return request, errors.New("not supported type of Vault component")
	}

	settings, err := buildSettingsServiceRequest(
		defaults{
			namespaces:      []string{"*"},
			serviceAccounts: []string{"*"},
		},
	)
	if err != nil {
		return request, errors.WrapIf(err, "failed to build settings to integratedservice activate request")
	}

	if request.Spec == nil {
		request.Spec = make(map[string]interface{}, 1)
	}

	request.Spec["settings"] = settings

	return request, nil
}

func (m Manager) BuildUpdateRequestInteractively(clusterCtx clustercontext.Context, request *pipeline.UpdateIntegratedServiceRequest) error {

	var spec specResponse
	if err := mapstructure.Decode(request.Spec, &spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	currentVaultType := vaultCP
	isCustomVault := spec.CustomVault.Enabled
	if isCustomVault {
		currentVaultType = vaultCustom
	}

	vaultType, err := askVaultComponent(currentVaultType)
	if err != nil {
		return errors.WrapIf(err, "error during choosing Vault type")
	}

	switch vaultType {
	case vaultCustom:
		customSpec, err := buildCustomVaultServiceRequest(m.banzaiCLI, defaults{
			address:  spec.CustomVault.Address,
			secretID: spec.CustomVault.SecretID,
			policy:   spec.CustomVault.Policy,
		})
		if err != nil {
			return errors.Wrap(err, "failed to build custom Vault integratedservice request")
		}
		request.Spec = customSpec
	case vaultCP:
	default:
		return errors.New("not supported type of Vault component")
	}

	settings, err := buildSettingsServiceRequest(
		defaults{
			namespaces:      spec.Settings.Namespaces,
			serviceAccounts: spec.Settings.ServiceAccounts,
		},
	)
	if err != nil {
		return errors.WrapIf(err, "failed to build settings to integratedservice update request")
	}

	request.Spec["settings"] = settings

	return nil
}

func (Manager) ValidateSpec(specObj map[string]interface{}) error {
	var spec struct {
		CustomVault struct {
			Enabled  bool   `mapstructure:"enabled"`
			SecretID string `mapstructure:"secretId"`
			Address  string `mapstructure:"address"`
			Policy   string `mapstructure:"policy"`
		} `mapstructure:"customVault"`
		Settings struct {
			Namespaces      []string `mapstructure:"namespaces"`
			ServiceAccounts []string `mapstructure:"serviceAccounts"`
		} `mapstructure:"settings"`
	}

	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "integratedservice specification does not conform to schema")
	}

	if spec.CustomVault.Enabled {
		// address is required in case of custom Vault
		if spec.CustomVault.Address == "" {
			return errors.New("Vault address cannot be empty in case of custom Vault option")
		}

		// policy is required in case of custom Vault with token
		if spec.CustomVault.Policy == "" && spec.CustomVault.SecretID != "" {
			return errors.New("policy field is required in case of custom Vault")
		}
	}

	if len(spec.Settings.Namespaces) == 1 && spec.Settings.Namespaces[0] == "*" &&
		len(spec.Settings.ServiceAccounts) == 1 && spec.Settings.ServiceAccounts[0] == "*" {
		return errors.New(`both namespaces and service accounts cannot be "*"`)
	}

	return nil
}

type outputResponse struct {
	Vault struct {
		Version        string `mapstructure:"version"`
		AuthMethodPath string `mapstructure:"authMethodPath"`
		Role           string `mapstructure:"role"`
		Policy         string `mapstructure:"policy"`
	} `mapstructure:"vault"`
	Wehhook struct {
		Version string `mapstructure:"version"`
	} `mapstructure:"webhook"`
}

type specResponse struct {
	CustomVault struct {
		Enabled  bool   `json:"enabled"`
		Address  string `json:"address"`
		SecretID string `json:"secretId"`
		Policy   string `json:"policy"`
	} `json:"customVault"`
	Settings struct {
		Namespaces      []string `json:"namespaces"`
		ServiceAccounts []string `json:"serviceAccounts"`
	} `json:"settings"`
}

func (Manager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	var output outputResponse
	if err := mapstructure.Decode(details.Output, &output); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return nil
	}

	var spec specResponse
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return nil
	}

	tableData["Vault_version"] = output.Vault.Version
	tableData["Auth_method_path"] = output.Vault.AuthMethodPath
	tableData["Role"] = output.Vault.Role
	tableData["Webhook_version"] = output.Wehhook.Version
	tableData["Namespaces"] = spec.Settings.Namespaces
	tableData["Service_accounts"] = spec.Settings.ServiceAccounts

	var policy string
	if spec.CustomVault.Enabled {
		tableData["Vault_address"] = spec.CustomVault.Address
		if spec.CustomVault.SecretID != "" {
			tableData["SecretID"] = spec.CustomVault.SecretID
		}

		policy = spec.CustomVault.Policy
	} else {
		policy = output.Vault.Policy
	}

	tableData["Policy"] = fmt.Sprintf("%q", strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(policy, "\t", " "), "\n", "")))

	return map[string]map[string]interface{}{
		"Vault": tableData,
	}
}

func buildCustomVaultServiceRequest(banzaiCLI cli.Cli, defaults defaults) (map[string]interface{}, error) {
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

func buildSettingsServiceRequest(defaults defaults) (map[string]interface{}, error) {
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
