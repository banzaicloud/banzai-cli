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

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)


// AskDeployment prompts user to select a deployment
func AskDeployment(banzaiCli cli.Cli, orgID, clusterID int32) (string, error) {
	deployments, _, err := banzaiCli.Client().DeploymentsApi.ListDeployments(context.Background(), orgID, clusterID, nil)
	if err != nil {
		return "", emperror.Wrap(err, "could not list deployments")
	}

	if len(deployments) == 0 {
		return "", errors.New("no deployments found in the cluster")
	}

	deploymentSurveyInput := make([]string, len(deployments))
	for i, deployment := range deployments {
		deploymentSurveyInput[i] = deployment.ReleaseName
	}

	releaseName := ""
	err = survey.AskOne(&survey.Select{Message: "Release name", Options: deploymentSurveyInput}, &releaseName, survey.Required)
	if err != nil {
		return "", emperror.Wrap(err, "error occurred while selecting deployment")
	}

	return releaseName, nil
}
