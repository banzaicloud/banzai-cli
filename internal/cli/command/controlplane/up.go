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
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/login"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
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

			return runUp(options, banzaiCli)
		},
	}

	options.initOptions = newInitOptions(cmd, banzaiCli)

	flags := cmd.Flags()
	flags.BoolVarP(&options.init, "init", "i", false, "Initialize workspace")

	return cmd
}

func printExternalHostRecord(host string) error {
	c := exec.Command("kubectl", "get", "services", "-n", "banzaicloud", "traefik", "-o", "jsonpath='{.status.loadBalancer.ingress[0].*}")
	out, err := c.Output()
	if err != nil {
		return errors.WrapIf(err, "failed to determine address of the load balancer")
	}

	lbAddress := strings.Trim(string(out), "'\n ") // TODO get from tf output
	lbRecordType := "A"
	if net.ParseIP(lbAddress) == nil {
		lbRecordType = "CNAME"
	}

	fmt.Printf("Please create a DNS record pointing to the load balancer:\n\n%s IN %s %s\n", host, lbRecordType, lbAddress)
	return nil
}

func runUp(options createOptions, banzaiCli cli.Cli) error {
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

	if options.pullInstaller {
		log.Info("Pulling Banzai Cloud Pipeline installer image...")
		if err := options.pullDockerImage(); err != nil {
			return errors.WrapIf(err, "failed to pull cp-installer")
		}
	}

	externalHost, _ := values["externalHostSource"].(string)
	var env map[string]string
	switch values["provider"] {
	case providerKind:
		err := ensureKINDCluster(banzaiCli, *options.cpContext)
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

		if err := ensureEC2Cluster(banzaiCli, *options.cpContext, creds, useGeneratedKey); err != nil {
			return errors.WrapIf(err, "failed to create EC2 cluster")
		}
		env = creds

	default:
		if !options.kubeconfigExists() {
			return errors.New("could not find Kubeconfig in workspace")
		}
	}

	log.Info("Deploying Banzai Cloud Pipeline to Kubernetes cluster...")
	if err := runInternal("apply", *options.cpContext, env); err != nil {
		return errors.WrapIf(err, "failed to deploy pipeline components")
	}

	url, err := options.readAddress()
	if err != nil {
		return errors.WrapIf(err, "can't read final URL of Pipeline")
	}
	log.Infof("Pipeline is ready at %s.", url)
	url += "pipeline"

	source, ok := values["externalHostSource"].(string)
	if !ok || source == "" {
	}

	if externalHost != "auto" && externalHost != defaultLocalhost {
		err := printExternalHostRecord(externalHost)
		if err != nil {
			log.Errorf("%v", err)
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

func runInternal(command string, options cpContext, env map[string]string, targets ...string) error {
	cmdEnv := map[string]string{"KUBECONFIG": "/root/" + kubeconfigFilename}
	for k, v := range env {
		cmdEnv[k] = v
	}

	cmd := []string{"terraform",
		command,
		"-parallelism=1"} // workaround for https://github.com/terraform-providers/terraform-provider-helm/issues/271

	if options.autoApprove {
		cmd = append(cmd, "-auto-approve")
	}

	for _, target := range targets {
		cmd = append(cmd, "-target", target)
	}

	return runInstaller(cmd, options, cmdEnv)
}

func runInstaller(command []string, options cpContext, env map[string]string) error {

	args := []string{
		"run", "-it", "--rm", "--net=host",
		"-v", fmt.Sprintf("%s:/root", options.workspace),
		"-e", fmt.Sprintf("KUBECONFIG=/root/%s", kubeconfigFilename),
	}

	envs := os.Environ()
	for key, value := range env {
		args = append(args, "-e", key)
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	args = append(append(append(args,
		fmt.Sprintf("banzaicloud/cp-installer:%s", options.installerTag)),
		command...),
		"-var", "workdir=/root",
		"-state=/root/"+tfstateFilename)

	log.Info("docker ", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)

	cmd.Env = envs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err
}
