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
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/features/logging"
)

type LoggingSubCommandManager struct{}

func (LoggingSubCommandManager) GetName() string {
	return "Logging"
}

func (LoggingSubCommandManager) ActivateManager() ActivateManager {
	return logging.NewActivateManager()
}

func (LoggingSubCommandManager) DeactivateManager() DeactivateManager {
	return logging.NewDeactivateManager()
}

func (LoggingSubCommandManager) GetManager() GetManager {
	return logging.NewGetManager()
}

func (LoggingSubCommandManager) UpdateManager() UpdateManager {
	return logging.NewUpdateManager()
}

func NewLoggingSubCommandManager() *LoggingSubCommandManager {
	return &LoggingSubCommandManager{}
}
