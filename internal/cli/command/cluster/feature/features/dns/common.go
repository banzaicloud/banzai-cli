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
	"github.com/mitchellh/mapstructure"
)

const (
	featureName = "dns"

	dnsRoute53     = "route53"
	dnsAzure       = "azure"
	dnsGoogle      = "google"
	dnsBanzaiCloud = "banzaicloud-dns"
)

var (
	providerMeta = map[string]struct {
		Name       string
		SecretType string
	}{
		dnsBanzaiCloud: {
			Name:       "Banzai Cloud DNS",
			SecretType: "amazon",
		},
		dnsRoute53: {
			Name:       "Amazon Route 53",
			SecretType: "amazon",
		},
		dnsAzure: {
			Name:       "Azure DNS",
			SecretType: "azure",
		},
		dnsGoogle: {
			Name:       "Google Cloud DNS",
			SecretType: "google",
		},
	}
)

type baseManager struct{}

type ExternalDNS struct {
	DomainFilters []string  `mapstructure:"domainFilters"`
	Policy        string    `mapstructure:"policy"` // sync | upsert-only
	Sources       []string  `mapstructure:"sources"`
	TxtOwnerId    string    `mapstructure:"txtOwnerId"`
	Provider      *Provider `mapstructure:"provider"`
}

type Provider struct {
	Name     string                 `mapstructure:"name"`
	SecretID string                 `mapstructure:"secretId"`
	Options  map[string]interface{} `mapstructure:"options"`
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

	if e.TxtOwnerId == "" {
		validationErrors = errors.Append(validationErrors, errors.New("txtOwnerId must not be empty"))
	}

	if e.Provider == nil {
		validationErrors = errors.Append(validationErrors, errors.New("provider must be specified"))
	} else {
		validationErrors = errors.Append(validationErrors, e.Provider.Validate())
	}

	return validationErrors
}

func (e *ExternalDNS) WriteAnswer(field string, value interface{}) error {
	// todo use reflection here
	switch field {
	case "DomainFilters":
		e.DomainFilters = strings.Split(value.(string), ",")
	case "Policy ":
		e.Policy = value.(string)
	case "Sources":
		answers, _ := value.([]core.OptionAnswer)
		srcs := make([]string, 0)
		for _, ov := range answers {
			srcs = append(srcs, ov.Value)
		}
		e.Sources = srcs
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

type spec struct {
	ExternalDNS   ExternalDNS `mapstructure:"externalDns"`
	ClusterDomain string      `mapstructure:"clusterDomain"`
}

func (baseManager) GetName() string {
	return featureName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

func validateSpec(specObj map[string]interface{}) error {
	var dnsSpec spec

	if err := mapstructure.Decode(specObj, &dnsSpec); err != nil {
		return errors.WrapIf(err, "feature specification does not conform to schema")
	}

	err := dnsSpec.ExternalDNS.Validate()

	if dnsSpec.ClusterDomain == "" {
		err = errors.Append(err, errors.New("cluster domain must not be empty"))
	}

	return err
}

type specResponse struct {
}

// helper type alias for id -> name maps
type idNameMap = map[string]string

func Names(sm idNameMap) []string {
	names := make([]string, len(sm))
	for _, name := range sm {
		names = append(names, name)
	}
	return names
}

func NameForID(sm idNameMap, idOf string) string {
	for id, n := range sm {
		if id == idOf {
			return n
		}
	}
	return ""
}

func nameToIDTransformer(sm idNameMap) func(name interface{}) interface{} {
	return func(name interface{}) interface{} {
		for id, n := range sm {
			if n == name.(core.OptionAnswer).Value {
				return core.OptionAnswer{
					Value: id,
				}
			}
		}
		return nil
	}
}
