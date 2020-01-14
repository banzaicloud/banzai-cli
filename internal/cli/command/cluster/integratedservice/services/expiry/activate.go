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

package expiry

import (
	"encoding/json"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
)

type ActivateManager struct {
	baseManager
}

func (ActivateManager) BuildRequestInteractively(_ cli.Cli, _ clustercontext.Context) (*pipeline.ActivateClusterFeatureRequest, error) {
	date, err := askForDate("")
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get date")
	}

	return &pipeline.ActivateClusterFeatureRequest{
		Spec: map[string]interface{}{
			"date": date,
		},
	}, nil
}

func (ActivateManager) ValidateRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}

func NewActivateManager() *ActivateManager {
	return &ActivateManager{}
}

func askForDate(defaultValue string) (string, error) {
	var date string

	for {
		for {
			if err := input.DoQuestions([]input.QuestionMaker{
				input.QuestionInput{
					QuestionBase: input.QuestionBase{
						Message: "Provide expiration date in UTC:",
						Help:    fmt.Sprintf("Date format should be: %s", time.Now().UTC().Format(time.RFC3339)),
					},
					DefaultValue: defaultValue,
					Output:       &date,
				}}); err != nil {
				return "", errors.WrapIf(err, "error during getting secret")
			}

			if err := validateDate(date); err != nil {
				log.Error("error during validation date: ", err.Error())
			} else {
				break
			}
		}

		// confirm date
		var isConfirmed bool
		if err := input.DoQuestions([]input.QuestionMaker{
			input.QuestionConfirm{
				QuestionBase: input.QuestionBase{
					Message: fmt.Sprintf("Are you sure you want this cluster to be deleted at %s?", date),
				},
				DefaultValue: true,
				Output:       &isConfirmed,
			},
		}); err != nil {
			return "", errors.WrapIf(err, "error during date confirm")
		}

		if isConfirmed {
			break
		}
	}

	return date, nil
}
