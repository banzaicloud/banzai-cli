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

package secret

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

// updateSecretOptions contains update secret flags for `banzai update secret` command
type updateSecretOptions struct {
	file       string
	secretName string
	secretID   string
	validate   string
	format     string
}

type secretNotFoundError struct {
	name  string
	value string
}

func (e *secretNotFoundError) Error() string {
	return fmt.Sprintf("could not find secret with %s: '%s' or it's a readonly secret", e.name, e.value)
}

// NewUpdateCommand returns a cobra command for `banzai secret update` command
func NewUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateSecretOptions{}

	cmd := &cobra.Command{
		Example: `
	Update secret
	---
	$ banzai secret update
	? Select secret: mysecret
	? Do you want modify fields of secret? Yes
	? Select field to modify: username
	? username myusername
	? Select field to modify: password
	? password mypassword
	? Select field to modify: skip
	? Do you want modify tags of secret? Yes
	? Do you want delete any tag of secret? Yes
	? Select tag(s) you want to delete: cli
	? Do you want to add tag(s) to this secret? Yes
	? Tag: banzai
	? Tag: skip
	? Do you want to validate this secret? Yes

	Update secret with flags
	---
	$ banzai secret update --name mysecret --validate false
	? Do you want modify fields of secret? Yes
	? Select field to modify: username
	? username myusername
	? Select field to modify: password
	? password mypassword
	? Select field to modify: skip
	? Do you want modify tags of secret? No
	
	Create secret via json
	---
	$ banzai secret update <<EOF
	> {
	>	"name": "mysecretname",
	>	"type": "password",
	>	"values": {
	>		"username": "myusername",
	>		"password": "mypassword"
	>	},
	>	"tags":[ "cli", "my-application" ],
	> 	"version": 1
	> }
	> EOF

`,
		Use:          "update",
		Aliases:      []string{"u"},
		Short:        "Update secret",
		Long:         "Update an existing secret in Pipeline's secret store interactively, or based on a json request from stdin or a file",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			return runUpdateSecret(banzaiCli, &options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Secret update descriptor file")
	flags.StringVarP(&options.secretName, "name", "n", "", "Name of the secret")
	flags.StringVarP(&options.secretID, "id", "i", "", "identification of the secret")
	flags.StringVarP(&options.validate, "validate", "v", "", "Secret validation (true|false)")

	return cmd
}

// runUpdateSecret starts to get secret update properties from the user via file or survey
func runUpdateSecret(banzaiCli cli.Cli, options *updateSecretOptions) error {
	out := &pipeline.CreateSecretRequest{}

	if err := getUpdateSecretRequest(banzaiCli, options, out); err != nil {
		return err
	}

	log.Debugf("update secret request: %#v", out)

	if len(options.secretID) == 0 {
		return errors.New("missing required --id flag")
	}

	orgID := input.GetOrganization(banzaiCli)
	response, _, err := banzaiCli.Client().SecretsApi.UpdateSecrets(
		context.Background(),
		orgID,
		options.secretID,
		*out,
		&pipeline.UpdateSecretsOpts{
			Validate: getValidationFlag(options.validate),
		},
	)
	if err != nil {
		cli.LogAPIError("create secret", err, out)
		return emperror.Wrap(err, "failed to create secret")
	}

	format.SecretWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), response)

	return nil
}

func getUpdateSecretRequest(banzaiCli cli.Cli, options *updateSecretOptions, out *pipeline.CreateSecretRequest) error {
	if banzaiCli.Interactive() {
		return buildInteractiveUpdateSecretRequest(banzaiCli, options, out)
	} else {
		return readFileAndValidate(options.file, out)
	}
}

