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
	"context"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
	cliutils "github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type Manager struct {
	banzaiCLI cli.Cli
}

func NewManager(banzaiCLI cli.Cli) Manager {
	return Manager{
		banzaiCLI: banzaiCLI,
	}
}

func (Manager) ReadableName() string {
	return "Security scan"
}

func (Manager) ServiceName() string {
	return "securityscan"
}

func (m Manager) BuildActivateRequestInteractively(clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {
	if err := isServiceEnabled(context.Background(), m.banzaiCLI); err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "securityscan is not enabled")
	}

	serviceSpec, err := assembleServiceSpec(context.Background(), m.banzaiCLI, m.banzaiCLI.Context().OrganizationID(), clusterCtx.ClusterID(), ServiceSpec{})
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to assemble integratedservice specification")
	}

	serviceSpecMap, err := securityScanSpecAsMap(&serviceSpec)
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to transform integratedservice specification")
	}

	return pipeline.ActivateIntegratedServiceRequest{Spec: serviceSpecMap}, nil
}

func (m Manager) BuildUpdateRequestInteractively(clusterCtx clustercontext.Context, request *pipeline.UpdateIntegratedServiceRequest) error {
	if err := isServiceEnabled(context.Background(), m.banzaiCLI); err != nil {
		return errors.WrapIf(err, "securityscan is not enabled")
	}

	serviceSpec := ServiceSpec{}
	if err := mapstructure.Decode(request.Spec, &serviceSpec); err != nil {
		return errors.WrapIf(err, "failed to decode service specification for update")
	}

	serviceSpec, err := assembleServiceSpec(context.Background(), m.banzaiCLI, m.banzaiCLI.Context().OrganizationID(), clusterCtx.ClusterID(), serviceSpec)
	if err != nil {
		return errors.WrapIf(err, "failed to assemble service specification")
	}

	serviceSpecMap, err := securityScanSpecAsMap(&serviceSpec)
	if err != nil {
		return errors.WrapIf(err, "failed to transform service specification")
	}

	request.Spec = serviceSpecMap

	return nil
}

func (Manager) ValidateSpec(spec map[string]interface{}) error {
	return nil
}

