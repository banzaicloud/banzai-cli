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
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type GetManager struct{}

func (GetManager) GetName() string {
	return serviceName
}

func (GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]map[string]interface{} {
	// helper for response processing
	type serviceDetails struct {
		Spec   ServiceSpec   `json:"spec" mapstructure:"spec"`
		Output ServiceOutput `json:"output" mapstructure:"output"`
	}

	tableData := map[string]interface{}{}

	boundResponse := serviceDetails{}
	if err := mapstructure.Decode(details, &boundResponse); err != nil {
		tableData["error"] = fmt.Sprintf("failed to decode spec %q", err)
		return map[string]map[string]interface{}{
			"DNS": tableData,
		}
	}

	tableData["Status"] = details.Status
	tableData["Version"] = boundResponse.Output.ExternalDns.Version

	return map[string]map[string]interface{}{
		"DNS": tableData,
	}
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
