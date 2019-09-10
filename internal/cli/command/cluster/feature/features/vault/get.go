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
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type GetManager struct{}

func (m *GetManager) GetCommandName() string {
	return featureName
}

type outputResponse struct {
	Vault struct {
		Version        string `mapstructure:"version"`
		AuthMethodPath string `mapstructure:"authMethodPath"`
		RolePath       string `mapstructure:"rolePath"`
	} `mapstructure:"vault"`
	Wehhook struct {
		Version string `mapstructure:"version"`
	} `mapstructure:"webhook"`
}

type specResponse struct {
	CustomVault struct {
		Enabled bool   `json:"enabled"`
		Address string `json:"address"`
		Token   string `json:"token"`
	} `json:"custom_vault"`
	Settings struct {
		Namespaces      []string `json:"namespaces"`
		ServiceAccounts []string `json:"serviceAccounts"`
	} `json:"settings"`
}

func (m *GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	var output outputResponse
	if err := mapstructure.Decode(details.Output, &output); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return tableData
	}

	var spec specResponse
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return tableData
	}

	tableData["Vault_version"] = output.Vault.Version
	tableData["Auth_method_path"] = output.Vault.AuthMethodPath
	tableData["Role_path"] = output.Vault.RolePath
	tableData["Webhook_version"] = output.Wehhook.Version
	tableData["Namespaces"] = spec.Settings.Namespaces
	tableData["Service_accounts"] = spec.Settings.ServiceAccounts

	if spec.CustomVault.Enabled {
		tableData["Vault_address"] = spec.CustomVault.Address
		tableData["Vault_token"] = spec.CustomVault.Token
	}

	return tableData
}

func NewGetManager() *GetManager {
	return &GetManager{}
}