func (Manager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {
	tableData := map[string]interface{}{
		"Status": details.Status,
	}

	if anchore, ok := getObj(details.Output, "anchore"); ok {
		if aVersion, ok := getStr(anchore, "version"); ok {
			tableData["AnchoreVersion"] = aVersion
		}
	}

	if imgValidator, ok := getObj(details.Output, "imageValidator"); ok {
		if aVersion, ok := getStr(imgValidator, "version"); ok {
			tableData["ImageValidatorVersion"] = aVersion
		}
	}

	return map[string]map[string]interface{}{
		"Security_scan": tableData,
	}
}

func getList(target map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := target[key]; ok {
		if list, ok := value.([]interface{}); ok {
			return list, true
		}
	}
	return nil, false
}

func getObj(target map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := target[key]; ok {
		if obj, ok := value.(map[string]interface{}); ok {
			return obj, true
		}
	}
	return nil, false
}

func getStr(target map[string]interface{}, key string) (string, bool) {
	if value, ok := target[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}
func readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateIntegratedServiceRequest) error {
	filename, raw, err := cliutils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func securityScanSpecAsMap(spec *ServiceSpec) (map[string]interface{}, error) {
	// fill the structure of the config - make filling up the values easier
	if spec == nil {
		spec = &ServiceSpec{
			CustomAnchore:    anchoreSpec{},
			Policy:           policySpec{},
			ReleaseWhiteList: nil,
			WebhookConfig:    webHookConfigSpec{},
		}
	}

	var specMap map[string]interface{}
	if err := mapstructure.Decode(spec, &specMap); err != nil {
		return nil, err
	}

	return specMap, nil
}

func isServiceEnabled(ctx context.Context, banzaiCLI cli.Cli) error {
	capabilities, r, err := banzaiCLI.Client().PipelineApi.ListCapabilities(ctx)
	if err := utils.CheckCallResults(r, err); err != nil {
		return errors.WrapIf(err, "failed to retrieve capabilities")
	}

	rawSecurityscanCapability, ok := capabilities["features"]["securityScan"]
	if !ok {
		return errors.New("no securityscan capabilities found")
	}

	var securityScanCapability = struct {
		Enabled bool `json:"enabled" mapstructure:"enabled"`
		Managed bool `json:"managed" mapstructure:"managed"`
	}{}

	if err := mapstructure.Decode(rawSecurityscanCapability, &securityScanCapability); err != nil {
		return errors.WrapIf(err, "failed to parse securityscan capabilities")
	}

	// todo change this implementation when adding support for non-managed anchore
	if !securityScanCapability.Enabled || !securityScanCapability.Managed {
		return errors.New("security scan is not enabled")
	}

	return nil
}

func askForAnchoreConfig(banzaiCLI cli.Cli, currentAnchoreSpec *anchoreSpec) (*anchoreSpec, error) {
	if currentAnchoreSpec == nil {
		currentAnchoreSpec = &anchoreSpec{}
	}

	var customAnchore bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Configure a custom anchore instance?",
		},
		&customAnchore,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	if !customAnchore {
		return &anchoreSpec{
			Enabled: false,
		}, nil
	}

	// custom anchore config
	var anchoreURL string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the custom anchore URL:",
			Default: currentAnchoreSpec.Url,
		},
		&anchoreURL,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read custom Anchore URL")
	}

	secretID, err := askForSecret(banzaiCLI, currentAnchoreSpec.SecretID, "Please select a secret to access the custom Anchore instance:", PasswordSecretType)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read secret for accessing custom Anchore")
	}

	return &anchoreSpec{
		Enabled:  true,
		Url:      anchoreURL,
		SecretID: secretID,
	}, nil
}

const (
	PasswordSecretType = "password"
	AmazonSecretType   = "amazon"
)

func askForSecret(banzaiCLI cli.Cli, currentSecretID string, message string, types ...string) (string, error) {
	orgID := banzaiCLI.Context().OrganizationID()

	var secrets []pipeline.SecretItem // nolint
	for _, secretType := range types {
		s, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{Type_: optional.NewString(secretType)})
		if err != nil {
			return "", errors.Wrap(err, "failed to retrieve secrets")
		}

		secrets = append(secrets, s...)
	}

	// TODO add create secret option
	if len(secrets) == 0 {
		return "", errors.New(fmt.Sprintf("there are no secrets with type %q", PasswordSecretType))
	}

	options := make([]string, len(secrets))
	currentSecretName := secrets[0].Name // default!
	for i, s := range secrets {
		options[i] = s.Name
		if s.Id == currentSecretID {
			currentSecretName = s.Name
		}
	}

	var secretName string
	if err := survey.AskOne(
		&survey.Select{
			Message: message,
			Options: options,
			Default: currentSecretName,
		},
		&secretName,
	); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve secrets")
	}

	for _, s := range secrets {
		if s.Name == secretName {
			return s.Id, nil
		}
	}

	return "", errors.Errorf("no secret with name %q", secretName)
}

func getNamespaces(ctx context.Context, banzaiCLI cli.Cli, orgID int32, clusterID int32) ([]string, error) {
	nsResponse, response, err := banzaiCLI.Client().ClustersApi.ListNamespaces(ctx, orgID, clusterID)
	if err := utils.CheckCallResults(response, err); err != nil {
		return nil, errors.WrapIf(err, "failed to retrieve namespaces")
	}

	// filter out system namespaces
	filtered := make([]string, 0, len(nsResponse.Namespaces))
	for _, ns := range nsResponse.Namespaces {
		if ns.Name == "kube-system" || ns.Name == "pipeline-system" {
			continue
		}

		filtered = append(filtered, ns.Name)
	}

	return filtered, nil
}

