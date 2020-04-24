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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"emperror.dev/errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	TypeGeneric    = "generic"
	TypeAmazon     = "amazon"
	TypeAzure      = "azure"
	TypeAlibaba    = "alibaba"
	TypeGoogle     = "google"
	TypeOracle     = "oracle"
	TypeKubernetes = "kubernetes"
)

// createSecretOptions contains create secret flags for `banzai create secret` command
type createSecretOptions struct {
	file       string
	secretName string
	secretType string
	tags       []string
	validate   string
	format     string
	magic      bool
}

// secretFieldQuestion contains all necessary field for a secret question (any type except generic)
type secretFieldQuestion struct {
	input    *survey.Password
	name     string
	output   string
	required bool
}

// NewCreateCommand returns a cobra command for `banzai create create` command
func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createSecretOptions{}

	cmd := &cobra.Command{
		Example: `
	Create secret
	---
	$ banzai secret create
	? Secret name mysecretname
	? Choose secret type: password
	? Set 'username' field: myusername
	? Set 'password' field: mypassword
	? Do you want to add tag(s) to this secret? Yes
	? Tag: tag1
	? Tag: tag2
	? Tag: skip

	Create secret with flags
	---
	$ banzai secret create --name mysecretname --type password --tag=cli --tag=my-application
	? Set 'username' field: myusername
	? Set 'password' field: mypassword

	Create secret via json
	---
	$ banzai secret create <<EOF
	> {
	>	"name": "mysecretname",
	>	"type": "password",
	>	"values": {
	>		"username": "myusername",
	>		"password": "mypassword"
	>	},
	>	"tags":[ "cli", "my-application" ]
	> }
	> EOF
		`,
		Use:          "create",
		Aliases:      []string{"c"},
		Short:        "Create secret",
		Long:         "Create a secret in Pipeline's secret store interactively, or based on a json request from stdin or a file",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			return runCreateSecret(banzaiCli, &options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Secret creation descriptor file")
	flags.StringVarP(&options.secretName, "name", "n", "", "Name of the secret")
	flags.StringVarP(&options.secretType, "type", "t", "", "Type of the secret")
	flags.StringArrayVarP(&options.tags, "tag", "", []string{}, "Tags to add to the secret")
	flags.StringVarP(&options.validate, "validate", "v", "", "Secret validation (true|false)")
	flags.BoolVar(&options.magic, "magic", false, "Try to import credentials from local environment (AWS only for now)")

	return cmd
}

// runCreateSecret starts to get secret properties from the user via file or survey
func runCreateSecret(banzaiCli cli.Cli, options *createSecretOptions) error {
	out := &pipeline.CreateSecretRequest{}

	if err := getCreateSecretRequest(banzaiCli, options, out); err != nil {
		return err
	}

	log.Debugf("create secret request: %#v", out)

	orgID := input.GetOrganization(banzaiCli)
	response, _, err := banzaiCli.Client().SecretsApi.AddSecrets(
		context.Background(),
		orgID,
		*out,
		&pipeline.AddSecretsOpts{
			Validate: getValidationFlag(options.validate),
		},
	)
	if err != nil {
		cli.LogAPIError("create secret", err, out)
		return errors.WrapIf(err, "failed to create secret")
	}

	format.SecretWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), response)

	return nil
}

func getCreateSecretRequest(banzaiCli cli.Cli, options *createSecretOptions, out *pipeline.CreateSecretRequest) error {
	out.Name = options.secretName
	out.Type = options.secretType
	out.Tags = options.tags

	if banzaiCli.Interactive() {
		return buildInteractiveCreateSecretRequest(banzaiCli, options, out)
	} else {
		if values, err := importLocalCredential(banzaiCli, options); err != nil {
			return err
		} else if values != nil {
			// TODO fix openapi
			out.Values = map[string]interface{}{}
			for k, v := range values {
				out.Values[k] = v
			}
		}

		if options.file == "" && options.magic {
			return nil
		}

		return readFileAndValidate(options.file, out)
	}
}

