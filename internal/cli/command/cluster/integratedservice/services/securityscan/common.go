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
	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
)

var (
	policyBundles = utils.IdToNameMap{
		"2c53a13c-1765-11e8-82ef-23527761d060": "Default bundle",
		"deny_all":                             "Deny all",
		"reject_high":                          "Reject high severity",
		"reject_critical":                      "Reject critical severity",
		"block_root":                           "Block root privileges",
		"allow_all_and_warn":                   "Allow all and warning",
		"Custom":                               "Custom",
	}
)

//ServiceSpec security scan cluster integratedservice specific specification
type ServiceSpec struct {
	CustomAnchore    anchoreSpec       `json:"customAnchore" mapstructure:"customAnchore"`
	Policy           policySpec        `json:"policy" mapstructure:"policy"`
	ReleaseWhiteList []releaseSpec     `json:"releaseWhiteList,omitempty" mapstructure:"releaseWhiteList"`
	WebhookConfig    webHookConfigSpec `json:"webhookConfig" mapstructure:"webhookConfig"`
	Registry         *registrySpec     `json:"registry,omitempty" mapstructure:"registry"`
}

// Validate validates the input security scan specification.
func (s ServiceSpec) Validate() error {
	var validationErrors error

	if s.CustomAnchore.Enabled {
		validationErrors = s.CustomAnchore.Validate()
	}

	if !s.Policy.CustomPolicy.Enabled && s.Policy.PolicyID == "" {
		validationErrors = errors.Combine(validationErrors, errors.New("policyId is required"))
	}

	for _, releaseSpec := range s.ReleaseWhiteList {
		validationErrors = errors.Combine(validationErrors, releaseSpec.Validate())
	}

	validationErrors = errors.Combine(validationErrors, s.WebhookConfig.Validate())

	if s.Registry != nil {
		validationErrors = errors.Combine(validationErrors, s.Registry.Validate())
	}

	return validationErrors
}

type anchoreSpec struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Url      string `json:"url" mapstructure:"url"`
	SecretID string `json:"secretId" mapstructure:"secretId"`
}

func (a anchoreSpec) Validate() error {
	if a.Enabled {
		if a.Url == "" && a.SecretID == "" {
			return errors.New("both anchore url and secretId are required")
		}
	}

	return nil
}

type policySpec struct {
	PolicyID     string           `json:"policyId,omitempty" mapstructure:"policyId"`
	CustomPolicy customPolicySpec `json:"customPolicy,omitempty" mapstructure:"customPolicy"`
}

type customPolicySpec struct {
	Enabled bool                   `json:"enabled" mapstructure:"enabled"`
	Policy  map[string]interface{} `json:"policy" mapstructure:"policy"`
}

type releaseSpec struct {
	Name   string `json:"name" mapstructure:"name"`
	Reason string `json:"reason" mapstructure:"reason"`
	Regexp string `json:"regexp,omitempty" mapstructure:"regexp"`
}

func (r releaseSpec) Validate() error {
	if r.Name == "" || r.Reason == "" {
		return errors.NewPlain("both name and reason must be specified")
	}

	return nil
}

type webHookConfigSpec struct {
	Enabled    bool     `json:"enabled" mapstructure:"enabled"`
	Selector   string   `json:"selector" mapstructure:"selector"`
	Namespaces []string `json:"namespaces" mapstructure:"namespaces"`
}

func (w webHookConfigSpec) Validate() error {
	if w.Enabled {
		if w.Selector == "" || len(w.Namespaces) < 1 {
			return errors.NewPlain("selector and namespaces must be filled")
		}
	}

	return nil
}

type registrySpec struct {
	Type     string `json:"type" mapstructure:"type"`
	Registry string `json:"registry" mapstructure:"registry"`
	SecretID string `json:"secretId" mapstructure:"secretId"`
	Insecure bool   `json:"insecure" mapstructure:"insecure"`
}

func (s registrySpec) Validate() error {
	if s.Registry == "" || s.SecretID == "" {
		return errors.New("both registry and secretId are required")
	}

	return nil
}
