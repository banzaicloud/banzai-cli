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

package logging

import (
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type GetManager struct{}

func (GetManager) GetName() string {
	return serviceName
}

type outputResponse struct {
	Logging struct {
		OperatorVersion  string `mapstructure:"operatorVersion"`
		FluentdVersion   string `mapstructure:"fluentdVersion"`
		FluentbitVersion string `mapstructure:"fluentbitVersion"`
	} `mapstructure:"logging"`
	Loki struct {
		Url        string `mapstructure:"url"`
		Version    string `mapstructure:"version"`
		ServiceURL string `mapstructure:"serviceUrl"`
		SecretID   string `mapstructure:"secretId"`
	} `mapstructure:"loki"`
}

type TableData map[string]interface{}

func (GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	var output outputResponse
	if err := mapstructure.Decode(details.Output, &output); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return map[string]map[string]interface{}{
			"Logging": tableData,
		}
	}

	var spec spec
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return map[string]map[string]interface{}{
			"Logging": tableData,
		}
	}

	if spec.Loki.Enabled && spec.Loki.Ingress.Enabled {
		var secretID string
		if spec.Loki.Ingress.Enabled {
			secretID = spec.Loki.Ingress.SecretID
			if secretID == "" {
				secretID = output.Loki.SecretID
			}
		}
		var lokiTable = TableData{
			"url":        output.Loki.Url,
			"version":    output.Loki.Version,
			"serviceUrl": output.Loki.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Loki.Ingress.Path,
			"domain":     spec.Loki.Ingress.Domain,
		}
		tableData["Loki"] = lokiTable
	}

	tableData["Logging"] = TableData{
		"metrics": spec.Logging.Metrics,
		"TLS":     spec.Logging.TLS,
	}

	return map[string]map[string]interface{}{
		"Logging": tableData,
	}
}

func NewGetManager() *GetManager {
	return &GetManager{}
}
