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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const basePath = "pipeline-debug-bundle/"

type debugOptions struct {
	filename string
	*cpContext
}

func (d *debugOptions) filepath() string {
	if !d.cpContext.flags.Changed("filename") {
		d.filename = fmt.Sprintf("pipeline-debug-bundle-%s.tgz", time.Now().Format("20060102-1504"))
	}
	return filepath.Join(d.cpContext.workspace, d.filename)
}

// NewDebugCommand creates a new cobra.Command for `banzai pipeline debug`.
func NewDebugCommand(banzaiCli cli.Cli) *cobra.Command {
	options := debugOptions{}

	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Generate debug bundle",
		Long:  "Collect and package status information and configuration data that is needed by Banzai Cloud support team to resolve issues on a remote Pipeline installation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runDebug(options, banzaiCli)
		},
	}

	options.cpContext = NewContext(cmd, banzaiCli)
	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "filename", "f", "", "Name or path to output file relative to the workspace (prefix with ./ for current working directory; default: pipeline-debug-bundle-DATE.tgz)")

	return cmd
}

type debugMetadata struct {
	Timestamp       time.Time
	Host            host.InfoStat
	CLIVersion      string
	WorkspacePath   string
	Workspaces      []string
	InstallerImages []string
	DockerVersion   string
}

