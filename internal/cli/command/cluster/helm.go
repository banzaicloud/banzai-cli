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

package cluster

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	serviceutils "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
)

type helmOptions struct {
	version string
}

func NewHelmCommand(banzaiCli cli.Cli) *cobra.Command {
	options := helmOptions{}

	cmd := &cobra.Command{
		Use:    "_helm",
		Hidden: true,
		Short:  "Wrapper to download and execute the Helm version matching the Tiller on the cluster of the current kubecontext, in a Helm home that is synchronized with Pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHelm(banzaiCli, options, args)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.version, "version", "v", "", "Helm version")

	return cmd
}

func writeHelm(url, name string) error {
	tgz, err := http.Get(url) // #nosec
	if err != nil {
		return errors.WrapIff(err, "failed to download helm from %q", url)
	}
	defer tgz.Body.Close()

	zr, err := gzip.NewReader(tgz.Body)
	if err != nil {
		return errors.WrapIf(err, "failed to uncompress helm archive")
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			return errors.WrapIf(err, "failed to extract helm from archive")
		}
		if filepath.Base(hdr.Name) == "helm" {
			tempName := name + "~"
			f, err := os.OpenFile(tempName, (os.O_WRONLY | os.O_CREATE | os.O_EXCL), 0755)
			if err != nil {
				return errors.WrapIf(err, "failed to create temporary file for helm binary")
			}

			_, err = io.Copy(f, tr)
			f.Close()

			if err != nil {
				return errors.WrapIf(err, "failed to write helm binary")
			}

			return errors.WrapIf(os.Rename(tempName, name), "failed to move helm binary to its final place")
		}
	}
}

func runHelm(banzaiCli cli.Cli, options helmOptions, args []string) error {
	env := os.Environ()
	envs := make(map[string]string, len(env))
	for _, pair := range env {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envs[parts[0]] = parts[1]
	}

	var version string
	var err error
	if options.version == "2" {
		version, err = tillerVersion()
		if err != nil {
			return err
		}
		envs, err = setHelm2Env(envs, banzaiCli)
		if err != nil {
			return err
		}
	} else {
		version, err = getHelmVersion(banzaiCli)
		if err != nil {
			return err
		}
		envs, err = setHelmEnv(envs, banzaiCli)
		if err != nil {
			return err
		}
	}

	name, err := getHelmBinary(version, banzaiCli)
	if err != nil {
		return err
	}

	env = make([]string, 0, len(envs))
	for k, v := range envs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	log.Debugf("Environment: %v", envs)
	return errors.WrapIf(syscall.Exec(name, append([]string{"helm"}, args...), env), "failed to exec helm")
}

// helmRepo is an item of the repositories config of Helm
type helmRepo struct {
	Name  string
	URL   string `yaml:"url"`
	Cache string
}

// helmRepo is the simplified structure of the repositories config of Helm
type helmRepos struct {
	ApiVersion   string `yaml:"apiVersion"`
	Repositories []helmRepo
	Generated    time.Time
}

func dumpRepositories(banzaiCli cli.Cli, reposdir string) error {
	filename := filepath.Join(reposdir, "repositories.yaml")
	if _, err := os.Stat(filename); err == nil {
		return nil
	}

	log.Infof("Creating Helm home for organization")
	org := banzaiCli.Context().OrganizationID()
	pipeline := banzaiCli.Client()
	repos, _, err := pipeline.HelmApi.HelmListRepos(context.Background(), org)
	if err != nil {
		return errors.WrapIf(err, "failed to get list of Helm repositories")
	}

	cachedir := filepath.Join(reposdir, "cache")
	if err := os.MkdirAll(cachedir, 0755); err != nil {
		return errors.WrapIf(err, "failed to create helm cache directory")
	}

	config := helmRepos{ApiVersion: "v1", Repositories: make([]helmRepo, len(repos)), Generated: time.Now()}
	empty, _ := yaml.Marshal(struct {
		Entries    map[string]interface{}
		ApiVersion string
		Generated  time.Time
	}{ApiVersion: "v1"})
	for i, repo := range repos {
		cache := filepath.Join(cachedir, fmt.Sprintf("%s-index.yaml", repo.Name)) // this is hardcoded in helm
		if _, err := os.Stat(cache); err != nil {
			err := ioutil.WriteFile(cache, empty, 0644)
			if err != nil {
				return errors.WrapIf(err, "failed to write initial repository config")
			}
		}
		config.Repositories[i] = helmRepo{Name: repo.Name, URL: repo.Url, Cache: cache}
	}

	content, err := yaml.Marshal(config)
	if err != nil {
		return errors.WrapIf(err, "failed to marshal Helm repositories list")
	}

	return errors.WrapIf(ioutil.WriteFile(filename, content, 0644), "failed to write repository list")
}