// policies are statically stored, the selection is made from a "wired" list
func askForPolicy(policySpecIn policySpec) (policySpec, error) {
	defaultPolicyBundle := utils.NameForID(policyBundles, policySpecIn.PolicyID)
	if defaultPolicyBundle == "" {
		defaultPolicyBundle = "Default bundle"
	}

	qs := []*survey.Question{
		{
			Name: "PolicyID",
			Prompt: &survey.Select{
				Message: "Please select the policy bundle",
				Options: utils.Names(policyBundles),
				Default: defaultPolicyBundle,
			},
			Transform: utils.NameToIDTransformer(policyBundles),
		},
	}

	if err := survey.Ask(qs, &policySpecIn); err != nil {
		return policySpec{}, errors.WrapIf(err, "failed to read policy")
	}

	if policySpecIn.PolicyID == "Custom" {
		var customPolicy string
		if err := survey.AskOne(&survey.Input{
			Message: "Please provide an Anchore policy document",
		}, &customPolicy); err != nil {
			return policySpec{}, errors.WrapIf(err, "failed to read custom policy")
		}

		err := json.Unmarshal([]byte(customPolicy), &policySpecIn.CustomPolicy.Policy)
		if err != nil {
			return policySpec{}, errors.WrapIf(err, "failed to parse custom policy")
		}

		policySpecIn.CustomPolicy.Enabled = true
		policySpecIn.PolicyID = ""
	}

	return policySpecIn, nil
}

func askForWebHookConfig(ctx context.Context, banzaiCLI cli.Cli, orgID int32, clusterID int32, webhookSpecIn webHookConfigSpec) (webHookConfigSpec, error) {
	qs := []*survey.Question{
		{
			Name: "Enabled",
			Prompt: &survey.Confirm{
				Message: "Enable the security scan webhook?",
				Default: webhookSpecIn.Enabled,
			},
		},
	}

	if err := survey.Ask(qs, &webhookSpecIn); err != nil {
		return webHookConfigSpec{}, errors.WrapIf(err, "failed to read webhook configuration value")
	}

	if !webhookSpecIn.Enabled {
		return webHookConfigSpec{
			Enabled: false,
		}, nil
	}

	// set the default selector
	if webhookSpecIn.Selector == "" {
		webhookSpecIn.Selector = "include"
	}

	// ignore the error
	namespaceOptions, _ := getNamespaces(ctx, banzaiCLI, orgID, clusterID)
	if len(namespaceOptions) == 0 {
		// couldn't retrieve namespaces / namespaces are not available
		namespaceOptions = append(namespaceOptions, "default")
	}

	// append the allStar
	namespaceOptions = append([]string{"*"}, namespaceOptions...)

	defaultNamespaces := webhookSpecIn.Namespaces
	// empty the namespaces field
	webhookSpecIn.Namespaces = make([]string, 0, len(defaultNamespaces))

	if len(defaultNamespaces) == 0 {
		defaultNamespaces = []string{"*"}
	}

	// questions to fill the remaining parts of the webhook configuration
	qs = []*survey.Question{
		{
			Name: "Selector",
			Prompt: &survey.Select{
				Message: "Choose the selector for namespaces:",
				Options: []string{"include", "exclude"},
				Default: webhookSpecIn.Selector,
				Help:    "The selector defines whether the selected namespaces are included or excluded from security scans",
			},
		},
		{
			Name: "Namespaces",
			Prompt: &survey.MultiSelect{
				Message: "Select the namespaces the selector applies to:",
				Options: namespaceOptions,
				Default: defaultNamespaces,
				Help:    "Selected namespaces will be included or excluded form security scans",
			},
			Validate: func(selection interface{}) error {
				selected := selection.([]core.OptionAnswer)
				for _, ns := range selected {
					if ns.Value == "*" && len(selected) > 1 {
						return errors.New("all namespaces (*) is selected; please deselect it or deselect the other namespaces")
					}
				}
				return nil
			},
		},
	}

	if err := survey.Ask(qs, &webhookSpecIn); err != nil {
		return webHookConfigSpec{}, errors.WrapIf(err, "failed to read webhook configuration value")
	}

	return webhookSpecIn, nil
}

