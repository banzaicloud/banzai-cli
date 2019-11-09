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
	"os"
	"path/filepath"
	"runtime"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	providerK8s      = "k8s"
	providerEc2      = "ec2"
	providerKind     = "kind"
	providerPke      = "pke"
	providerEks      = "eks"
	defaultLocalhost = "default.localhost.banzaicloud.io"
	autoHost         = "auto"
	externalHost     = "externalHost"
)

type initOptions struct {
	file     string
	provider string
	*cpContext
}

func newInitOptions(cmd *cobra.Command, banzaiCli cli.Cli) *initOptions {
	cp := NewContext(cmd, banzaiCli)
	options := initOptions{cpContext: cp}

	flags := cmd.Flags()
	flags.StringVarP(&options.file, "file", "f", "", "Input Banzai Cloud Pipeline instance descriptor file")
	flags.StringVar(&options.provider, "provider", "", "Provider of the infrastructure for the deployment (k8s|kind|ec2|pke)")
	return &options
}

const initLongDescription = `

Depending on the --provider selection, the installer will work in the current Kubernetes context (k8s), deploy a KIND (Kubernetes in Docker) cluster to the local machine (kind), install a single-node PKE cluster (pke), or deploy a PKE cluster in Amazon EC2 (ec2).

The directory specified with --workspace, set in the installer.workspace key of the config, or $BANZAI_INSTALLER_WORKSPACE (default: ~/.banzai/pipeline/default) will be used for storing the applied configuration and deployment status.

The command requires docker or ctr (containerd) to be accessible in the system and able to run containers.

The input file will be copied to the workspace during initialization. Further changes can be done there before re-running the command (without --file).`

// NewInitCommand creates a new cobra.Command for `banzai pipeline init`.
func NewInitCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration for Banzai Cloud Pipeline",
		Long:  `Prepare a workspace for the deployment of an instance of Banzai Cloud Pipeline based on a values file or an interactive session.` + initLongDescription,
		Args:  cobra.NoArgs,
		PostRun: func(*cobra.Command, []string) {
			log.Info("Successfully initialized workspace. You can run now `banzai pipeline up` to deploy Pipeline.")
		},
	}

	options := newInitOptions(cmd, banzaiCli)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return runInit(*options, banzaiCli)
	}

	return cmd
}

func askProvider(k8sContext string) (string, error) {
	choices := []string{"Create single-node cluster in Amazon EC2"}
	lookup := []string{providerEc2}

	if k8sContext != "" {
		choices = append(choices, fmt.Sprintf("Use %q Kubernetes context", k8sContext))
		lookup = append(lookup, providerK8s)
	}

	choices = append(choices, "Create Amazon EKS cluster")
	lookup = append(lookup, providerEks)

	if hasTool("docker") == nil {
		choices = append(choices, "Create KIND (Kubernetes in Docker) cluster locally")
		lookup = append(lookup, providerKind)
	} else if checkPKESupported() == nil && checkRoot() == nil {
		choices = append(choices, "Install single-node PKE cluster")
		lookup = append(lookup, providerPke)
	}

	var provider int
	if err := survey.AskOne(&survey.Select{Message: "Select provider:", Options: choices}, &provider); err != nil {
		return "", err
	}

	return lookup[provider], nil
}

