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

package features

import (
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features/vault"
)

type VaultSubCommandManager struct{}

func (VaultSubCommandManager) GetName() string {
	return "Vault"
}

func (VaultSubCommandManager) ActivateManager() ActivateManager {
	return vault.NewActivateManager()
}

func (VaultSubCommandManager) DeactivateManager() DeactivateManager {
	return vault.NewDeactivateManager()
}

func (VaultSubCommandManager) GetManager() GetManager {
	return vault.NewGetManager()
}

func (VaultSubCommandManager) UpdateManager() UpdateManager {
	return vault.NewUpdateManager()
}

func NewVaultSubCommandManager() *VaultSubCommandManager {
	return &VaultSubCommandManager{}
}
