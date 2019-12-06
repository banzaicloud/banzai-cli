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

package controlplane

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/login"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

const (
	generatedValuesFileName = "generated-values-cli.yaml"
)

type createOptions struct {
	init bool
	*initOptions
}

// NewUpCommand creates a new cobra.Command for `banzai pipeline up`.
func NewUpCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"c"},
		Short:   "Deploy Banzai Cloud Pipeline",
		Long:    `Deploy or upgrade an instance of Banzai Cloud Pipeline based on a values file in the workspace, or initialize the workspace from an input file or an interactive session.` + initLongDescription,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runUp(&options, banzaiCli)
		},
	}

	options.initOptions = newInitOptions(cmd, banzaiCli)

	flags := cmd.Flags()
	flags.BoolVarP(&options.init, "init", "i", false, "Initialize workspace")

	return cmd
}

func printExternalHostRecord(host, lbAddress string) {
	lbRecordType := "A"
	if net.ParseIP(lbAddress) == nil {
		lbRecordType = "CNAME"
	}

	fmt.Printf("Please create a DNS record pointing to the load balancer:\n\n%s IN %s %s\n", host, lbRecordType, lbAddress)
}

func runUp(options *createOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	if !options.valuesExists() {
		if !options.init && banzaiCli.Interactive() {
			if err := survey.AskOne(
				&survey.Confirm{
					Message: "The workspace is not initialized. Do you want to initialize it now?",
					Default: true,
				},
				&options.init,
			); err != nil {
				options.init = false
			}
		}
		if options.init {
			if err := runInit(*options.initOptions, banzaiCli); err != nil {
				return err
			}
		} else {
			return errors.New("workspace is uninitialized")
		}
	} else {
		log.Debugf("using existing workspace %q", options.workspace)
		if options.initOptions.file != "" {
			return errors.New("workspace is already initialized but --file is specified")
		}
	}

	var values map[string]interface{}
	if err := options.readValues(&values); err != nil {
		return err
	}

	if uuidValue, ok := values["uuid"]; !ok {
		if uuidString, ok := uuidValue.(string); !ok || uuidString == "" {
			log.Infof("An uuid field that identifies the Banzai Cloud Pipeline instance to deploy is missing from the values file. You can add one with `echo 'uuid: %s' >>%q`", uuid.New().String(), options.valuesPath())
			return errors.New("uuid field is missing from the values file")
		}
	}

	if options.provider != "" && options.provider != values["provider"] {
		return errors.New("workspace is already initialized but a different --provider is specified")
	}

	source := "/export"

	// for backward compatibility
	hasExports, err := imageFileExists(options.cpContext, source)
	if err != nil {
		return err
	}

	if hasExports {
		var defaultValues map[string]interface{}
		exportHandlers := []ExportedFilesHandler{
			defaultValuesExporter(&defaultValues),
		}
		if err := processExports(options.cpContext, source, exportHandlers); err != nil {
			return err
		}
		if err := writeMergedValues(options.cpContext, defaultValues, values); err != nil {
			return err
		}
	} else {
		log.Warnf("%s is not available in the image, skipping export handlers", source)
		// this is the legacy behaviour that should be removed as soon as we can deprecate old image versions
		// where the null_resource.preapply_hook did the merging
		if err := runTerraform("apply", options.cpContext, nil, "null_resource.preapply_hook"); err != nil {
			return errors.WrapIf(err, "failed to run null_resource.preapply_hook as a standalone target")
		}
	}

	var env map[string]string
	switch values["provider"] {
	case providerPke:
		err := ensurePKECluster(banzaiCli, options.cpContext)
		if err != nil {
			return errors.WrapIf(err, "failed to deploy PKE cluster")
		}

	case providerKind:
		err := ensureKINDCluster(banzaiCli, options.cpContext)
		if err != nil {
			return errors.WrapIf(err, "failed to create KIND cluster")
		}

	case providerEc2:
		_, creds, err := input.GetAmazonCredentials()
		if err != nil {
			return errors.WrapIf(err, "failed to get AWS credentials")
		}

		useGeneratedKey := true
		if pc, ok := values["providerConfig"]; ok {
			if pc, ok := pc.(map[string]interface{}); ok {
				useGeneratedKey = pc["key_name"] != nil && pc["key_name"] != ""
			}
		}

		if err := ensureEC2Cluster(options.cpContext, creds, useGeneratedKey); err != nil {
			return errors.WrapIf(err, "failed to create EC2 cluster")
		}
		env = creds

	case providerCustom:
		creds := map[string]string{}
		if pc, ok := values["providerConfig"]; ok {
			pc := cast.ToStringMap(pc)
			if _, ok := pc["accessKey"]; ok {
				_, awsCreds, err := input.GetAmazonCredentials()
				if err != nil {
					return errors.WrapIf(err, "failed to get AWS credentials")
				}
				creds = awsCreds
			}
		}

		if err := ensureCustomCluster(options.cpContext, creds); err != nil {
			return errors.WrapIf(err, "failed to create Custom infrastructure")
		}

	default:
		if !options.kubeconfigExists() {
			return errors.New("could not find Kubeconfig in workspace")
		}
	}

	if !isStateBackendInited(options) {
		log.Info("Migrating workspace to state backend...")
		if err := initStateBackend(options.cpContext); err != nil {
			return err
		}
	}

	log.Info("Deploying Banzai Cloud Pipeline to Kubernetes cluster...")
	if values["provider"] == providerCustom {
		_, creds, err := input.GetAmazonCredentials()
		if err != nil {
			return errors.WrapIf(err, "failed to get AWS credentials")
		}
		env = map[string]string{}
		for k, v := range creds {
			env[k] = v
		}
	}

	if err := runTerraform("apply", options.cpContext, env); err != nil {
		return errors.WrapIf(err, "failed to deploy pipeline components")
	}

	return postInstall(options, banzaiCli, values)
}