func tillerVersion() (string, error) {
	c := exec.Command("kubectl", "get", "deployment", "-n", "kube-system", "-o", "jsonpath={.items[0].spec.template.spec.containers[0].image}", "-l", "app=helm")
	out, err := c.Output()
	if err != nil {
		return "", errors.WrapIf(err, "failed to determine version of Tiller on the cluster")
	}
	parts := strings.Split(string(out), ":")
	if len(parts) != 2 {
		return "", errors.Errorf("failed to parse Tiller image name %q", out)
	}
	version := parts[1] // TODO check format
	return version, nil
}

func getHelmVersion(banzaiCli cli.Cli) (string, error) {
	caps, r, err := banzaiCli.Client().PipelineApi.ListCapabilities(context.Background())
	if err := serviceutils.CheckCallResults(r, err); err != nil {
		return "", errors.WrapIf(err, "failed to retrieve capabilities")
	}

	return fmt.Sprintf("v%s", caps["helm"]["version"]), nil
}

func getHelmBinary(version string, banzaiCli cli.Cli) (string, error) {
	bindir := filepath.Join(banzaiCli.Home(), "bin")
	if err := os.MkdirAll(bindir, 0755); err != nil {
		return "", errors.WrapIff(err, "failed to create %q directory", bindir)
	}

	url := fmt.Sprintf("https://get.helm.sh/helm-%s-%s-amd64.tar.gz", version, runtime.GOOS)
	name := filepath.Join(bindir, fmt.Sprintf("helm-%s", version))

	if _, err := os.Stat(name); err != nil {
		log.Infof("Downloading helm %s...", version)
		if runtime.GOARCH != "amd64" {
			return "", errors.Errorf("unsupported architecture: %v", runtime.GOARCH)
		}
		if err := writeHelm(url, name); err != nil {
			return "", errors.WrapIf(err, "failed to download helm client")
		}
		log.Infof("Helm %s downloaded successfully", version)
	}
	return name, nil
}

func setHelmEnv(envs map[string]string, banzaiCli cli.Cli) (map[string]string, error) {
	org := banzaiCli.Context().OrganizationID()
	helmDataHome := filepath.Join(banzaiCli.Home(), fmt.Sprintf("helm/org-%d/data", org))
	if err := os.MkdirAll(helmDataHome, 0755); err != nil {
		return envs, errors.WrapIff(err, "failed to create %q directory", helmDataHome)
	}
	helmConfigHome := filepath.Join(banzaiCli.Home(), fmt.Sprintf("helm/org-%d/config", org))
	if err := os.MkdirAll(helmConfigHome, 0755); err != nil {
		return envs, errors.WrapIff(err, "failed to create %q directory", helmConfigHome)
	}
	helmCacheHome := filepath.Join(banzaiCli.Home(), fmt.Sprintf("helm/org-%d/cache", org))
	if err := os.MkdirAll(helmCacheHome, 0755); err != nil {
		return envs, errors.WrapIff(err, "failed to create %q directory", helmConfigHome)
	}

	envs["XDG_DATA_HOME"] = helmDataHome
	envs["XDG_CONFIG_HOME"] = helmConfigHome
	envs["XDG_CACHE_HOME"] = helmCacheHome

	return envs, nil
}

// TODO remove after deprecation
func setHelm2Env(envs map[string]string, banzaiCli cli.Cli) (map[string]string, error) {
	org := banzaiCli.Context().OrganizationID()
	helmHome := filepath.Join(banzaiCli.Home(), fmt.Sprintf("helm/org-%d", org))
	helmRepos := filepath.Join(helmHome, "repository")
	if err := os.MkdirAll(helmRepos, 0755); err != nil {
		return envs, errors.WrapIff(err, "failed to create %q directory", helmRepos)
	}

	if err := dumpRepositories(banzaiCli, helmRepos); err != nil {
		return envs, errors.WrapIf(err, "failed to sync Helm repositories")
	}

	helmPlugins := filepath.Join(helmHome, "plugins")
	if err := os.MkdirAll(helmPlugins, 0755); err != nil {
		return envs, errors.WrapIff(err, "failed to create %q directory", helmPlugins)
	}

	envs["HELM_HOME"] = helmHome

	return envs, nil
}
