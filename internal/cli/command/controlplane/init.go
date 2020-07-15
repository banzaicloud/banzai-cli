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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	providerK8s       = "k8s"
	providerEc2       = "ec2"
	providerKind      = "kind"
	providerPke       = "pke"
	providerCustom    = "custom"
	defaultLocalhost  = "default.localhost.banzaicloud.io"
	defaultWorkspace  = "default"
	autoHost          = "auto"
	externalHost      = "externalHost"
	localStateBackend = `{
	"terraform": {
		"backend": {
			"local": {
				"path": "/workspace/%s",
				"workspace_dir": "/workspace/"
			}
		}
	}
}`
	exportPath   = "/export"
	metadataFile = "export/metadata.yaml"
)

type ImageMetadata struct {
	Custom struct {
		CredentialType      string `yaml:"credentialType,omitempty"`
		Enabled             bool   `yaml:"enabled"`
		GenerateClusterName bool   `yaml:"generateClusterName"`
	}
}

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
	}

	options := newInitOptions(cmd, banzaiCli)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return runInit(*options, banzaiCli)
	}

	// print this only if init is not run as part of the `banzai pipeline up` command
	cmd.PostRun = func(*cobra.Command, []string) {
		upArgs := []string{"banzai", "pipeline", "up"}
		if options.workspace != defaultWorkspace {
			upArgs = append(upArgs, fmt.Sprintf("--workspace=%q", options.workspace))
		}
		if !options.pullInstaller {
			upArgs = append(upArgs, "--image-pull=false")
		}
		log.Infof("Successfully initialized workspace. "+
			"You can now edit the values file at %q and run `%s` to deploy Pipeline.", options.valuesPath(), strings.Join(upArgs, " "))
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

	if hasTool("docker") == nil {
		choices = append(choices, "Create KIND (Kubernetes in Docker) cluster locally")
		lookup = append(lookup, providerKind)
	} else if checkPKESupported() == nil && checkRoot() == nil {
		choices = append(choices, "Install single-node PKE cluster")
		lookup = append(lookup, providerPke)
	}

	choices = append(choices, "Create custom infrastructure (subscription neeeded)")
	lookup = append(lookup, providerCustom)

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
		upArgs := []string{"banzai", "pipeline", "up"}
		if options.workspace != defaultWorkspace {
			upArgs = append(upArgs, fmt.Sprintf("--workspace=%q", options.workspace))
		}
		if !options.pullInstaller {
			upArgs = append(upArgs, "--image-pull=false")
		}
		log.Infof("You can create another workspace with --workspace, "+
			"or run `%s` to deploy the current one.", strings.Join(upArgs, " "))
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
		} else if out == nil {
			log.Info("no configuration provided on stdin")
			out = make(map[string]interface{})
		}

		if provider, ok := out["provider"].(string); ok && provider != "" {
			options.provider = provider
		}
	}

	// add defaults to values in case of missing values file
	if options.file == "" {
		out["tlsInsecure"] = true
		out["defaultStorageBackend"] = "mysql"
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

	installer := make(map[string]interface{})
	if inst, ok := out["installer"]; ok {
		if inst, ok := inst.(map[string]interface{}); ok {
			installer = inst
		}
	}

	uuID := uuid.New().String()
	if uuidValue, ok := out["uuid"].(string); ok && uuidValue != "" {
		uuID = uuidValue
	} else {
		out["uuid"] = uuID
	}

	out["ingressHostPort"] = true
	hostname, _ := os.Hostname()

	options.installerImageRepo, options.installerTag = initImageValues(options, out)
	switch options.provider {
	case providerKind:
		out[externalHost] = defaultLocalhost
		// TODO check if it resolves to 127.0.0.1 (user's dns recursor may drop this)
	case providerEc2:
		assumeRole := cast.ToString(providerConfig["assume_role"])
		id, region, err := getAmazonCredentialsRegion(os.Getenv("AWS_PROFILE"), defaultAwsRegion, assumeRole)
		if err != nil {
			return err
		}
		providerConfig["region"] = region
		providerConfig["tags"] = map[string]string{
			"banzaicloud-pipeline-controlplane-uuid": uuID,
			"local-id":                               fmt.Sprintf("%s@%s/%s", os.Getenv("USER"), hostname, filepath.Base(options.workspace)),
		}

		log.Infof("The following AWS key will be used: %v", id)

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
		out["ingressHostPort"] = false

	case providerPke:
		out[externalHost] = guessExternalAddr()

	case providerCustom:
		hasExports, err := imageFileExists(options.cpContext, exportPath)
		if err != nil {
			return err
		}

		if !hasExports {
			return errors.New("The provided custom image has no metadata")
		}

		imageMeta := &ImageMetadata{}
		exportHandlers := []ExportedFilesHandler{
			metadataExporter(metadataFile, imageMeta),
		}

		if err := processExports(options.cpContext, exportPath, exportHandlers); err != nil {
			return err
		}

		switch imageMeta.Custom.CredentialType {
		case "aws":
			assumeRole := cast.ToString(providerConfig["assume_role"])
			id, region, err := getAmazonCredentialsRegion(os.Getenv("AWS_PROFILE"), defaultAwsRegion, assumeRole)
			if err != nil {
				return err
			}
			providerConfig["region"] = region

			log.Infof("The following AWS key will be used: %v", id)
		}

		out["ingressHostPort"] = false
		providerConfig["tags"] = map[string]string{
			"banzaicloud-pipeline-controlplane-uuid": uuID,
			"local-id":                               fmt.Sprintf("%s@%s/%s", os.Getenv("USER"), hostname, filepath.Base(options.workspace)),
		}

		if banzaiCli.Interactive() {
			autoApprove := false
			if err := survey.AskOne(&survey.Confirm{
				Message: "Do you want to automatically approve changes when executing `banzai pipeline up`?",
				Help:    "If you choose No, you will get a chance to review and approve the actual changes before executing them.",
			}, &autoApprove); err == nil {
				installer["autoApprove"] = autoApprove
			}
		}

		if imageMeta.Custom.GenerateClusterName {
			user := os.Getenv("USER")
			if user == "" {
				user = strings.Split(hostname, ".")[0]
			}
			name := fmt.Sprintf("banzai-%s-%s", strings.ToLower(user), filepath.Base(options.workspace))
			if banzaiCli.Interactive() {
				if err := survey.AskOne(&survey.Input{
					Message: "Name of cluster to create:",
					Default: name,
				}, &name); err == nil {
					providerConfig["cluster_name"] = name
				}
			}
		}

		if options.provider == providerCustom {
			if hasExports && !imageMeta.Custom.Enabled || !hasExports && options.installerImageRepo == defaultImage {
				return errors.New("Custom provisioning is available by specifying a custom installer image. Please refer to your deployment guide or use one of our support channels.")
			}
		}
	}

	if len(providerConfig) > 0 {
		out["providerConfig"] = providerConfig
	}

	err = options.ensureImagePulled()
	if err != nil {
		return errors.WrapIf(err, "failed to pull installer image")
	}
	if options.installerTag == latestTag && options.containerRuntime == runtimeDocker {
		ref, err := exec.Command("docker", "inspect", "-f", "{{index .RepoDigests 0}}", options.installerImage()).Output()
		if err != nil {
			return errors.WrapIf(err, "failed to determine installer image hash")
		}
		installer["image"] = strings.TrimSpace(string(ref))
	} else {
		installer["image"] = options.installerImage()
	}

	out["installer"] = installer

	return options.writeValues(out)
}