func isStateBackendInited(options *createOptions) bool {
	_, err := os.Stat(options.workspace + "/.terraform/terraform.tfstate")
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func postInstall(options *createOptions, banzaiCli cli.Cli, values map[string]interface{}) error {
	url, err := options.readExternalAddress()
	if err != nil {
		return errors.WrapIf(err, "can't read final URL of Pipeline")
	}
	log.Infof("Pipeline is ready at %s.", url)
	url += "pipeline"

	externalHost, _ := values["externalHost"].(string)

	if externalHost != "auto" && externalHost != defaultLocalhost {
		var target string
		switch values["provider"] {
		case providerKind:
			target = "127.0.0.1"
		case providerEc2:
			target, err = options.readEc2Host()
			if err != nil {
				log.Errorf("%v", err)
			}
		case providerK8s, providerCustom:
			target, err = options.readTraefikAddress()
			if err != nil {
				log.Errorf("%v", err)
			}
		}

		if target != "" {
			printExternalHostRecord(externalHost, target)
		}
	}

	var loginNow bool
	if banzaiCli.Interactive() {
		if err := survey.AskOne(
			&survey.Confirm{
				Message: "Do you want to login this CLI tool now?",
				Default: true,
			},
			&loginNow,
		); err != nil {
			loginNow = false
		}
	}

	log.Infof("The certificate of this environment is signed by an unknown authority by default. You can safely accept this.")

	if loginNow {
		return login.Login(banzaiCli, url, "", true, false)
	} else {
		log.Infof("Pipeline is ready, now you can login with: \x1b[1mbanzai login --endpoint=%q\x1b[0m", url)
	}
	return nil
}

type ExportedFilesHandler func(string) error

func processExports(options *cpContext, source string, exportedFilesHandlers []ExportedFilesHandler) error {
	destination, err := ioutil.TempDir(options.workspace, "export")
	if err != nil {
		return errors.Wrap(err, "failed to create temp directory")
	}

	if err := exportFilesFromContainer(options, source, destination); err != nil {
		return errors.WrapIf(err, "failed to export files from the image")
	}

	for _, h := range exportedFilesHandlers {
		if err := h(filepath.Join(destination, source)); err != nil {
			return errors.WrapIff(err, "failed to run handler on exported files %s", destination)
		}
	}

	if err := os.RemoveAll(destination); err != nil {
		return errors.Wrap(err, "failed to remove temporary exports directory")
	}

	return nil
}

func writeMergedValues(options *cpContext, defaultValues, overrideValues map[string]interface{}) error {
	var mergedValues map[string]interface{}
	if err := mergeValues(&mergedValues, defaultValues, overrideValues); err != nil {
		return err
	}
	bytes, err := yaml.Marshal(&mergedValues)
	if err != nil {
		return errors.Wrap(err, "failed to marshal merged values")
	}
	if err := ioutil.WriteFile(filepath.Join(options.workspace, generatedValuesFileName ), bytes, 0600); err != nil {
		return errors.Wrap(err, "failed to write out generated values file")
	}
	return nil
}

func mergeValues(mergedValues *map[string]interface{}, defaultValues, overrideValues map[string]interface{}) error {
	if err := mergo.Merge(mergedValues, &defaultValues, mergo.WithOverride); err != nil {
		return errors.Wrap(err, "failed to process default values in the image")
	}
	if err := mergo.Merge(mergedValues, &overrideValues, mergo.WithOverride); err != nil {
		return errors.Wrap(err, "failed to merge override values from the workspace on top of default values in the image")
	}
	return nil
}

func imageFileExists(options *cpContext, source string) (bool, error) {
	errorMsg := &bytes.Buffer{}
	cmdOpt := func(cmd *exec.Cmd) error {
		cmd.Stderr = errorMsg
		return nil
	}
	if err := runContainerCommandGeneric(options, []string{"ls", source}, nil, cmdOpt); err != nil {
		if strings.Contains(errorMsg.String(), "No such file or directory") {
			return false, nil
		}
		return false, err
	} else {
		return true, nil
	}
}

func defaultValuesExporter(defaultValues *map[string]interface{}) ExportedFilesHandler {
	return ExportedFilesHandler(func(destination string) error {
		destBytes, err := ioutil.ReadFile(filepath.Join(destination, "values.yaml"))
		if err != nil {
			return errors.Wrapf(err, "failed to read %s", destination)
		}
		if err := yaml.Unmarshal(destBytes, defaultValues); err != nil {
			return errors.Wrap(err, "failed to unmarshal default values exported from the image")
		}
		return nil
	})
}
