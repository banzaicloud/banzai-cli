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

package input

import (
	"context"

	"github.com/AlecAivazis/survey/v2"
	"github.com/goph/emperror"
	"github.com/pkg/errors"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

// AskResourceGroup asks for Azure resource group available with the given secret ID
func AskResourceGroup(banzaiCli cli.Cli, orgID int32, secretID, defaultResourceGroup string) (string, error) {
	var resourceGroup string

	rgs, _, err := banzaiCli.Client().InfoApi.GetResourceGroups(context.Background(), orgID, secretID)
	if err != nil {
		return "", emperror.Wrap(utils.ConvertError(err), "can't list resource groups")
	}

	err = survey.AskOne(&survey.Select{Message: "Resource group:", Options: rgs, Default: defaultResourceGroup}, &resourceGroup)
	if err != nil {
		return "", emperror.Wrap(err, "no resource group selected")
	}

	return resourceGroup, nil
}

// IsResourceGroupValid checks whether the given resource group name is valid
func IsResourceGroupValid(banzaiCli cli.Cli, orgID int32, secretID string, resourceGroup string) error {
	rgs, _, err := banzaiCli.Client().InfoApi.GetResourceGroups(context.Background(), orgID, secretID)
	if err != nil {
		return emperror.Wrap(utils.ConvertError(err), "could not list resource groups")
	}

	for _, rg := range rgs {
		if rg == resourceGroup {
			return nil
		}
	}

	return errors.Errorf("invalid resource group: %s", resourceGroup)
}
