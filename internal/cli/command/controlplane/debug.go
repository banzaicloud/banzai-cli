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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type debugOptions struct {
	outputFile string
	*cpContext
}

func (d *debugOptions) Init() error {
	if err := d.cpContext.Init(); err != nil {
		return err
	}
	return d.setOutputFilename()
}

func (d *debugOptions) setOutputFilename() error {
	value := d.outputFile

	base := d.cpContext.workspace
	if filepath.IsAbs(value) {
		base = "/"
	} else if strings.HasPrefix(value, "./") {
		base, _ = os.Getwd()
	}
	value = filepath.Clean(filepath.Join(base, value))

	stat, err := os.Stat(value)
	if err == nil {
		if stat.IsDir() {
			value = filepath.Join(
				value,
				fmt.Sprintf("pipeline-debug-bundle-%s.tgz", time.Now().Format("20060102-1504")))
		} else {
			return errors.New(fmt.Sprintf("output file named %q already exists", value))
		}
	}
	d.outputFile = filepath.Clean(value)
	return nil
}

// NewDebugCommand creates a new cobra.Command for `banzai pipeline debug`.
func NewDebugCommand(banzaiCli cli.Cli) *cobra.Command {
	options := debugOptions{}

	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Generate debug bundle",
		Long: `Collect and package status information and configuration data required by the Banzai Cloud support team to resolve issues on a remote Pipeline installation.

The command will generate a tar archive, which contains information about the
machine used for creating the support bundle, the configuration values and the
deployment logs from the workspace, and the description and logs of relevant
resources on the cluster.

The following data is included:

* Timestamp of the debug bundle
* Banzai CLI version number
* Local file system path to the Banzai Pipeline workspace
* List of other workspaces managed on the local machine
* List of docker images available on the local machine
* Version of the local Docker daemon
* Basic local machine information (hostname, uptime, boot time, number of processes, operating system type and version, virtualization, host identifier).
* The values.yaml file from the workspace (should not contain sensitive data)
* List of all files in the workspace
* Logs (output) of all previous terraform runs issued by the Banzai CLI in the specific local workspace
* Logs (output) of terraform plan, graph and state list
* Full YAML description of the following resource kinds in the banzaicloud namespace of the Pipeline cluster: pods, services, ingresses, persistentvolumes, persistentvolumeclaims, events, clusterflows, clusteroutputs, flows, loggings, outputs.
* List of secrets and configmaps in the banzaicloud namespace (without the content)
* Output of helm list in the banzaicloud namespace.
* Logs of all pods and containers in the banzaicloud namespace
* Log of the support bundle execution`,
		Example: `banzai pipeline debug --workspace prod
banzai pipeline debug --output-file ./`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runDebug(options, banzaiCli)
		},
	}

	options.cpContext = NewContext(cmd, banzaiCli)
	flags := cmd.Flags()
	flags.StringVarP(&options.outputFile, "output-file", "O", "", "Name or path to output file relative to the workspace (prefix with ./ for current working directory; default: pipeline-debug-bundle-DATE.tgz)")

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

type loggerHook struct {
	*log.Logger
}

func (l loggerHook) Fire(e *log.Entry) error {
	l.Log(e.Level, e.Message)
	return nil
}

func (loggerHook) Levels() []log.Level {
	return log.AllLevels
}

