// Copyright © 2020 Banzai Cloud
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

package ingress

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/pkg/stringslice"
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
	return "Ingress"
}

func (Manager) ServiceName() string {
	return "ingress"
}

func (m Manager) BuildActivateRequestInteractively(clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {
	var request pipeline.ActivateIntegratedServiceRequest

	var spec spec
	if err := mapstructure.Decode(request.Spec, &spec); err != nil {
		return request, errors.WrapIf(err, "failed to decode spec in request")
	}

	if err := m.askControllerType(&spec); err != nil {
		return request, errors.WrapIf(err, "failed to ask for controller type")
	}

	if err := m.askControllerConfig(&spec); err != nil {
		return request, errors.WrapIf(err, "failed to ask for controller config")
	}

	if err := m.askIngressClass(&spec); err != nil {
		return request, errors.WrapIf(err, "failed to ask for ingress class")
	}

	if err := m.askServiceType(&spec); err != nil {
		return request, errors.WrapIf(err, "failed to ask for service type")
	}

	if err := m.askServiceAnnotations(&spec); err != nil {
		return request, errors.WrapIf(err, "failed to ask for service annotations")
	}

	if err := mapstructure.Decode(spec, &request.Spec); err != nil {
		return request, errors.WrapIf(err, "failed to decode typed spec to request spec")
	}

	return request, nil
}

func (m Manager) BuildUpdateRequestInteractively(clusterCtx clustercontext.Context, request *pipeline.UpdateIntegratedServiceRequest) error {
	var spec spec
	if err := mapstructure.Decode(request.Spec, &spec); err != nil {
		return errors.WrapIf(err, "failed to decode spec in request")
	}

	if err := m.askControllerType(&spec); err != nil {
		return errors.WrapIf(err, "failed to ask for controller type")
	}

	if err := m.askControllerConfig(&spec); err != nil {
		return errors.WrapIf(err, "failed to ask for controller config")
	}

	if err := m.askIngressClass(&spec); err != nil {
		return errors.WrapIf(err, "failed to ask for ingress class")
	}

	if err := m.askServiceType(&spec); err != nil {
		return errors.WrapIf(err, "failed to ask for service type")
	}

	if err := m.askServiceAnnotations(&spec); err != nil {
		return errors.WrapIf(err, "failed to ask for service annotations")
	}

	if err := mapstructure.Decode(spec, &request.Spec); err != nil {
		return errors.WrapIf(err, "failed to decode typed spec to request spec")
	}

	return nil
}

func (m Manager) ValidateSpec(rawSpec map[string]interface{}) error {
	var spec spec
	if err := mapstructure.Decode(rawSpec, &spec); err != nil {
		return errors.WrapIf(err, "service spec does not conform to schema")
	}

	return spec.Validate(m.banzaiCLI)
}

func (Manager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {
	result := map[string]map[string]interface{}{
		"Ingress": map[string]interface{}{
			"Status": details.Status,
		},
	}

	var o struct {
		Traefik struct {
			Version string `mapstructure:"version"`
		} `mapstructure:"traefik"`
	}

	if err := mapstructure.Decode(details.Output, &o); err != nil {
		result["Ingress"]["Error"] = "service output does not conform to schema"
		return result
	}

	var s spec
	if err := mapstructure.Decode(details.Spec, &s); err != nil {
		result["Ingress"]["Error"] = "service spec does not conform to schema"
		return result
	}

	switch s.Controller.Type {
	case ControllerTypeTraefik:
		traefikDetails := map[string]interface{}{
			"Service type": s.Service.Type,
			"Version":      o.Traefik.Version,
		}

		if s.IngressClass != "" {
			traefikDetails["Ingress class"] = s.IngressClass
		}

		result["Traefik"] = traefikDetails
	}

	return result
}

func (m Manager) askControllerType(s *spec) error {
	availableControllers, err := getAvailableControllerTypes(context.Background(), m.banzaiCLI)
	if err != nil {
		return errors.Wrap(err, "failed to get available controller types")
	}

	if !stringslice.Contains(availableControllers, s.Controller.Type) {
		s.Controller.Type = ""
	}

	var selectDefault interface{}
	if s.Controller.Type != "" {
		selectDefault = s.Controller.Type
	}

	if err := survey.AskOne(&survey.Select{
		Message: "Select the type of ingress controller to deploy:",
		Default: selectDefault,
		Options: availableControllers,
	}, &s.Controller.Type); err != nil {
		return errors.WrapIf(err, "failed to select ingress controller type")
	}

	return nil
}

func (m Manager) askControllerConfig(s *spec) error {
	switch s.Controller.Type {
	case ControllerTypeTraefik:
		var c traefikConfig
		_ = mapstructure.Decode(s.Controller.RawConfig, &c)

		if err := m.askTraefikConfig(&c); err != nil {
			return errors.WrapIf(err, "failed to ask for Traefik config")
		}

		if err := mapstructure.Decode(c, &s.Controller.RawConfig); err != nil {
			return errors.WrapIf(err, "failed to decode typed config to raw config")
		}
	}

	return nil
}

func (m Manager) askTraefikConfig(c *traefikConfig) error {
	var modify bool
	for {
		if err := showTraefikConfig(m.banzaiCLI.Out(), *c); err != nil {
			return errors.WrapIf(err, "failed to show Traefik config")
		}

		if err := survey.AskOne(&survey.Confirm{
			Message: "Would you like to modify current Traefik settings?",
			Default: false,
		}, &modify); err != nil {
			return errors.WrapIf(err, "failed to ask question")
		}

		if !modify {
			break
		}

		var setting string
		if err := survey.AskOne(&survey.Select{
			Message: "Which setting would you like to modify?",
			Options: []string{
				"SSL",
				"cancel",
			},
			Default: "cancel",
		}, &setting); err != nil {
			return errors.WrapIf(err, "failed to select setting to modify")
		}

		switch setting {
		case "SSL":
			{
				if err := survey.AskOne(&survey.Input{
					Message: "Default CN:",
					Default: c.SSL.DefaultCN,
				}, &c.SSL.DefaultCN, survey.WithValidator(validateCN)); err != nil {
					return errors.WrapIf(err, "failed to get value for default CN")
				}
			}

			{
				var defaultSANList string
				if err := survey.AskOne(&survey.Input{
					Message: "Default SAN DNS list:",
					Default: strings.Join(c.SSL.DefaultSANList, ", "),
				}, &defaultSANList, survey.WithValidator(validateSANList)); err != nil {
					return errors.WrapIf(err, "failed to get value for default SAN DNS list")
				}

				c.SSL.DefaultSANList = splitCommaSeparatedList(defaultSANList)
			}

			{
				var defaultIPList string
				if err := survey.AskOne(&survey.Input{
					Message: "Default SAN IP list:",
					Default: strings.Join(c.SSL.DefaultIPList, ", "),
				}, &defaultIPList, survey.WithValidator(validateIPList)); err != nil {
					return errors.WrapIf(err, "failed to get value for default SAN IP list")
				}

				c.SSL.DefaultIPList = splitCommaSeparatedList(defaultIPList)
			}
		}
	}

	return nil
}

const traefikConfigTemplateString = `Current Traefik settings:
# SSL:
  # Default CN: {{ if .SSL.DefaultCN }}{{ .SSL.DefaultCN }}{{ else }}(empty){{ end }}
  # Default SAN DNS List:{{ range .SSL.DefaultSANList }}
	> {{ . }}{{ else }} (empty){{ end }}
  # Default SAN IP List:{{ range .SSL.DefaultIPList }}
	> {{ . }}{{ else }} (empty){{ end }}
`

var traefikConfigTemplate = template.Must(template.New("traefikConfig").Parse(traefikConfigTemplateString))

func showTraefikConfig(w io.Writer, config traefikConfig) error {
	return traefikConfigTemplate.Execute(w, config)
}

func (Manager) askIngressClass(s *spec) error {
	if err := survey.AskOne(&survey.Input{
		Message: "Provide a custom ingress class or leave empty to use default:",
		Default: s.IngressClass,
	}, &s.IngressClass); err != nil {
		return errors.WrapIf(err, "failed to ask for ingress class")
	}
	return nil
}

func (Manager) askServiceType(s *spec) error {
	const defaultValue = "(default)"

	serviceType := s.Service.Type
	if serviceType == "" {
		serviceType = defaultValue
	}

	if err := survey.AskOne(&survey.Select{
		Message: "Select the service type used by the ingress controller:",
		Default: serviceType,
		Options: []string{
			"(default)",
			"ClusterIP",
			"LoadBalancer",
			"NodePort",
		},
	}, &serviceType); err != nil {
		return errors.WrapIf(err, "failed to ask for service type")
	}

	if serviceType == defaultValue {
		serviceType = ""
	}

	s.Service.Type = serviceType

	return nil
}

func (m Manager) askServiceAnnotations(s *spec) error {
	const (
		nothingOption = "Nothing"
		setOption     = "Set an annotation"
		removeOption  = "Remove an annotation"
	)

	var action string
Loop:
	for {
		if err := showServiceAnnotations(m.banzaiCLI.Out(), s.Service.Annotations); err != nil {
			return errors.WrapIf(err, "failed to show service annotations")
		}

		options := []string{
			nothingOption,
			setOption,
		}

		if len(s.Service.Annotations) != 0 {
			options = append(options, removeOption)
		}

		if err := survey.AskOne(&survey.Select{
			Message: "What would you like to do with the annotations?",
			Options: options,
			Default: nothingOption,
		}, &action); err != nil {
			return errors.WrapIf(err, "failed to ask for action")
		}

		switch action {
		case nothingOption:
			break Loop

		case setOption:
			var key string
			if err := survey.AskOne(&survey.Input{
				Message: "Provide the key of the annotation:",
			}, &key, survey.WithValidator(validateAnnotationKey)); err != nil {
				return errors.WrapIf(err, "failed to ask for annotation key")
			}

			val := s.Service.Annotations[key]
			if err := survey.AskOne(&survey.Input{
				Message: "Provide the value of the annotation:",
				Default: val,
			}, &val); err != nil {
				return errors.WrapIf(err, "failed to ask for annotation value")
			}

			if s.Service.Annotations == nil {
				s.Service.Annotations = make(map[string]string)
			}
			s.Service.Annotations[key] = val

		case removeOption:
			const cancel = "Cancel"

			options := make([]string, 0, len(s.Service.Annotations)+1)
			optionsMap := make([]string, 0, len(s.Service.Annotations))
			for key, val := range s.Service.Annotations {
				options = append(options, fmt.Sprintf("Remove %s: %s", key, val))
				optionsMap = append(optionsMap, key)
			}
			options = append(options, cancel)

			var idx int
			if err := survey.AskOne(&survey.Select{
				Message: "Which annotation would you like to remove?",
				Options: options,
				Default: cancel,
			}, &idx); err != nil {
				return errors.WrapIf(err, "failed to ask for annotation to remove")
			}

			if idx < len(optionsMap) {
				delete(s.Service.Annotations, optionsMap[idx])
			}
		}
	}
	return nil
}

const serviceAnnotationsTemplateString = `Current service annotations:{{range $key, $value := .}}
> {{ $key }}: {{ $value }}{{ else }} (empty){{ end }}
`

var serviceAnnotationsTemplate = template.Must(template.New("serviceAnnotations").Parse(serviceAnnotationsTemplateString))

func showServiceAnnotations(w io.Writer, annotations map[string]string) error {
	return serviceAnnotationsTemplate.Execute(w, annotations)
}

func validateCN(ans interface{}) error {
	cn, ok := ans.(string)
	if !ok {
		return errors.Errorf("%#v is not a string", ans)
	}

	if cn == "" {
		return nil
	}

	return validateDomainOrIP(cn)
}

func validateSANList(ans interface{}) error {
	listStr, ok := ans.(string)
	if !ok {
		return errors.Errorf("%#v is not a string", ans)
	}

	var errs error
	for _, e := range splitCommaSeparatedList(listStr) {
		errs = errors.Append(errs, validateDomain(e))
	}

	return errs
}

func validateIPList(ans interface{}) error {
	listStr, ok := ans.(string)
	if !ok {
		return errors.Errorf("%#v is not a string", ans)
	}

	var errs error
	for _, e := range splitCommaSeparatedList(listStr) {
		errs = errors.Append(errs, validateIP(e))
	}

	return errs
}

func validateAnnotationKey(ans interface{}) error {
	return nil
}
