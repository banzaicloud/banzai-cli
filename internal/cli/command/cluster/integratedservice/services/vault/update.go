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
	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"

	"github.com/mitchellh/mapstructure"
)

type UpdateManager struct {
	baseManager
}

func (UpdateManager) ValidateSpec(spec map[string]interface{}) error {
	return validateSpec(spec)
}

func (UpdateManager) BuildUpdateRequestInteractively(banzaiCli cli.Cli, updateServiceRequest *pipeline.UpdateIntegratedServiceRequest, clusterCtx clustercontext.Context) error {

	var spec specResponse
	if err := mapstructure.Decode(updateServiceRequest.Spec, &spec); err != nil {
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
		customSpec, err := buildCustomVaultServiceRequest(banzaiCli, defaults{
			address:  spec.CustomVault.Address,
			secretID: spec.CustomVault.SecretID,
			policy:   spec.CustomVault.Policy,
		})
		if err != nil {
			return errors.Wrap(err, "failed to build custom Vault integratedservice request")
		}
		updateServiceRequest.Spec = customSpec
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

	updateServiceRequest.Spec["settings"] = settings

	return nil
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{}
}
