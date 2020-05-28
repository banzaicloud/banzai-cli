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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	kind "sigs.k8s.io/kind/pkg/apis/config/v1alpha3"
)

const version = "v0.8.1"
const clusterName = "banzai"
const kindCmd = "kind"

func isKINDInstalled(banzaiCli cli.Cli) bool {
	path, err := findKINDPath(banzaiCli)
	if path != "" && err == nil {
		versionOutput, err := exec.Command(path, "version").CombinedOutput()
		if err != nil {
			return false
		}

		versionSplit := strings.SplitN(string(versionOutput), " ", 3)
		if len(versionSplit) != 3 {
			return false
		}

		if version == versionSplit[1] {
			return true
		}

		log.Infof("KIND version mismatch, have %s, wanted %s...", versionSplit[1], version)
	}

	return false
}

func findKINDPath(banzaiCli cli.Cli) (string, error) {
	path := filepath.Join(banzaiCli.Home(), "bin", kindCmd)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return path, nil
}

func downloadKIND(banzaiCli cli.Cli) error {
	src := fmt.Sprintf("https://github.com/kubernetes-sigs/kind/releases/download/%s/kind-%s-amd64", version, runtime.GOOS)

	binDir := filepath.Join(banzaiCli.Home(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return errors.WrapIff(err, "failed to create %q directory", binDir)
	}
	kindPath := filepath.Join(binDir, kindCmd)

	resp, err := http.Get(src) // #nosec
	if err != nil {
		return errors.WrapIf(err, "failed to HTTP GET kind binary")
	}

	defer resp.Body.Close()

	tempName := kindPath + "~"
	f, err := os.OpenFile(tempName, (os.O_WRONLY | os.O_CREATE | os.O_EXCL), 0700)
	if err != nil {
		return errors.WrapIf(err, "failed to create temporary file for kind binary")
	}

	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return errors.WrapIf(err, "failed to write kind binary")
	}

	return errors.WrapIf(os.Rename(tempName, kindPath), "failed to move kind binary to its final place")
}

func ensureKINDCluster(banzaiCli cli.Cli, options *cpContext) error {
	if !isKINDInstalled(banzaiCli) {
		log.Infof("KIND binary (kind) is not available in $PATH, downloading version %s...", version)
		err := downloadKIND(banzaiCli)
		if err != nil {
			return errors.WrapIf(err, "failed to download kind binary")
		}
		log.Info("KIND installed")
	}

	kindPath, err := findKINDPath(banzaiCli)
	if err != nil {
		return err
	}

	cmd := exec.Command(kindPath, "get", "kubeconfig", "--name", clusterName)
	if err := cmd.Run(); err == nil {
		if options.kubeconfigExists() {
			return nil
		}

		return errors.Errorf("a KIND cluster named %q already exists", clusterName)
	}

	cluster := kind.Cluster{
		TypeMeta: kind.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.sigs.k8s.io/v1alpha3",
		},
		Nodes: []kind.Node{
			{
				Role: kind.ControlPlaneRole,
				ExtraPortMappings: []kind.PortMapping{
					{
						ContainerPort: 80,
						HostPort:      80,
						ListenAddress: "127.0.0.1",
					},
					{
						ContainerPort: 443,
						HostPort:      443,
						ListenAddress: "127.0.0.1",
					},
				},
			},
		},
	}

	buff, err := yaml.Marshal(&cluster)
	if err != nil {
		return errors.WrapIf(err, "failed to prepare KIND cluster config")
	}

	kindConfigFile := filepath.Join(options.workspace, "kind-config.yaml")
	err = ioutil.WriteFile(kindConfigFile, buff, 0600)
	if err != nil {
		return errors.WrapIf(err, "failed to write KIND cluster config")
	}

	cmd = exec.Command(kindPath, "create", "cluster", "--config", kindConfigFile, "--name", clusterName, "--kubeconfig", options.kubeconfigPath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return errors.WrapIf(err, "failed to create KIND cluster")
	}

	cmd = exec.Command(kindPath, "get", "kubeconfig", "--name", clusterName)
	kubeconfig, err := cmd.Output()
	if err != nil {
		return errors.WrapIf(err, "failed to get KIND kubeconfig")
	}

	if runtime.GOOS != "linux" {
		// non-native docker daemons can't access the host machine directly even if running in host networking mode
		// we have to rewrite configs referring to localhost to use the special name `host.docker.internal` instead
		kubeconfig, err = input.RewriteLocalhostToHostDockerInternal(kubeconfig)
		if err != nil {
			return errors.WrapIf(err, "failed to rewrite Kubernetes config")
		}
	}

	return nil
}

func deleteKINDCluster(banzaiCli cli.Cli) error {
	kindPath, err := findKINDPath(banzaiCli)
	if err != nil {
		return err
	}

	cmd := exec.Command(kindPath, "delete", "cluster", "--name", clusterName)

	_, err = cmd.Output()
	if err != nil {
		return errors.WrapIf(err, "failed to delete KIND cluster")
	}

	return nil
}
