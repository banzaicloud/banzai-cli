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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/login"
)

const (
	generatedValuesFileName = "generated-values-cli.yaml"
)

type createOptions struct {
	init          bool
	terraformInit bool
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
	flags.BoolVar(&options.terraformInit, "terraform-init", true, "Run terraform init before apply")

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

	_, env, err := getImageMetadata(options.cpContext, values, true)
	if err != nil {
		return err
	}

	if options.terraformInit {
		if err := initStateBackend(options.cpContext, values, env); err != nil {
			return errors.WrapIf(err, "failed to initialize state backend")
		}
	}

	switch values["provider"] {
	case providerPke:
		err := ensurePKECluster(banzaiCli, options.cpContext)
		if err != nil {
			return errors.WrapIf(err, "failed to deploy PKE cluster")
		}

	case providerKind:
		listenAddress := "127.0.0.1"
		if pc, ok := values["providerConfig"]; ok {
			if pc, ok := pc.(map[interface{}]interface{}); ok {
				if addr, ok := pc["listenAddress"].(string); ok && addr != "" {
					listenAddress = addr
				}
			}
		}
		log.Debugf("listening on %q", listenAddress)
		err := ensureKINDCluster(banzaiCli, options.cpContext, listenAddress)
		if err != nil {
			return errors.WrapIf(err, "failed to create KIND cluster")
		}

	case providerEc2:
		useGeneratedKey := true
		if pc, ok := values["providerConfig"]; ok {
			if pc, ok := pc.(map[interface{}]interface{}); ok {
				if pc["key_name"] != nil && pc["key_name"] != "" {
					useGeneratedKey = false
				}
			}
		}
		if err := ensureEC2Cluster(options.cpContext, env, useGeneratedKey); err != nil {
			return errors.WrapIf(err, "failed to create EC2 cluster")
		}

	case providerCustom:
		if err := ensureCustomCluster(options.cpContext, env); err != nil {
			return errors.WrapIf(err, "failed to create Custom infrastructure")
		}

	default:
		if !options.kubeconfigExists() {
			return errors.New("could not find Kubeconfig in workspace")
		}
	}

	log.Info("Deploying Banzai Cloud Pipeline to Kubernetes cluster...")

	if err := runTerraform("apply", options.cpContext, env); err != nil {
		return errors.WrapIf(err, "failed to deploy pipeline components")
	}

	return postInstall(options, banzaiCli, values)
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

func initStateBackend(options *cpContext, values map[string]interface{}, env map[string]string) error {
	var stateData []byte

	if stateValues, ok := values["state"]; ok {
		values := stringifyMap(stateValues)
		var err error
		stateData, err = json.MarshalIndent(values, "", "  ")
		if err != nil {
			return errors.WrapIf(err, "failed to marshal state backend configuration")
		}
		options.explicitState = true
	} else {
		stateData = []byte(fmt.Sprintf(localStateBackend, tfstateFilename))
	}

	err := ioutil.WriteFile(options.workspace+"/state.tf.json", stateData, 0600)
	if err != nil {
		return errors.WrapIf(err, "failed to create state backend configuration")
	}

	err = os.MkdirAll(options.workspace+"/.terraform", 0700)
	if err != nil {
		return errors.WrapIf(err, "failed to create state backend directory")
	}

	stateFile, err := os.Create(options.workspace + "/.terraform/terraform.tfstate")
	if err != nil {
		return errors.WrapIf(err, "failed to create state config file")
	}
	_ = stateFile.Close()

	if err := runTerraform("init", options, env); err != nil {
		return errors.WrapIf(err, "failed to init state backend")
	}

	// remove old state.tf if any
	_ = os.Remove(options.workspace + "/state.tf")

	return nil
}