func initImageValues(options initOptions, out map[string]interface{}) (image string, tag string) {
	installer, ok := out["installer"].(map[interface{}]interface{})
	image = options.installerImageRepo
	tag = options.installerTag
	if options.installerImageRepo != defaultImage {
		if t := strings.Split(options.installerImageRepo, ":"); len(t) > 1 {
			image = t[0]
			tag = t[1]
		} else {
			image = options.installerImageRepo
		}
	} else {
		if ok {
			if img, ok := installer["image"].(string); ok {
				image = img
			}
		}
	}
	if options.installerTag != latestTag {
		tag = options.installerTag
	} else {
		if ok {
			if t, ok := installer["image-tag"].(string); ok {
				tag = t
			}
		}
	}
	return image, tag
}

func getAmazonCredentialsRegion(profile string, defaultAwsRegion string, assumeRole string) (string, string, error) {
	id, region, _, err := input.GetAmazonCredentialsRegion(profile, "", assumeRole)
	if err != nil {
		id, region, _, err = input.GetAmazonCredentialsRegion(profile, defaultAwsRegion, assumeRole)
		if err != nil {
			log.Info("Please set your AWS credentials using aws-cli. See https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration")
			return "", "", errors.WrapIf(err, "failed to use local AWS credentials")
		} else {
			log.Infof("Using AWS region: %q", region)
		}
	}
	return id, region, err
}

func metadataExporter(source string, metadata *ImageMetadata) ExportedFilesHandler {
	return ExportedFilesHandler(func(files map[string][]byte) error {
		if valuesFileContent, ok := files[source]; ok {
			if err := yaml.Unmarshal(valuesFileContent, metadata); err != nil {
				return errors.Wrap(err, "failed to unmarshal metadata values exported from the image")
			}
		}
		return nil
	})
}
