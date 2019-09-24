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
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features/monitoring"
)

type monitoringSubCommandManager struct {
}

func (scm *monitoringSubCommandManager) GetName() string {
	return "Monitoring"
}

func (scm *monitoringSubCommandManager) ActivateManager() ActivateManager {
	return monitoring.NewActivateManager()
}

func (scm *monitoringSubCommandManager) DeactivateManager() DeactivateManager {
	return monitoring.NewDeactivateManager()
}

func (scm *monitoringSubCommandManager) GetManager() GetManager {
	return monitoring.NewGetManager()
}

func (scm *monitoringSubCommandManager) UpdateManager() UpdateManager {
	return monitoring.UpdateGetManager()
}

func NewMonitoringSubCommandManager() *monitoringSubCommandManager {
	return &monitoringSubCommandManager{}
}
