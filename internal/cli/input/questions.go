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
	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

type QuestionMaker interface {
	Do() error
}

type QuestionBase struct {
	Message string
	Help    string
}

type QuestionConfirm struct {
	QuestionBase
	DefaultValue bool
	Output       *bool
}

type QuestionInput struct {
	QuestionBase
	DefaultValue string
	Output       *string
}

type QuestionSelect struct {
	QuestionInput
	Options []string
}

func (q QuestionConfirm) Do() error {
	if err := survey.AskOne(
		&survey.Confirm{
			Message: q.Message,
			Default: q.DefaultValue,
			Help:    q.Help,
		},
		q.Output,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	return nil
}

func (q QuestionInput) Do() error {
	if err := survey.AskOne(
		&survey.Input{
			Message: q.Message,
			Help:    q.Help,
			Default: q.DefaultValue,
		},
		q.Output,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	return nil
}

func (q QuestionSelect) Do() error {
	err := survey.AskOne(
		&survey.Select{
			Message: q.Message,
			Help:    q.Help,
			Options: q.Options,
			Default: q.DefaultValue,
		},
		q.Output,
	)
	if err != nil {
		return errors.WrapIf(err, "failure during survey")
	}

	return nil
}

func DoQuestions(questions []QuestionMaker) error {
	for _, q := range questions {
		if err := q.Do(); err != nil {
			return errors.WrapIf(err, "failure during survey")
		}
	}
	return nil
}
