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

	"github.com/antihax/optional"
	"github.com/goph/emperror"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

func AskSecret(banzaiCli cli.Cli, orgID int32, cloud string) (string, error) {
	var secretName string

	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{
		Type_: optional.NewString(cloud),
	})
	if err != nil {
		return "", emperror.Wrap(utils.ConvertError(err), "could not get secret")
	}

	secretOptions := make([]string, len(secrets))
	secretIds := make(map[string]string, len(secrets))
	for i, s := range secrets {
		secretOptions[i] = s.Name
		secretIds[s.Name] = s.Id
	}
	err = survey.AskOne(&survey.Select{Message: "Secret:", Options: secretOptions}, &secretName, survey.Required)
	if err != nil {
		return "", emperror.Wrap(err, "failed to select secret")
	}

	return secretIds[secretName], nil
}