func readFileAndValidate(filename string, out *pipeline.CreateSecretRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filename)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := validateCreateSecretRequest(raw); err != nil {
		return errors.WrapIf(err, "failed to parse create cluster request")
	}

	if err := utils.Unmarshal(raw, &out); err != nil {
		return errors.WrapIf(err, "failed to unmarshal create cluster request")
	}

	return nil
}

func validateCreateSecretRequest(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		if bytes, ok := val.([]byte); ok {
			str = string(bytes)
		} else {
			return errors.New("value is not a string or []byte")
		}
	}

	decoder := json.NewDecoder(strings.NewReader(str))

	var typer struct{ Type string }
	err := decoder.Decode(&typer)
	if err != nil {
		return errors.WrapIf(err, "invalid JSON request")
	}

	decoder = json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()

	return errors.WrapIf(decoder.Decode(&pipeline.CreateSecretRequest{}), "invalid request")
}

func buildInteractiveCreateSecretRequest(banzaiCli cli.Cli, options *createSecretOptions, out *pipeline.CreateSecretRequest) error {
	if options.file != "" {
		return readCreateSecretRequestFromFile(options.file, out)
	}

	secretTypes, _, err := banzaiCli.Client().SecretsApi.ListSecretTypes(context.Background())
	if err != nil {
		cli.LogAPIError("could not list secret types", err, nil)
		log.Fatalf("could not list secret types: %v", err)
	}

	surveySecretName(options)

	surveySecretType(options, secretTypes)

	values, err := importLocalCredential(banzaiCli, options)
	if err != nil {
		return err
	} else if values == nil {
		if err := surveySecretFields(options, secretTypes, out); err != nil {
			log.Fatalf("could not get secret fields: %v", err)
		}
	} else {
		// TODO fix openapi
		out.Values = map[string]interface{}{}
		for k, v := range values {
			out.Values[k] = v
		}
	}

	surveyTags(options)

	out.Name = options.secretName
	out.Type = options.secretType
	out.Tags = options.tags

	if options.validate == "" {
		if options.secretType == TypeAmazon ||
			options.secretType == TypeAzure ||
			options.secretType == TypeAlibaba ||
			options.secretType == TypeGoogle ||
			options.secretType == TypeOracle {
			// request validation just in case of cloud types
			options.validate = "true"
			var v bool
			prompt := &survey.Confirm{
				Message: "Do you want to validate this secret?",
				Help:    "Pipeline can optionally try to connect to the cloud provider, and execute some basic tests.",
				Default: true,
			}
			_ = survey.AskOne(prompt, &v)
			if !v {
				options.validate = "false"
			}
		}
	}

	return nil
}

// surveyGenericSecretType starts to get fields (key/value pair) for generic secret
func surveyGenericSecretType(out *pipeline.CreateSecretRequest) {
	out.Values = make(map[string]interface{})

	for {
		// ask for key
		var key string
		_ = survey.AskOne(
			&survey.Input{
				Message: "Key of field:",
			},
			&key,
			survey.WithValidator(survey.Required),
		)

		// ask for value
		var value string
		_ = survey.AskOne(
			&survey.Input{
				Message: "Value of field:",
			},
			&value,
			survey.WithValidator(survey.Required),
		)

		// add to values field
		out.Values[key] = value

		// confirm continue
		isContinue := false
		prompt := &survey.Confirm{
			Message: "Do you want to add another key/value pair?",
		}
		_ = survey.AskOne(prompt, &isContinue)
		if !isContinue {
			return
		}
	}
}

// readCreateSecretRequestFromFile reads file from the getting filename into CreateSecretRequest
func readCreateSecretRequestFromFile(fileName string, out *pipeline.CreateSecretRequest) error {
	if raw, err := ioutil.ReadFile(fileName); err != nil {
		return errors.WrapIff(err, "failed to read file: %s", fileName)
	} else if err := utils.Unmarshal(raw, &out); err != nil {
		return errors.WrapIf(err, "failed to parse CreateSecretRequest")
	}
	return nil
}

