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
	"fmt"
)

type featureSpec struct {
	Grafana      grafanaAndPrometheusSpec `json:"grafana" mapstructure:"grafana"`
	Alertmanager alertmanagerSpec         `json:"alertmanager" mapstructure:"alertmanager"`
	Prometheus   grafanaAndPrometheusSpec `json:"prometheus" mapstructure:"prometheus"`
}

type baseSpec struct {
	Enabled bool        `json:"enabled" mapstructure:"enabled"`
	Public  ingressSpec `json:"public,omitempty" mapstructure:"public"`
}

type grafanaAndPrometheusSpec struct {
	baseSpec `mapstructure:",squash"`
	SecretId string `json:"secretId,omitempty" mapstructure:"secretId"`
}

type alertmanagerSpec struct {
	baseSpec `mapstructure:",squash"`
	Provider providerSpec `json:"provider" mapstructure:"provider"`
}

type providerSpec struct {
	Slack     slackPropertiesSpec     `json:"slack" mapstructure:"slack"`
	Pagerduty pagerdutyPropertiesSpec `json:"pagerduty" mapstructure:"pagerduty"`
	Email     emailPropertiesSpec     `json:"email" mapstructure:"email"`
}

type slackPropertiesSpec struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	ApiUrl       string `json:"apiUrl,omitempty" mapstructure:"apiUrl"`
	Channel      string `json:"channel,omitempty" mapstructure:"channel"`
	SendResolved bool   `json:"sendResolved" mapstructure:"sendResolved"`
}

type emailPropertiesSpec struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	To           string `json:"to,omitempty" mapstructure:"to"`
	From         string `json:"from,omitempty" mapstructure:"from"`
	SendResolved bool   `json:"sendResolved" mapstructure:"sendResolved"`
}

type pagerdutyPropertiesSpec struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	RoutingKey   string `json:"routingKey,omitempty" mapstructure:"routingKey"`
	ServiceKey   string `json:"serviceKey,omitempty" mapstructure:"serviceKey"`
	Url          string `json:"url,omitempty" mapstructure:"url"`
	SendResolved bool   `json:"sendResolved" mapstructure:"sendResolved"`
}

type ingressSpec struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Domain  string `json:"domain,omitempty" mapstructure:"domain"`
	Path    string `json:"path,omitempty" mapstructure:"path"`
}

type requiredFieldError struct {
	fieldName string
}

func (e requiredFieldError) Error() string {
	return fmt.Sprintf("%q cannot be empty", e.fieldName)
}

func (s ingressSpec) Validate(ingressType string) error {
	if len(s.Path) == 0 {
		return requiredFieldError{fieldName: fmt.Sprintf("%s path", ingressType)}
	}

	return nil
}
