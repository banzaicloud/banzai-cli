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
	Url        string `mapstructure:"url"`
	SecretID   string `mapstructure:"secretId"`
	Version    string `mapstructure:"version"`
	ServiceURL string `mapstructure:"serviceUrl"`
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
		baseOutputItems `mapstructure:",squash"`
	} `mapstructure:"pushgateway"`
}

type TableData map[string]interface{}

func (GetManager) WriteDetailsTable(details pipeline.ClusterFeatureDetails) map[string]map[string]interface{} {
	tableData := map[string]map[string]interface{}{
		"Monitoring": {
			"Status": details.Status,
		},
	}

	if details.Status == "INACTIVE" {
		return tableData
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
		var secretID string
		if spec.Alertmanager.Ingress.Enabled {
			secretID = spec.Alertmanager.Ingress.SecretId
			if secretID == "" {
				secretID = output.Alertmanager.SecretID
			}
		}
		var alertmanagerTable = TableData{
			"url":        output.Alertmanager.Url,
			"version":    output.Alertmanager.Version,
			"serviceUrl": output.Alertmanager.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Alertmanager.Ingress.Path,
			"domain":     spec.Alertmanager.Ingress.Domain,
		}
		tableData["Alertmanager"] = alertmanagerTable
		// todo (colin): add provider outputs
	}

	// Grafana outputs
	if spec.Grafana.Enabled {
		var secretID = spec.Grafana.SecretId
		if secretID == "" {
			secretID = output.Grafana.SecretID
		}
		var grafanaTable = TableData{
			"url":        output.Grafana.Url,
			"version":    output.Grafana.Version,
			"serviceUrl": output.Grafana.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Grafana.Ingress.Path,
			"domain":     spec.Grafana.Ingress.Domain,
		}
		tableData["Grafana"] = grafanaTable
	}

	// Prometheus outputs
	if spec.Prometheus.Enabled {
		var secretID string
		if spec.Prometheus.Ingress.Enabled {
			secretID = spec.Prometheus.Ingress.SecretId
			if secretID == "" {
				secretID = output.Prometheus.SecretID
			}
		}
		var prometheusTable = TableData{
			"url":        output.Prometheus.Url,
			"version":    output.Prometheus.Version,
			"serviceUrl": output.Prometheus.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Prometheus.Ingress.Path,
			"domain":     spec.Prometheus.Ingress.Domain,
		}
		tableData["Prometheus"] = prometheusTable

		tableData["Prometheus_storage"] = TableData{
			"class":     spec.Prometheus.Storage.Class,
			"size":      spec.Prometheus.Storage.Size,
			"retention": spec.Prometheus.Storage.Retention,
		}
	}

	if spec.Pushgateway.Enabled {
		var secretID string
		if spec.Pushgateway.Ingress.Enabled {
			secretID = spec.Pushgateway.Ingress.SecretId
			if secretID == "" {
				secretID = output.Pushgateway.SecretID
			}
		}
		var pushgatewayTable = TableData{
			"url":        output.Pushgateway.Url,
			"version":    output.Pushgateway.Version,
			"serviceUrl": output.Pushgateway.ServiceURL,
			"secretID":   secretID,
			"path":       spec.Pushgateway.Ingress.Path,
			"domain":     spec.Pushgateway.Ingress.Domain,
		}
		tableData["Pushgateway"] = pushgatewayTable
	}

	if spec.Exporters.Enabled {
		tableData["Exporters"] = TableData{
			"nodeExporter":     spec.Exporters.NodeExporter.Enabled,
			"kubeStateMetrics": spec.Exporters.KubeStateMetrics.Enabled,
		}
	}

	tableData["Prometheus_operator"] = TableData{
		"verison": output.PrometheusOperator.Version,
	}

	return tableData
}

func NewGetManager() *GetManager {
	return &GetManager{}
}
