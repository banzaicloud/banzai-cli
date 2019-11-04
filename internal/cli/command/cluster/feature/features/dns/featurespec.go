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

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2/core"
)

// ExternalDNS  part of the DNS feature spec representation (used for validation, user input handling)
type ExternalDNS struct {
	DomainFilters []string  `json:"domainFilters" mapstructure:"domainFilters"`
	Policy        string    `json:"policy" mapstructure:"policy"` // sync | upsert-only
	Sources       []string  `json:"sources" mapstructure:"sources"`
	TxtOwnerId    string    `json:"txtOwnerId,omitempty" mapstructure:"txtOwnerId"`
	Provider      *Provider `json:"provider" mapstructure:"provider"`
}

// Provider DNS provider data
type Provider struct {
	Name     string                 `json:"name" mapstructure:"name"`
	SecretID string                 `json:"secretId,omitempty" mapstructure:"secretId"`
	Options  map[string]interface{} `json:"options,omitempty" mapstructure:"options"`
}

// DNSFeatureSpec DNS feature specification
type DNSFeatureSpec struct {
	ExternalDNS   ExternalDNS `mapstructure:"externalDns"`
	ClusterDomain string      `mapstructure:"clusterDomain"`
}

// DNSFeatureOutput used to parse / display feature output
type DNSFeatureOutput struct {
}

func (e ExternalDNS) Validate() error {
	var validationErrors error

	if len(e.DomainFilters) == 0 {
		validationErrors = errors.Append(validationErrors, errors.New("at least one domain filter must be specified"))
	}

	for _, df := range e.DomainFilters {
		if df == "" {
			validationErrors = errors.Append(validationErrors, errors.New("domain filters must not be empty strings"))
		}
	}

	if e.Policy == "" || (e.Policy != "sync" && e.Policy != "upsert-only") {
		validationErrors = errors.Append(validationErrors,
			errors.New("policy must not be empty, it should be one of the values sync|upsert-only"))
	}

	if len(e.Sources) == 0 {
		validationErrors = errors.Append(validationErrors, errors.New("sources must not be empty"))
	}

	for _, src := range e.Sources {
		if src != "service" && src != "ingress" {
			validationErrors = errors.Append(validationErrors, errors.Errorf("invalid source value: %s", src))
		}
	}

	if e.Provider == nil {
		validationErrors = errors.Append(validationErrors, errors.New("provider must be specified"))
	} else {
		validationErrors = errors.Append(validationErrors, e.Provider.Validate())
	}

	return validationErrors
}

// implement core.Settable in order to do transform the answer
func (e *ExternalDNS) WriteAnswer(field string, value interface{}) error {
	switch field {
	case "DomainFilters":
		// this is read as a string, the struct expects string slice
		// this is the reason the ExternalDNS needs to implement core.Settable
		e.DomainFilters = strings.Split(value.(string), ",")
	case "Policy":
		e.Policy = value.(core.OptionAnswer).Value
	case "Sources":
		answers, _ := value.([]core.OptionAnswer)
		sources := make([]string, 0)
		for _, ov := range answers {
			sources = append(sources, ov.Value)
		}
		e.Sources = sources
	case "TxtOwnerId":
		e.TxtOwnerId = value.(string)
	default:
	}

	return nil
}

func (p Provider) Validate() error {
	var validationErrors error

	if p.Name != dnsBanzaiCloud {
		if p.SecretID == "" {
			validationErrors = errors.Append(validationErrors, errors.Errorf("secret id must be specified for provider %s", p.Name))
		}
	}

	// todo validate specific options
	switch current := p.Name; current {
	case dnsBanzaiCloud:
	case dnsRoute53:
	case dnsGoogle:
	case dnsAzure:
	default:
		validationErrors = errors.Append(validationErrors, errors.Errorf("provider %s is not supported", current))
	}

	return validationErrors
}
