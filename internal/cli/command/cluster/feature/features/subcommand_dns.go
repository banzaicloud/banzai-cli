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
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features/dns"
)

type DnsSubCommandManager struct{}

func (DnsSubCommandManager) GetName() string {
	return "DNS"
}

func (DnsSubCommandManager) ActivateManager() ActivateManager {
	return dns.NewActivateManager()
}

func (DnsSubCommandManager) DeactivateManager() DeactivateManager {
	return dns.NewDeactivateManager()
}

func (DnsSubCommandManager) GetManager() GetManager {
	return dns.NewGetManager()
}

func (DnsSubCommandManager) UpdateManager() UpdateManager {
	return dns.NewUpdateManager()
}

func NewDNSSubCommandManager() *DnsSubCommandManager {
	return &DnsSubCommandManager{}
}
