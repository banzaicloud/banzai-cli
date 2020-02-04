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

package expiry

import (
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type GetManager struct{}

func (GetManager) GetName() string {
	return serviceName
}

type TableData map[string]interface{}

func (GetManager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {

	const (
		expiryTitle = "Expiry"
		statusTitle = "Status"
		dateTitle   = "Date"
	)

	var baseOutput = map[string]map[string]interface{}{
		expiryTitle: {
			statusTitle: details.Status,
		},
	}

	if details.Status == "INACTIVE" {
		return baseOutput
	}

	var spec serviceSpec
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return baseOutput
	}

	return map[string]map[string]interface{}{
		expiryTitle: {
			statusTitle: details.Status,
			dateTitle:   spec.Date,
		},
	}
}

func NewGetManager() *GetManager {
	return &GetManager{}
}
