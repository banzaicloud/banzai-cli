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
	"net"
	"regexp"

	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/pkg/stringslice"
	"github.com/mitchellh/mapstructure"
)

type spec struct {
	Controller struct {
		Type      string                 `mapstructure:"type"`
		RawConfig map[string]interface{} `mapstructure:"config"`
	} `mapstructure:"controller"`
	IngressClass string `mapstructure:"ingressClass"`
	Service      struct {
		Type        string            `mapstructure:"type"`
		Annotations map[string]string `mapstructure:"annotations"`
	} `mapstructure:"service"`
}

func (s spec) Validate(banzaiCLI cli.Cli) error {
	var errs error

	if s.Controller.Type == "" {
		errs = errors.Append(errs, errors.New("controller type must be specified"))
	}

	availableControllers, err := getAvailableControllerTypes(context.Background(), banzaiCLI)
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to get available controller types"))
	}

	if !stringslice.Contains(availableControllers, s.Controller.Type) {
		errs = errors.Append(errs, errors.Errorf("controller type %q is not available", s.Controller.Type))
	}

	switch s.Controller.Type {
	case ControllerTypeTraefik:
		var c traefikConfig
		if err := mapstructure.Decode(s.Controller.RawConfig, &c); err != nil {
			errs = errors.Append(errs, errors.WrapIf(err, "failed to decode controller config as Traefik config"))
		} else {
			errs = errors.Append(errs, errors.WrapIf(c.Validate(), "Traefik config validation failed"))
		}
	}

	switch s.Service.Type {
	case "", "ClusterIP", "LoadBalancer", "NodePort":
		// ok
	default:
		errs = errors.Append(errs, errors.Errorf("%q is not a valid service type", s.Service.Type))
	}

	return errs
}

type traefikConfig struct {
	SSL struct {
		DefaultCN      string   `mapstructure:"defaultCN"`
		DefaultSANList []string `mapstructure:"defaultSANList"`
		DefaultIPList  []string `mapstructure:"defaultIPList"`
	} `mapstructure:"ssl"`
}

func (c traefikConfig) Validate() error {
	var errs error
	if c.SSL.DefaultCN != "" {
		errs = errors.Append(errs, validateDomainOrIP(c.SSL.DefaultCN))
	}
	if len(c.SSL.DefaultSANList) != 0 {
		for _, e := range c.SSL.DefaultSANList {
			errs = errors.Append(errs, validateDomain(e))
		}
	}
	if len(c.SSL.DefaultIPList) != 0 {
		for _, e := range c.SSL.DefaultIPList {
			errs = errors.Append(errs, validateIP(e))
		}
	}
	return errs
}

var domainRegexp = regexp.MustCompile(`^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9]+$`)

func validateDomain(domain string) error {
	if !domainRegexp.MatchString(domain) {
		return errors.Errorf("%q is not a valid domain name", domain)
	}
	return nil
}

func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return errors.Errorf("%q is not a valid IP address", ip)
	}
	return nil
}

func validateDomainOrIP(domainOrIP string) error {
	if validateDomain(domainOrIP) == nil {
		return nil
	}

	if validateIP(domainOrIP) == nil {
		return nil
	}

	return errors.Errorf("%q is not a valid domain name or IP address", domainOrIP)
}
