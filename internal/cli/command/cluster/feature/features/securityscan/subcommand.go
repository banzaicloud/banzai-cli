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
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features"
)

type subCommandManager struct {
}

func (scm *subCommandManager) GetName() string {
	return "securityscan"
}

func (scm *subCommandManager) ActivateManager() features.ActivateManager {
	return NewActivateManager()
}

func (scm *subCommandManager) DeactivateManager() features.DeactivateManager {
	return NewDeactivateManager()
}

func (scm *subCommandManager) GetManager() features.GetManager {
	return NewGetManager()
}

func (scm *subCommandManager) UpdateManager() features.UpdateManager {
	return NewUpdateManager()
}

func NewSecurityScanSubCommandManager() *subCommandManager {
	return &subCommandManager{}
}
