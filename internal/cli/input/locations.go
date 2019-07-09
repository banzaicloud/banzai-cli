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
	"sort"

	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

// AskLocation asks for cloud provider location for the given provider
func AskLocation(banzaiCli cli.Cli, cloud string) (string, error) {
	var locationName string

	regions, _, err := banzaiCli.CloudinfoClient().RegionsApi.GetRegions(context.Background(), cloud, "compute")
	if err != nil {
		return "", utils.ConvertError(err)
	}

	locationOptions := make([]string, len(regions))
	locationIDs := make(map[string]string, len(regions))
	for i, region := range regions {
		locationOptions[i] = region.Name
		locationIDs[region.Name] = region.Id
	}
	sort.Strings(locationOptions)
	err = survey.AskOne(&survey.Select{Message: "Location:", Options: locationOptions}, &locationName, survey.Required)
	if err != nil {
		return "", emperror.Wrap(err, "failed to select location")
	}

	return locationIDs[locationName], nil
}

// IsLocationValid checks whether the given location is valid the the specified cloud provider
func IsLocationValid(banzaiCli cli.Cli, cloud, location string) error {
	regions, _, err := banzaiCli.CloudinfoClient().RegionsApi.GetRegions(context.Background(), cloud, "compute")
	if err != nil {
		return utils.ConvertError(err)
	}

	for _, region := range regions {
		if region.Id == location {
			return nil
		}
	}

	return errors.Errorf("invalid location: %s", location)
}
