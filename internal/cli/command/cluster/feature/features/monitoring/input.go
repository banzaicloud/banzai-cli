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

package monitoring

import (
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

type questionMaker interface {
	Do() error
}

type questionBase struct {
	message string
	help    string
}

type questionConfirm struct {
	questionBase
	defaultValue bool
	output       *bool
}

type questionInput struct {
	questionBase
	defaultValue string
	output       *string
}

type questionSelect struct {
	questionInput
	options []string
}

func (q questionConfirm) Do() error {
	if err := survey.AskOne(
		&survey.Confirm{
			Message: q.message,
			Default: q.defaultValue,
			Help:    q.help,
		},
		q.output,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	return nil
}

func (q questionInput) Do() error {
	if err := survey.AskOne(
		&survey.Input{
			Message: q.message,
			Help:    q.help,
			Default: q.defaultValue,
		},
		q.output,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	return nil
}

func (q questionSelect) Do() error {
	err := survey.AskOne(
		&survey.Select{
			Message: q.message,
			Help:    q.help,
			Options: q.options,
			Default: q.defaultValue,
		},
		q.output,
	)
	if err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	return nil
}

func doQuestions(questions []questionMaker) error {
	for _, q := range questions {
		if err := q.Do(); err != nil {
			return errors.WrapIf(err, "failure during survey")
		}
	}
	return nil
}
