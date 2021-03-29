// Copyright © 2020 Banzai Cloud
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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func dirExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if info.IsDir() {
		return true, nil
	}
	return false, errors.New("file exists but not a directory")
}

type ExportedFilesHandler func(map[string][]byte) error

func processExports(options *cpContext, source string, exportedFilesHandlers []ExportedFilesHandler) error {
	files, err := readFilesFromContainerToMemory(options, source)
	if err != nil {
		return errors.WrapIf(err, "failed to export files from the image")
	}

	for _, h := range exportedFilesHandlers {
		if err := h(files); err != nil {
			return errors.WrapIf(err, "failed to run handler on exported files")
		}
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
	if err := ioutil.WriteFile(filepath.Join(options.workspace, generatedValuesFileName), bytes, 0600); err != nil {
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

func stringifyMap(m interface{}) interface{} {
	switch v := m.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{})
		for k, v := range v {
			out[k] = stringifyMap(v)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]interface{})
		for k, v := range v {
			out[fmt.Sprint(k)] = stringifyMap(v)
		}
		return out
	default:
		return v
	}
}

func defaultValuesExporter(source string, defaultValues *map[string]interface{}) ExportedFilesHandler {
	return ExportedFilesHandler(func(files map[string][]byte) error {
		if valuesFileContent, ok := files[source]; ok {
			if err := yaml.Unmarshal(valuesFileContent, defaultValues); err != nil {
				return errors.Wrap(err, "failed to unmarshal default values exported from the image")
			}
		}
		return nil
	})
}

func getImageMetadata(cpContext *cpContext, values map[string]interface{}, writeValues bool) (string, map[string]string, error) {
	var defaultValues map[string]interface{}
	exportHandlers := []ExportedFilesHandler{
		defaultValuesExporter("export/values.yaml", &defaultValues),
	}

	env := make(map[string]string)
	var imageMeta ImageMetadata
	if values["provider"] == providerCustom {
		log.Debug("parsing metadata")
		exportHandlers = append(exportHandlers, metadataExporter(metadataFile, &imageMeta))
	}

	if err := processExports(cpContext, exportPath, exportHandlers); err != nil {
		return "", env, err
	}

	log.Debugf("custom image metadata: %+v", imageMeta)

	if writeValues {
		if err := writeMergedValues(cpContext, defaultValues, values); err != nil {
			return "", env, err
		}
	}

	awsAccessKeyID := ""
	if values["provider"] == providerEc2 || values["passAWSCredentials"] == true || imageMeta.Custom.CredentialType == "aws" {
		profile := ""
		assumeRole := ""
		if v, ok := values["providerConfig"]; ok {
			providerConfig := cast.ToStringMap(v)
			profile = cast.ToString(providerConfig["profile"])
			assumeRole = cast.ToString(providerConfig["assume_role"])
		}
		if envProfile, ok := os.LookupEnv("AWS_PROFILE"); ok {
			if profile != "" {
				log.Warnf("AWS profile `%s` in the providerConfig is overridden to `%s` by AWS_PROFILE env var explicitly", profile, envProfile)
			}
			profile = envProfile
		}
		log.Debug("using local AWS credentials")
		id, creds, err := input.GetAmazonCredentials(profile, assumeRole)
		if err != nil {
			return "", env, errors.WrapIf(err, "failed to get AWS credentials")
		}
		env = creds
		awsAccessKeyID = id
	}
	return awsAccessKeyID, env, nil
}