func askForWhiteLists() ([]releaseSpec, error) {
	addMore := true
	releaseWhiteList := make([]releaseSpec, 0)

	for addMore {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Add a release whitelist item to the security scan? ",
			},
			&addMore,
		); err != nil {
			return nil, errors.WrapIf(err, "failure during survey")
		}

		if !addMore {
			continue
		}

		item, err := askForWhiteListItem()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to read release whitelist item")
		}
		releaseWhiteList = append(releaseWhiteList, *item)
	}

	return releaseWhiteList, nil
}

func askForWhiteListItem() (*releaseSpec, error) {
	var releaseName string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the name of the release whitelist item:",
		},
		&releaseName,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the name of the release whitelist item")
	}

	var reason string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the reason of the release whitelist item:",
		},
		&reason,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the reason of the release whitelist item")
	}

	var regexp string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the regexp for the release whitelist item:",
		},
		&regexp,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read the regexp of the release whitelist item")
	}

	return &releaseSpec{
		Name:   releaseName,
		Reason: reason,
		Regexp: regexp,
	}, nil
}

func askForCustomRegistry(banzaiCLI cli.Cli) (*registrySpec, error) {
	var registry string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Please enter the custom container registry:",
			Default: registry,
		},
		&registry,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to read custom container registry")
	}

	var registryType string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Registry type",
			Options: []string{"docker_v2", "awsecr"},
			Default: "docker_v2",
		},
		&registryType,
	); err != nil {
		return nil, errors.WrapIf(err, "failed to choose registry type")
	}

	var insecure bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Access the registry without TLS verification?",
		},
		&insecure,
	); err != nil {
		return nil, errors.WrapIf(err, "failure during survey")
	}

	secretID, err := askForSecret(banzaiCLI, "", "Please select a secret to access the custom Anchore registry:", PasswordSecretType, AmazonSecretType)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to read secret for accessing custom Anchore registry")
	}

	return &registrySpec{
		Registry: registry,
		SecretID: secretID,
		Type:     registryType,
		Insecure: insecure,
	}, nil
}

func assembleServiceSpec(ctx context.Context, banzaiCLI cli.Cli, orgID int32, clusterID int32, serviceSpecIn ServiceSpec) (ServiceSpec, error) {
	anchoreConfig, err := askForAnchoreConfig(banzaiCLI, &serviceSpecIn.CustomAnchore)
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assemble anchore data")
	}

	policy, err := askForPolicy(serviceSpecIn.Policy)
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assemble policy data")
	}

	webhookConfig, err := askForWebHookConfig(ctx, banzaiCLI, orgID, clusterID, serviceSpecIn.WebhookConfig)
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assemble webhook data")
	}

	releaseWhiteList, err := askForWhiteLists()
	if err != nil {
		return ServiceSpec{}, errors.WrapIf(err, "failed to assemble release data")
	}

	var customRegistries []*registrySpec

	var addCustomRegistry bool

	for {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Add a custom registry in anchore?",
			},
			&addCustomRegistry,
		); err != nil {
			return ServiceSpec{}, errors.WrapIf(err, "failure during survey")
		}

		if !addCustomRegistry {
			break
		}
		customRegistry, err := askForCustomRegistry(banzaiCLI)
		if err != nil {
			return ServiceSpec{}, errors.WrapIf(err, "failed to assemble registry data")
		}
		customRegistries = append(customRegistries, customRegistry)
	}

	return ServiceSpec{
		CustomAnchore:    *anchoreConfig,
		Policy:           policy,
		WebhookConfig:    webhookConfig,
		ReleaseWhiteList: releaseWhiteList,
		Registries:       customRegistries,
	}, nil
}