func runDebug(options debugOptions, banzaiCli cli.Cli) error {
	logger := log.StandardLogger()    // TODO add logger to cli.Cli
	logHandler := errorLogger{logger} // error handler for best-effort operations

	logBuffer := new(bytes.Buffer)
	bufferLogger := log.New()
	bufferLogger.SetOutput(logBuffer)
	logger.AddHook(loggerHook{bufferLogger})

	if err := options.Init(); err != nil {
		return err
	}

	if !options.valuesExists() {
		return errors.New(fmt.Sprintf("%q is not an initialized workspace (no values file found)", options.workspace))
	}

	logger.Debugf("creating %q", options.outputFile)
	f, err := os.Create(options.outputFile)
	if err != nil {
		return errors.WrapIff(err, "failed to create archive at %q", options.outputFile)
	}
	defer func() { logHandler.Handle(f.Close()) }()

	gzWriter := gzip.NewWriter(f)
	defer func() { logHandler.Handle(gzWriter.Close()) }()

	tm := newTarManager(tar.NewWriter(gzWriter), logHandler, logger, "pipeline-debug-bundle")
	defer tm.Close()

	meta := debugMetadata{
		Timestamp:       time.Now(),
		CLIVersion:      banzaiCli.Version(),
		WorkspacePath:   options.workspace,
		Workspaces:      strings.Split(simpleCommand("ls ~/.banzai/pipeline"), "\n"),
		InstallerImages: strings.Split(simpleCommand("docker image list | grep pipeline-installer"), "\n"),
		DockerVersion:   simpleCommand("docker --version"),
	}

	if info, err := host.Info(); err == nil {
		meta.Host = *info
	}

	metaBytes, _ := yaml.Marshal(meta)
	tm.AddFile("meta.yaml", metaBytes)

	tm.CopyFile("pipeline/values.yaml", options.valuesPath())
	tm.AddFile("pipeline/files.txt", simpleCommand("find", options.workspace, "-ls"))

	// run some terraform diagnostics commands to catch their output in the logs folder
	var values map[string]interface{}
	logHandler.Handle(options.readValues(&values))
	_, env, err := getImageMetadata(options.cpContext, values, true)
	logHandler.Handle(err)
	logHandler.Handle(runTerraform("plan", options.cpContext, env))
	logHandler.Handle(runTerraform("graph", options.cpContext, env))
	logHandler.Handle(runTerraform("state list", options.cpContext, env))

	tm.Cd("/pipeline/installer-logs")
	logDir, logFiles, err := options.listLogs()
	if err != nil {
		logHandler.Handle(errors.WrapIf(err, "listing log files failed"))
	} else {
		for _, file := range logFiles {
			tm.CopyFile(file, filepath.Join(logDir, file))
		}
	}

	if !options.kubeconfigExists() {
		logger.Errorf("No kubeconfig found in the workspace. This means that the debug bundle will contain very limited information. If the cluster is running, please create the support bundle from a workspace where `banzai pipeline up` has been run.")
	} else {
		tm.Cd("/pipeline/resources/banzaicloud")
		for _, resource := range []string{
			"pods", "services", "ingresses", "persistentvolumes", "persistentvolumeclaims", "events",
			"clusterflows,clusteroutputs,flows,loggings,outputs",
		} {
			tm.AddFile(resource+".yaml", combineOutput(runContainerCommand(options.cpContext, []string{"kubectl", "get", resource, "-oyaml", "-nbanzaicloud"}, env)))
		}
		for _, resource := range []string{"secrets", "configmaps"} {
			tm.AddFile(resource+".txt", combineOutput(runContainerCommand(options.cpContext, []string{"kubectl", "get", resource, "-owide", "-nbanzaicloud"}, env)))
		}

		tm.AddFile("helm_list.txt", combineOutput(runContainerCommand(options.cpContext, []string{"helm", "list", "--namespace", "banzaicloud", "--all"}, env)))

		tm.Cd("/pipeline/logs/banzaicloud")
		pods, err := runContainerCommand(options.cpContext, []string{"kubectl", "get", "pods", "-oname", "-nbanzaicloud"}, env)
		if err != nil {
			logHandler.Handle(errors.WrapIf(err, "failed to list pods"))
		} else {
			for _, pod := range strings.Split(strings.TrimSpace(pods), "\n") {
				pod = strings.TrimPrefix(strings.TrimSpace(pod), "pod/")
				tm.AddFile(fmt.Sprintf("%s.log", pod), combineOutput(runContainerCommand(options.cpContext, []string{"kubectl", "logs", "--namespace", "banzaicloud", pod, "--all-containers"}, env)))
			}
		}
	}

	tm.Cd("/")
	tm.AddFile("meta.log", logBuffer)

	log.Infof("debug bundle has been written to %q", options.outputFile)
	log.Infof("You may want to encrypt the archive with a command like `gpg --encrypt -r 37E2B4AEBEB1F45B %q`, and send %q to the Banzai Cloud Support Team. Ask for the public key on a trusted channel.", options.outputFile, options.outputFile+".gpg")
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
	return combineOutput(output, err)
}

