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
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type getManager struct {
	baseManager
}

func NewGetManager() *getManager {
	return new(getManager)
}

// todo remove this once the called method gets renamed
func (g getManager) GetCommandName() string {
	return g.GetName()
}

func (g getManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	if anchore, ok := getObj(details.Output, "anchore"); ok {
		if aVersion, ok := getStr(anchore, "version"); ok {
			tableData["AnchoreVersion"] = aVersion
		}
	}

	if imgValidator, ok := getObj(details.Output, "imageValidator"); ok {
		if aVersion, ok := getStr(imgValidator, "version"); ok {
			tableData["ImageValidatorVersion"] = aVersion
		}
	}

	return tableData
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