func buildInteractiveUpdateSecretRequest(banzaiCli cli.Cli, options *updateSecretOptions, out *pipeline.CreateSecretRequest) error {
	fileReadOk := false
	if len(options.file) != 0 {
		if err := readCreateSecretRequestFromFile(options.file, out); err != nil {
			// failed to load file, we can ask the user via survey
			cli.LogAPIError("update secret", err, out)
		} else {
			fileReadOk = true
			options.secretName = out.Name
		}
	}

	// get secrets from API, filter out hidden and readonly secrets
	secrets, err := getSecrets(banzaiCli)
	if err != nil {
		return err
	}

	// ask for secret if --id and --name flags are not defined
	surveySecret(secrets, options)

	// search for selected secret
	selectedSecret, err := findSelectedSecret(secrets, options)
	if err != nil {
		return err
	}

	// set secret id, need to the path
	options.secretID = selectedSecret.Id

	if fileReadOk {
		// ask for validation if needed
		options.validate = confirmValidation(options.validate, selectedSecret.Type)
		out.Version = selectedSecret.Version

		return nil
	}

	// ask for field(s) update
	surveyFieldsToUpdate(banzaiCli, selectedSecret)

	// ask for tag(s) update
	surveyTagsUpdate(selectedSecret)

	// ask for validation if needed
	options.validate = confirmValidation(options.validate, selectedSecret.Type)

	out.Name = selectedSecret.Name
	out.Type = selectedSecret.Type
	out.Values = selectedSecret.Values
	out.Tags = selectedSecret.Tags
	out.Version = selectedSecret.Version

	return nil
}

// getSecrets returns secrets from Pipeline and filter out the `hidden` and `readonly` secrets
func getSecrets(banzaiCli cli.Cli) ([]pipeline.SecretItem, error) {
	orgID := input.GetOrganization(banzaiCli)
	secretsFromAPI, _, err := banzaiCli.Client().SecretsApi.GetSecrets(
		context.Background(),
		orgID,
		&pipeline.GetSecretsOpts{
			Values: optional.NewBool(true),
		})
	if err != nil {
		cli.LogAPIError("could not list secrets", err, nil)
		return nil, emperror.Wrap(err, "could not list secrets")
	}

	filteredSecrets := filterOutHiddenAndReadonlySecrets(secretsFromAPI)

	return filteredSecrets, nil
}

// surveySecret asks the user for an secret to modify
// in case of --id or --name flags are defined we skip this option
func surveySecret(secrets []pipeline.SecretItem, options *updateSecretOptions) {
	if len(options.secretName) == 0 && len(options.secretID) == 0 {
		secret := ""
		prompt := &survey.Select{
			Message:  "Select secret:",
			Options:  getSecretNames(secrets),
			Help:     "Select an existing secret you want to update",
			PageSize: 20,
		}
		_ = survey.AskOne(prompt, &secret, survey.Required)

		options.secretName = secret
	}
}

// getSecretNames returns secret name slice for select survey
func getSecretNames(secrets []pipeline.SecretItem) []string {
	names := make([]string, len(secrets))
	for _, s := range secrets {
		names = append(names, s.Name)
	}

	return names
}

// findSelectedSecret searching for secret by `id` or `name`
func findSelectedSecret(secrets []pipeline.SecretItem, options *updateSecretOptions) (*pipeline.SecretItem, error) {
	for idx, s := range secrets {
		if s.Id == options.secretID || s.Name == options.secretName {
			return &secrets[idx], nil
		}
	}

	var err *secretNotFoundError
	if len(options.secretID) != 0 {
		err = &secretNotFoundError{
			name:  "id",
			value: options.secretID,
		}
	} else {
		err = &secretNotFoundError{
			name:  "name",
			value: options.secretName,
		}
	}

	return nil, err
}

