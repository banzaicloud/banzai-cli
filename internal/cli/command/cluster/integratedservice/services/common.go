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

package services

import (
	"fmt"

	"emperror.dev/errors"
)

const (
	serviceKeyOnCap = "features"
	enabledKeyOnCap = "enabled"
)

type Cap map[string]map[string]interface{}

func (capabilities Cap) isServiceEnabled(serviceName string) error {
	//capabilities, r, err := banzaiCLI.Client().PipelineApi.ListCapabilities(ctx)
	//if err := utils.CheckCallResults(r, err); err != nil {
	//	return errors.WrapIf(err, "failed to retrieve capabilities")
	//}

	if services, ok := capabilities[serviceKeyOnCap]; ok {
		if s, ok := services[serviceName]; ok {
			if svc, ok := s.(map[string]interface{}); ok {
				if en, ok := svc[enabledKeyOnCap]; ok {
					if enabled, ok := en.(bool); ok {
						if enabled {
							return nil
						}
					}
				}
			}
		}
	}

	return errors.New(fmt.Sprintf("%s service disabled", serviceName))
}