func runDebug(options debugOptions, banzaiCli cli.Cli) error {
	logBuffer := new(bytes.Buffer)
	log.SetOutput(io.MultiWriter(os.Stderr, logBuffer))

	if err := options.Init(); err != nil {
		return err
	}

	path := options.filepath()
	f, err := os.Create(path)
	if err != nil {
		return errors.WrapIff(err, "failed to create archive at %q", path)
	}
	defer f.Close()

	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	err = addDir(tarWriter, "")
	if err != nil {
		return errors.Wrapf(err, "adding bundle base directory at %s failed", basePath)
	}

	meta := debugMetadata{
		Timestamp:       time.Now(),
		CLIVersion:      banzaiCli.Version(),
		WorkspacePath:   options.workspace,
		Workspaces:      strings.Split(simpleCommand("ls ~/.banzai/pipeline"), "\n"),
		InstallerImages: strings.Split(simpleCommand("docker image list|grep pipeline-installer"), "\n"),
		DockerVersion:   simpleCommand("docker --version"),
	}

	if info, err := host.Info(); err == nil {
		meta.Host = *info
	}

	bytes, _ := yaml.Marshal(meta)
	logResult("add meta.yaml", addFile(tarWriter, "meta.yaml", string(bytes)))

	logResult("create pipeline/", addDir(tarWriter, "pipeline"))
	logResult("add values.yaml", copyFile(tarWriter, "pipeline/values.yaml", options.valuesPath()))

	// run some terraform diagnostics commands to catch their output in the logs folder
	var values map[string]interface{}
	logResult("read values", options.readValues(&values))
	_, env, err := getImageMetadata(options.cpContext, values, true)
	logResult("get image metadata", err)
	logResult("run tf plan", runTerraform("plan", options.cpContext, env))
	logResult("run tf graph", runTerraform("graph", options.cpContext, env))
	logResult("run tf state list", runTerraform("state list", options.cpContext, env))

	logResult("create pipeline/installer-logs", addDir(tarWriter, "pipeline/installer-logs"))
	logDir, logFiles, err := options.listLogs()
	if err != nil {
		log.Errorf("listing log files failed: %v", err)
	} else {
		for _, file := range logFiles {
			logResult("add log file", copyFile(tarWriter, filepath.Join("pipeline/installer-logs", file), filepath.Join(logDir, file)))
		}
	}

	logResult("add pipeline/files.txt", addFile(tarWriter, "pipeline/files.txt", simpleCommand("find", options.workspace, "-ls")))

	logResult("create pipeline/resources", addDir(tarWriter, "pipeline/resources"))
	logResult("add pipeline/resources/all.txt", addFile(tarWriter, "pipeline/resources/all.txt", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "all", "-A", "-owide"}, env)))
	logResult("add pipeline/resources/namespaces.yaml", addFile(tarWriter, "pipeline/resources/namespaces.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "ns", "-oyaml"}, env)))
	logResult("add pipeline/resources/nodes.yaml", addFile(tarWriter, "pipeline/resources/nodes.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "nodes", "-oyaml"}, env)))
	logResult("add pipeline/resources/top_node.txt", addFile(tarWriter, "pipeline/resources/top_node.txt", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "top", "node"}, env)))
	logResult("add pipeline/resources/webhooks.yaml", addFile(tarWriter, "pipeline/resources/webhooks.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "mutatingwebhookconfigurations,validatingwebhookconfigurations", "-oyaml"}, env)))

	logResult("create pipeline/resources/banzaicloud", addDir(tarWriter, "pipeline/resources/banzaicloud"))

	logResult("add pipeline/resources/banzaicloud/pods.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/pods.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "pods", "-oyaml", "-nbanzaicloud"}, env)))
	logResult("add pipeline/resources/banzaicloud/services.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/services.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "services", "-oyaml", "-nbanzaicloud"}, env)))
	logResult("add pipeline/resources/banzaicloud/ingresses.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/ingresses.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "ingresses", "-oyaml", "-nbanzaicloud"}, env)))
	logResult("add pipeline/resources/banzaicloud/persistentvolumes.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/persistentvolumes.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "persistentvolumes", "-oyaml", "-nbanzaicloud"}, env)))
	logResult("add pipeline/resources/banzaicloud/persistentvolumeclaims.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/persistentvolumeclaims.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "persistentvolumeclaims", "-oyaml", "-nbanzaicloud"}, env)))
	logResult("add pipeline/resources/banzaicloud/configmaps.txt", addFile(tarWriter, "pipeline/resources/banzaicloud/configmaps.txt", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "configmaps", "-owide", "-nbanzaicloud"}, env))) // -owide does not include contents
	logResult("add pipeline/resources/banzaicloud/secrets.txt", addFile(tarWriter, "pipeline/resources/banzaicloud/secrets.txt", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "secrets", "-owide", "-nbanzaicloud"}, env)))          // -owide does not include contents
	logResult("add pipeline/resources/banzaicloud/events.yaml", addFile(tarWriter, "pipeline/resources/banzaicloud/events.yaml", fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "events", "-oyaml", "-nbanzaicloud"}, env)))
	// TODO add helm binary to installer image
	logResult("add pipeline/resources/banzaicloud/helm_list.txt", addFile(tarWriter, "pipeline/resources/banzaicloud/helm_list.txt", fetchContainerCommandOutputAndError(options.cpContext, []string{"helm", "list", "--namespace", "banzaicloud", "--all"}, env)))

	logResult("create pipeline/logs", addDir(tarWriter, "pipeline/logs"))
	logResult("create pipeline/logs/banzaicloud", addDir(tarWriter, "pipeline/logs/banzaicloud"))
	pods := strings.Split(strings.TrimSpace(fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "get", "pods", "-oname", "-nbanzaicloud"}, env)), "\n")
	for _, pod := range pods {
		pod = strings.TrimPrefix(strings.TrimSpace(pod), "pod/")
		logResult("add pod log", addFile(tarWriter, filepath.Join("pipeline/logs/banzaicloud/", pod+".log"), fetchContainerCommandOutputAndError(options.cpContext, []string{"kubectl", "logs", "-nbanzaicloud", pod, "--all-containers"}, env)))
	}

	log.Infof("debug bundle has been written to %q", path)
	logResult("add meta.log", addFile(tarWriter, "meta.log", logBuffer.String()))

	return nil
}

// simpleCommand runs the given shell command, and returns its outputs, an error message or both
func simpleCommand(command string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if len(args) > 0 {
		cmd = exec.CommandContext(ctx, command, args...)
	}
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			output += "\n"
		}
		output += err.Error()
	}
	return output
}

func logResult(what string, err error) {
	if err == nil {
		log.Debugf("%s succeeded", what)
	} else {
		log.Errorf("%s failed: %v", what, err)
	}
}

func addDir(w *tar.Writer, name string) error {
	name = strings.TrimRight(filepath.Join(basePath, name), "/") + "/"
	err := w.WriteHeader(&tar.Header{Name: name, Mode: 0777})
	return errors.WrapIf(err, "failed to create directory in archive")
}

func addFile(w *tar.Writer, name, content string) (err error) {
	name = filepath.Join(basePath, name)
	bytes := []byte(content)
	err = w.WriteHeader(&tar.Header{Name: name, Mode: 0666, Size: int64(len(bytes))})
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(w *tar.Writer, name, path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return addFile(w, name, string(content))
}
