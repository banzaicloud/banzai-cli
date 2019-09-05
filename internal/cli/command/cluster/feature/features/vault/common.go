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
	"fmt"

	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"
)

const (
	featureName = "vault"

	vaultCustom = "Custom vault"
	vaultCP     = "Pipeline's Vault"
)

type obj = map[string]interface{}

type defaults struct {
	address         string
	token           string
	namespaces      []string
	serviceAccounts []string
}

func NewDeactivateManager() *ActivateManager {
	return &ActivateManager{}
}

func validateSpec(specObj map[string]interface{}) error {

	var spec struct {
		CustomVault struct {
			Enabled bool   `mapstructure:"enabled"`
			Token   string `mapstructure:"token"`
			Address string `mapstructure:"address"`
		} `mapstructure:"customVault"`
		Settings struct {
			Namespaces      []string `mapstructure:"namespaces"`
			ServiceAccounts []string `mapstructure:"serviceAccounts"`
		} `mapstructure:"settings"`
	}

	if err := mapstructure.Decode(specObj, &spec); err != nil {
		return errors.WrapIf(err, "feature specification does not conform to schema")
	}
	fmt.Println("spec", spec)

	if spec.CustomVault.Enabled && len(spec.CustomVault.Address) == 0 {
		return errors.New("Vault address cannot be empty in case of custom Vault option")
	}

	if len(spec.Settings.Namespaces) == 1 && spec.Settings.Namespaces[0] == "*" &&
		len(spec.Settings.ServiceAccounts) == 1 && spec.Settings.ServiceAccounts[0] == "*" {
		return errors.New("both namespaces and service accounts can not be \"*\"")
	}

	return nil
}