// surveyFieldsToUpdate asks fields to modify
func surveyFieldsToUpdate(banzaiCli cli.Cli, secret *pipeline.SecretItem) {

	var fieldUpdate bool
	prompt := &survey.Confirm{
		Message: "Do you want modify fields of secret?",
		Default: false,
	}
	_ = survey.AskOne(prompt, &fieldUpdate, nil)

	if fieldUpdate {
		allowedTypes := getSupportedFieldsFromAPI(banzaiCli, secret)
		optionsForFields := getSupportedFieldNames(allowedTypes)
		for {

			fieldName := ""
			prompt := &survey.Select{
				Message:  "Select field to modify:",
				Options:  optionsForFields,
				Default:  "skip",
				Help:     "Leave empty to update.",
				PageSize: 20,
			}
			_ = survey.AskOne(prompt, &fieldName, nil)

			if fieldName == "skip" {
				return
			}

			if secret.Values[fieldName] == nil {
				secret.Values[fieldName] = ""
			}

			var value string
			i := &survey.Input{
				Message: fmt.Sprintf(fieldName),
				Default: fmt.Sprintf("%s", secret.Values[fieldName]),
				Help:    getFieldHelp(allowedTypes, fieldName),
			}
			isRequired := isFieldRequired(allowedTypes, fieldName)
			v := survey.Required
			if !isRequired {
				v = nil
			}

			_ = survey.AskOne(i, &value, v)
			secret.Values[fieldName] = value
		}
	}

}

// getSupportedFieldsFromAPI returns supported fields from API base on secret type
func getSupportedFieldsFromAPI(banzaiCli cli.Cli, secret *pipeline.SecretItem) []pipeline.AllowedSecretTypeResponseFields {
	types, _, err := banzaiCli.Client().SecretsApi.AllowedSecretsTypesKeys(context.Background(), secret.Type)
	if err != nil {
		cli.LogAPIError("could not list keys for secret type", err, nil)
		log.Fatalf("could not list keys for secret type: %v", err)
	}
	return types.Fields
}

// getSupportedFieldNames returns field name slice and add `skip` option to the end
func getSupportedFieldNames(fields []pipeline.AllowedSecretTypeResponseFields) []string {
	options := make([]string, len(fields))
	for _, f := range fields {
		options = append(options, f.Name)
	}
	return append(options, "skip")
}

// isFieldRequired decides the given field is required or not
func isFieldRequired(fields []pipeline.AllowedSecretTypeResponseFields, fieldName string) bool {
	for _, f := range fields {
		if f.Name == fieldName {
			return f.Required
		}
	}
	return false
}

// isFieldRequired returns description for the given field
func getFieldHelp(fields []pipeline.AllowedSecretTypeResponseFields, fieldName string) string {
	for _, f := range fields {
		if f.Name == fieldName {
			return f.Description
		}
	}
	return ""
}

// surveyTagsUpdate ask for tags delete/add
func surveyTagsUpdate(secret *pipeline.SecretItem) {
	var tagUpdate bool
	prompt := &survey.Confirm{
		Message: "Do you want modify tags of secret?",
		Default: false,
	}
	_ = survey.AskOne(prompt, &tagUpdate, nil)

	if tagUpdate {
		// ask for delete
		tagsToDelete := surveyTagsMarkDelete(secret.Tags)

		// ask to add
		tagsToAdd := surveyTags(nil)

		for _, t := range secret.Tags {
			if !contains(tagsToDelete, []string{t}) {
				tagsToAdd = append(tagsToAdd, t)
			}
		}

		secret.Tags = tagsToAdd
	}

}

// surveyTagsMarkDelete ask for tag(s) to delete via MultiSelect
func surveyTagsMarkDelete(tags []string) []string {
	var tagsMarkedToDelete []string
	if tags != nil && len(tags) != 0 {
		// ask delete
		var tagDelete bool
		prompt := &survey.Confirm{
			Message: "Do you want delete any tag of secret?",
			Default: false,
		}
		_ = survey.AskOne(prompt, &tagDelete, nil)

		// mark for delete
		if tagDelete {
			prompt := &survey.MultiSelect{
				Message: "Select tag(s) you want to delete:",
				Options: tags,
			}
			_ = survey.AskOne(prompt, &tagsMarkedToDelete, nil)

		}

	}
	return tagsMarkedToDelete
}
