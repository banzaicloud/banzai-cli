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

package monitoring

import (
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type GetManager struct{}

func (GetManager) GetName() string {
	return featureName
}

type baseOutputItems struct {
	Url      string `mapstructure:"url"`
	SecretID string `mapstructure:"secretId"`
	Version  string `mapstructure:"version"`
}

type outputResponse struct {
	Alertmanager struct {
		baseOutputItems `mapstructure:",squash"`
	} `mapstructure:"alertmanager"`
	Grafana struct {
		baseOutputItems `mapstructure:",squash"`
	} `mapstructure:"grafana"`
	Prometheus struct {
		baseOutputItems `mapstructure:",squash"`
	} `mapstructure:"prometheus"`
	PrometheusOperator struct {
		Version string `mapstructure:"version"`
	} `mapstructure:"prometheusOperator"`
	Pushgateway struct {
		Version string `mapstructure:"version"`
	} `mapstructure:"pushgateway"`
}

func (GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	var output outputResponse
	if err := mapstructure.Decode(details.Output, &output); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return tableData
	}

	var spec featureSpec
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return tableData
	}

	// Alertmanager outputs
	if spec.Alertmanager.Enabled {
		tableData["Alertmanager_url"] = output.Alertmanager.Url
		tableData["Alertmanager_version"] = output.Alertmanager.Version

		tableData["Alertmanager_public"] = spec.Alertmanager.Public.Enabled
	}

	// Grafana outputs
	if spec.Grafana.Enabled {
		var secretID = spec.Grafana.SecretId
		if secretID == "" {
			secretID = output.Grafana.SecretID
		}
		tableData["Grafana_url"] = output.Grafana.Url
		tableData["Grafana_secretID"] = secretID
		tableData["Grafana_version"] = output.Grafana.Version

		tableData["Grafana_public"] = spec.Grafana.Public.Enabled
	}

	// Prometheus outputs
	if spec.Prometheus.Enabled {
		var secretID = spec.Prometheus.SecretId
		if secretID == "" {
			secretID = output.Prometheus.SecretID
		}
		tableData["Prometheus_url"] = output.Prometheus.Url
		tableData["Prometheus_secretID"] = secretID
		tableData["Prometheus_version"] = output.Prometheus.Version

		tableData["Prometheus_public"] = spec.Prometheus.Public.Enabled
	}

	tableData["Pushgateway_version"] = output.Pushgateway.Version
	tableData["Prometheus_operator_version"] = output.PrometheusOperator.Version

	return tableData
}

func NewGetManager() *GetManager {
	return &GetManager{}
}
