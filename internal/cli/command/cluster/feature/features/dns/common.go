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
	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"
)

const (
	featureName = "dns"

	dnsRoute53     = "route53"
	dnsAzure       = "azure"
	dnsGoogle      = "google"
	dnsBanzaiCloud = "banzaicloud-dns"

	sourceIngress = "ingress"
	sourceService = "service"

	policyUpsertOnly = "upsert-only"
	policySync       = "sync"
)

var (
	sources = []string{sourceIngress, sourceService}
	policies = []string{policySync, policyUpsertOnly}

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

func (baseManager) GetName() string {
	return featureName
}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

func validateSpec(specObj map[string]interface{}) error {
	var dnsSpec DNSFeatureSpec

	if err := mapstructure.Decode(specObj, &dnsSpec); err != nil {
		return errors.WrapIf(err, "feature specification does not conform to schema")
	}

	err := dnsSpec.ExternalDNS.Validate()

	if dnsSpec.ClusterDomain == "" {
		err = errors.Append(err, errors.New("cluster domain must not be empty"))
	}

	return err
}
