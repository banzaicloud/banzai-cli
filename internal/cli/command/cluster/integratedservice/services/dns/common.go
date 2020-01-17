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

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

const (
	serviceName = "dns"

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
	sources  = []string{sourceIngress, sourceService}
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
	return serviceName
}

func (baseManager) CheckCapabilities(cap map[string]interface{}) {

}

func NewDeactivateManager() *baseManager {
	return &baseManager{}
}

func NewCapabilitiesManager() *baseManager {
	return &baseManager{}
}

func validateSpec(specObj map[string]interface{}) error {
	var dnsSpec ServiceSpec

	if err := mapstructure.Decode(specObj, &dnsSpec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	err := dnsSpec.ExternalDNS.Validate()

	if dnsSpec.ClusterDomain == "" {
		err = errors.Append(err, errors.New("cluster domain must not be empty"))
	}

	return err
}

const (
	actionNew    = "newAction"
	actionUpdate = "updateAction"
)

// holds values related to the current operation (create, update)s
type actionContext struct {
	action       string
	providerName string
}

func NewActionContext(action string) actionContext {
	return actionContext{
		action: action,
	}
}

func (ac *actionContext) SetProvider(providerName string) {
	ac.providerName = providerName
}

func (ac actionContext) IsUpdate() bool {
	return ac.action == actionUpdate
}

// assembleServiceRequest assembles the request for activate and update the ExternalDNS integrated service
// if the input rawSpec is nil -> activate flow, otherwise update flow
func assembleServiceRequest(banzaiCli cli.Cli, clusterCtx clustercontext.Context, spec ServiceSpec, actionContext actionContext) (map[string]interface{}, error) {

	// select the provider
	selectedProviderInfo, err := selectProvider(*spec.ExternalDNS.Provider)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider data")
	}

	actionContext.SetProvider(selectedProviderInfo.Name)
	spec, err = getServiceSpecDefaults(banzaiCli, clusterCtx, spec, actionContext)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get dns service defaults")
	}
	spec.ExternalDNS.Provider = &selectedProviderInfo

	// read secret
	selectedProviderInfo, err = decorateProviderSecret(banzaiCli, selectedProviderInfo)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider secret")
	}

	// read options
	selectedProviderInfo, err = decorateProviderOptions(banzaiCli, selectedProviderInfo)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read provider options")
	}

	spec.ExternalDNS.Provider = &selectedProviderInfo

	spec.ExternalDNS, err = readExternalDNS(spec.ExternalDNS, actionContext)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read external dns data")
	}

	if selectedProviderInfo.Name != dnsBanzaiCloud {
		// in case of Banzai Cloud  DNS this value gets generated / it's read only
		clusterDomain, err := readClusterDomain(spec.ClusterDomain)
		if err != nil {
			return nil, errors.WrapIf(err, "failed to read cluster domain")
		}
		spec.ClusterDomain = clusterDomain
	}

	var jsonSpec map[string]interface{}
	if err := mapstructure.Decode(spec, &jsonSpec); err != nil {
		return nil, errors.WrapIf(err, "failed to assemble DNSServiceSpec")
	}

	return jsonSpec, nil
}