// surveySecretName starts to get secret name from the user
func surveySecretName(options *createSecretOptions) {
	if len(options.secretName) == 0 {
		_ = survey.AskOne(&survey.Input{Message: "Secret name"},
			&options.secretName,
			survey.WithValidator(survey.Required),
		)
	}
}

// surveySecretType starts to get secret type from the user
func surveySecretType(options *createSecretOptions, secretTypes map[string]pipeline.SecretTypeResponse) {
	if len(options.secretType) == 0 {
		var typeOptions []string
		for name := range secretTypes {
			typeOptions = append(typeOptions, name)
		}

		sort.Strings(typeOptions)

		selectTypePrompt := &survey.Select{
			Message:  "Choose secret type:",
			Options:  typeOptions,
			PageSize: 13,
		}
		_ = survey.AskOne(selectTypePrompt, &options.secretType, survey.WithValidator(survey.Required))
	}
}

// surveySecretFields starts to get secret fields base on selected secret type and pipeline response
func surveySecretFields(options *createSecretOptions, secretTypes map[string]pipeline.SecretTypeResponse, out *pipeline.CreateSecretRequest) error {
	if options.secretType == TypeGeneric {
		surveyGenericSecretType(out)
	} else if secretType, ok := secretTypes[options.secretType]; ok {
		// set fields
		fields := secretType.Fields
		questions := make([]secretFieldQuestion, len(fields))
		for index, f := range fields {
			questions[index] = secretFieldQuestion{
				name: f.Name,
				input: &survey.Password{
					Message: f.Name,
					Help:    f.Description,
				},
				required: f.Required,
			}
		}

		for i, q := range questions {
			opts := []survey.AskOpt{}
			if q.required {
				opts = append(opts, survey.WithValidator(survey.Required))
			}
			if err := survey.AskOne(q.input, &questions[i].output, opts...); err != nil {
				return errors.WrapIf(err, "failed to ask for value")
			}
		}

		// set create secret request fields
		out.Values = make(map[string]interface{})

		for _, q := range questions {
			if len(q.output) != 0 {
				out.Values[q.name] = q.output
			}
		}
	} else {
		return errors.New("not supported secret type")
	}

	return nil
}

// surveyTags starts to get tag(s) for the secret until `skip`
func surveyTags(options *createSecretOptions) {
	if options.tags == nil || len(options.tags) == 0 {
		isTagAdd := false
		prompt := &survey.Confirm{
			Message: "Do you want to add tag(s) to this secret?",
		}
		_ = survey.AskOne(prompt, &isTagAdd)

		if isTagAdd {
			for {
				var tag string
				_ = survey.AskOne(
					&survey.Input{
						Message: "Tag:",
						Default: "skip",
						Help:    "Leave empty to cancel.",
					},
					&tag,
					survey.WithValidator(survey.Required),
				)

				if tag == "skip" {
					return
				}

				options.tags = append(options.tags, tag)
			}
		}
	}
}

func getValidationFlag(validation string) optional.Bool {
	switch validation {
	case "false":
		return optional.NewBool(false)
	default:
		return optional.NewBool(true)
	}
}

func importLocalCredential(banzaiCli cli.Cli, options *createSecretOptions) (map[string]string, error) {
	if !banzaiCli.Interactive() && !options.magic {
		return nil, nil
	}

	var id string
	var values map[string]string
	var err error

	switch options.secretType {
	case TypeAmazon:
		id, values, err = input.GetAmazonCredentials()
	case TypeKubernetes:
		var config []byte
		id, config, err = input.GetCurrentKubecontext()

		values = map[string]string{
			"K8Sconfig": string(config),
		}
	default:
		if options.magic {
			return nil, errors.New("unsupported secret type for local credential import")
		}
		return nil, nil
	}

	if values != nil && !options.magic && banzaiCli.Interactive() {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Do you want to create the secret from your local credential (%s)?", id),
			Help:    fmt.Sprintf("We can extract your local AWS credentials if you want."),
		}
		_ = survey.AskOne(prompt, &options.magic)
	}

	if options.magic {
		return values, err
	}
	return nil, nil
}