type errorHandler interface {
	Handle(err error)
}

type errorLogger struct {
	logger
}

func (e errorLogger) Handle(err error) {
	if err != nil {
		e.logger.Error(err.Error())
	}
}

type logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

type tarManager struct {
	tarWriter    *tar.Writer
	errorHandler errorHandler
	logger       logger
	directories  map[string]bool
	baseDir      string
	cwd          string
	dirPerm      int64
	filePerm     int64
	time         func() time.Time
}

func newTarManager(tarWriter *tar.Writer, errorHandler errorHandler, logger logger, baseDir string) tarManager {
	tm := tarManager{
		tarWriter:    tarWriter,
		errorHandler: errorHandler,
		logger:       logger,
		directories:  make(map[string]bool),
		baseDir:      filepath.Clean(filepath.Join("/", baseDir)),
		dirPerm:      0777,
		filePerm:     0666,
		time:         time.Now,
	}
	tm.Cd("/")
	return tm
}

// Cd changes the working directory to the given folder in the tar archive.
// Directories are created implicitly. You can give relative path to the
// current working directory, or an absolute path starting with a /
// (where / refers to the baseDir).
func (t *tarManager) Cd(dir string) {
	path := t.path(dir)
	t.Mkdir(path)
	t.cwd = path
}

func (t *tarManager) path(dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Clean(filepath.Join(t.cwd, dir))
}

// Mkdir creates the given directory and its parents when any of them are missing
func (t *tarManager) Mkdir(dir string) {
	path := t.path(dir)
	if t.directories[path] {
		return
	}
	parent := filepath.Clean(filepath.Join(dir, ".."))
	if parent != "/" {
		t.Mkdir(parent)
	}
	t.addDir(path)
	t.directories[path] = true
}

func (t tarManager) addDir(name string) {
	// trailing slash makes this a directory, relative paths are normally preferred in the archive
	name = strings.TrimLeft(filepath.Clean(filepath.Join(t.baseDir, name))+"/", "/")
	err := t.tarWriter.WriteHeader(&tar.Header{Name: name, Mode: t.dirPerm, ModTime: t.time(), ChangeTime: t.time()})
	if err != nil {
		t.errorHandler.Handle(errors.WrapIf(err, "failed to create directory in archive"))
	} else {
		t.logger.Debugf("added directory %q", name)
	}
}

func (t tarManager) addFile(name, content string) (err error) {
	// relative paths are normally preferred in the archive
	name = strings.TrimLeft(filepath.Clean(filepath.Join(t.baseDir, t.path(name))), "/")
	bytes := []byte(content)
	err = t.tarWriter.WriteHeader(&tar.Header{Name: name, Mode: t.filePerm, ModTime: t.time(), ChangeTime: t.time(), Size: int64(len(bytes))})
	if err != nil {
		return err
	}

	n, err := t.tarWriter.Write(bytes)
	t.logger.Debugf("%d bytes written to %q", n, name)
	return err
}

// AddFile adds the file with the given name and content to the archive. The
// name is relative to the working directory, but it can also be an
// absolute path (starting from the base directory).
// Directories are created implicitly.
func (t *tarManager) AddFile(name string, content interface{}) {
	path := t.path(name)
	dir, _ := filepath.Split(path)
	t.Mkdir(dir)

	var str string
	switch v := content.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	case interface{ String() string }:
		str = v.String()
	default:
		str = fmt.Sprint(v)
	}
	err := t.addFile(path, str)
	if err != nil {
		t.errorHandler.Handle(errors.WrapIff(err, "failed to add %q", path))
	} else {
		t.logger.Infof("file added to archive: %q", path)
	}
}

// CopyFile copies the contents of the local file at the given path as name to the archive.
func (t tarManager) CopyFile(name, path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.errorHandler.Handle(err)
	}

	t.AddFile(name, content)
}

func (t tarManager) Close() {
	err := t.tarWriter.Close()
	if err != nil {
		t.errorHandler.Handle(errors.Wrap(err, "closing tar archive"))
	} else {
		t.logger.Debug("closed tar archive")
	}
}
