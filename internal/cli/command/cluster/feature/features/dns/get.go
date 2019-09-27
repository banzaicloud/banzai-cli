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

package dns

import (
	"strings"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type GetManager struct{}

func (GetManager) GetName() string {
	return featureName
}

func (GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	if autodns, ok := getObj(details.Output, "autoDns"); ok {
		if zone, ok := getStr(autodns, "zone"); ok {
			tableData["AutoDNS_zone"] = zone
		}
		if clusterDomain, ok := getStr(autodns, "clusterDomain"); ok {
			tableData["AutoDNS_cluster domain"] = clusterDomain
		}
	}

	if customDNS, ok := getObj(details.Spec, "customDns"); ok {
		if clusterDomain, ok := getObj(customDNS, "clusterDomain"); ok {
			tableData["CustomDNS_cluster_domain"] = clusterDomain
		}
		if domainFilters, ok := getList(customDNS, "domainFilters"); ok {
			filters := make([]string, 0, len(domainFilters))
			for _, f := range domainFilters {
				if s, ok := f.(string); ok {
					filters = append(filters, s)
				}
			}
			tableData["CustomDNS_domain_filters"] = strings.Join(filters, ",")
		}
		if provider, ok := getObj(customDNS, "provider"); ok {
			if name, ok := getStr(provider, "name"); ok {
				tableData["CustomDNS_provider"] = name
			}
		}
	}

	return tableData
}

func NewGetManager() *GetManager {
	return &GetManager{}
}

func getList(target map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := target[key]; ok {
		if list, ok := value.([]interface{}); ok {
			return list, true
		}
	}
	return nil, false
}

func getObj(target map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := target[key]; ok {
		if obj, ok := value.(map[string]interface{}); ok {
			return obj, true
		}
	}
	return nil, false
}

func getStr(target map[string]interface{}, key string) (string, bool) {
	if value, ok := target[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}
