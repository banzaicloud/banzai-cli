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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

const serviceName = "expiry"

type Manager struct{}

func (Manager) GetName() string {
	return serviceName
}

func (Manager) BuildActivateRequestInteractively(_ cli.Cli, _ clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error) {
	date, err := askForDate("")
	if err != nil {
		return pipeline.ActivateIntegratedServiceRequest{}, errors.WrapIf(err, "failed to get date")
	}

	return pipeline.ActivateIntegratedServiceRequest{
		Spec: map[string]interface{}{
			"date": date,
		},
	}, nil
}

func (Manager) BuildUpdateRequestInteractively(_ cli.Cli, req *pipeline.UpdateIntegratedServiceRequest, _ clustercontext.Context) error {
	var spec serviceSpec
	if err := mapstructure.Decode(req.Spec, &spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	date, err := askForDate(spec.Date)
	if err != nil {
		return errors.WrapIf(err, "error during getting date")
	}

	spec.Date = date
	if err := mapstructure.Decode(spec, &req.Spec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	return nil
}

func (Manager) ValidateSpec(spec map[string]interface{}) error {
	var typedSpec serviceSpec

	if err := mapstructure.Decode(spec, &typedSpec); err != nil {
		return errors.WrapIf(err, "service specification does not conform to schema")
	}

	return typedSpec.Validate()
}

func (Manager) WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{} {

	const (
		expiryTitle = "Expiry"
		statusTitle = "Status"
		dateTitle   = "Date"
	)

	var baseOutput = map[string]map[string]interface{}{
		expiryTitle: {
			statusTitle: details.Status,
		},
	}

	if details.Status == "INACTIVE" {
		return baseOutput
	}

	var spec serviceSpec
	if err := mapstructure.Decode(details.Spec, &spec); err != nil {
		log.Errorf("failed to unmarshal output: %s", err.Error())
		return baseOutput
	}

	return map[string]map[string]interface{}{
		expiryTitle: {
			statusTitle: details.Status,
			dateTitle:   spec.Date,
		},
	}
}

func askForDate(defaultValue string) (string, error) {
	var date string

	for {
		for {
			var formattedNow = time.Now().UTC().Format(time.RFC3339)
			if err := input.DoQuestions([]input.QuestionMaker{
				input.QuestionInput{
					QuestionBase: input.QuestionBase{
						Message: fmt.Sprintf("Provide expiration date in UTC ( your local time in UTC is %s ):", formattedNow),
						Help:    fmt.Sprintf("Date format should be: %s", formattedNow),
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