func runInit(options initOptions, banzaiCli cli.Cli) error {
	if err := options.Init(); err != nil {
		return err
	}

	if options.valuesExists() {
		log.Info("You can create another workspace with --workspace, or run `banzai pipeline up` to deploy the current one.")
		return errors.Errorf("workspace is already initialized in %q", options.workspace)
	}

	out := make(map[string]interface{})

	if !banzaiCli.Interactive() || options.file != "" {
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return errors.WrapIff(err, "failed to read file %q", filename)
		}

		err = utils.Unmarshal(raw, &out)
		if err != nil {
			return errors.WrapIf(err, "failed to parse descriptor")
		}

		if provider, ok := out["provider"].(string); ok && provider != "" {
			options.provider = provider
		}
	}

	// add defaults to values in case of missing values file
	if options.file == "" {
		out["tlsInsecure"] = true
		out["defaultStorageBackend"] = "postgres"
	}

	k8sContext, k8sConfig, err := input.GetCurrentKubecontext()
	if err != nil {
		if options.provider == providerK8s {
			return errors.WrapIf(err, "failed to use current Kubernetes context")
		}
		k8sContext = ""
		log.Debugf("won't use local kubernetes context: %v", err)
	} else if runtime.GOOS != "linux" {
		// non-native docker daemons can't access the host machine directly even if running in host networking mode
		// we have to rewrite configs referring to localhost to use the special name `host.docker.internal` instead
		k8sConfig, err = input.RewriteLocalhostToHostDockerInternal(k8sConfig)
		if err != nil {
			return errors.WrapIf(err, "failed to rewrite Kubernetes config")
		}
	}

	if provider, ok := out["provider"]; ok {
		if providerStr, ok := provider.(string); ok {
			options.provider = providerStr
		}
	}

	if options.provider == "" {
		if !banzaiCli.Interactive() {
			return errors.New("please select provider (--provider or provider field of values file)")
		}
		provider, err := askProvider(k8sContext)
		if err != nil {
			return err
		}

		options.provider = provider
	}

	out["provider"] = options.provider
	providerConfig := make(map[string]interface{})
	if pc, ok := out["providerConfig"]; ok {
		if pc, ok := pc.(map[string]interface{}); ok {
			providerConfig = pc
		}
	}

	uuID := uuid.New().String()
	if uuidValue, ok := out["uuid"].(string); ok && uuidValue != "" {
		uuID = uuidValue
	} else {
		out["uuid"] = uuID
	}

	out["ingressHostPort"] = options.provider != providerK8s

	switch options.provider {
	case providerKind:
		out[externalHost] = defaultLocalhost
		// TODO check if it resolves to 127.0.0.1 (user's dns recursor may drop this)
	case providerEc2:
		id, region, _, err := input.GetAmazonCredentialsRegion("")
		if err != nil {
			id, region, _, err = input.GetAmazonCredentialsRegion(defaultAwsRegion)
			if err != nil {
				log.Info("Please set your AWS credentials using aws-cli. See https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration")
				return errors.WrapIf(err, "failed to use local AWS credentials")
			} else {
				log.Infof("Using AWS region: %q", region)
			}
		}
		providerConfig["region"] = region
		providerConfig["accessKey"] = id
		hostname, _ := os.Hostname()
		providerConfig["tags"] = map[string]string{
			"banzaicloud-pipeline-controlplane-uuid": uuID,
			"local-id":                               fmt.Sprintf("%s@%s/%s", os.Getenv("USER"), hostname, filepath.Base(options.workspace)),
		}

		var confirmed bool
		_ = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Do you want to use the following AWS access key: %s?", id)}, &confirmed)
		if !confirmed {
			return errors.New("cancelled")
		}

		if out[externalHost] == nil {
			out[externalHost] = autoHost // address of ec2 instance
		}
	case providerK8s:
		if err := options.writeKubeconfig(k8sConfig); err != nil {
			return err
		}
		if out[externalHost] == nil {
			out[externalHost] = autoHost // address of lb service
		}

	case providerPke:
		out[externalHost] = guessExternalAddr()

	case providerEks:
		id, region, _, err := input.GetAmazonCredentialsRegion("")
		if err != nil {
			id, region, _, err = input.GetAmazonCredentialsRegion(defaultAwsRegion)
			if err != nil {
				log.Info("Please set your AWS credentials using aws-cli. See https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration")
				return errors.WrapIf(err, "failed to use local AWS credentials")
			} else {
				log.Infof("Using AWS region: %q", region)
			}
		}
		providerConfig["cluster_name"] = "poke-banzaicli-eks"
		providerConfig["instance_type"] = "m4.large"
		providerConfig["eks_service_role"] = "eksServiceRole"
		providerConfig["eks_node_role"] = "eksNodeRole"
		providerConfig["max_size"] = 3
		providerConfig["min_size"] = 2
		providerConfig["desired_capacity"] = 3
		providerConfig["region"] = "eu-central-1"
		providerConfig["accessKey"] = id
		hostname, _ := os.Hostname()
		providerConfig["tags"] = map[string]string{
			"banzaicloud-pipeline-controlplane-uuid": uuID,
			"local-id":                               fmt.Sprintf("%s@%s/%s", os.Getenv("USER"), hostname, filepath.Base(options.workspace)),
		}

		var confirmed bool
		_ = survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Do you want to use the following AWS access key: %s?", id)}, &confirmed)
		if !confirmed {
			return errors.New("cancelled")
		}
	}

	out["providerConfig"] = providerConfig

	return options.writeValues(out)
}
