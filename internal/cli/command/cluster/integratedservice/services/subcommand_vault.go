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

package services

import (
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/vault"
)

type VaultSubCommandManager struct{}

func (VaultSubCommandManager) GetName() string {
	return "Vault"
}

func (VaultSubCommandManager) ActivateManager() ActivateManager {
	return vault.Manager{}
}

func (VaultSubCommandManager) DeactivateManager() DeactivateManager {
	return vault.Manager{}
}

func (VaultSubCommandManager) GetManager() GetManager {
	return vault.Manager{}
}

func (VaultSubCommandManager) UpdateManager() UpdateManager {
	return vault.Manager{}
}

func NewVaultSubCommandManager() VaultSubCommandManager {
	return VaultSubCommandManager{}
}
