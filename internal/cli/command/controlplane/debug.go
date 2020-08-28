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
	"compress/gzip"
	"context"
	"fmt"
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

	log.Infof("debug bundle has been written to %q", path)
	return nil
}

// simpleCommand runs the given shell command, and returns its outputs, an error message or both
func simpleCommand(command string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
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
		log.Errorf("%s failed: %v", err)
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
