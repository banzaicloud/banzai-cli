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
	"regexp"
	"runtime"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kind "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const version = "v0.9.0"
const clusterName = "banzai"
const kindCmd = "kind"

func isKINDInstalled(banzaiCli cli.Cli) bool {
	path, err := findKINDPath(banzaiCli)
	if path != "" && err == nil {
		versionOutput, err := exec.Command(path, "version").CombinedOutput()
		if err != nil {
			return false
		}

		semver := regexp.MustCompile("v[0-9]+.[0-9]+.[0-9]+")
		currentVersion := semver.FindString(string(versionOutput))

		if currentVersion == version {
			return true
		}

		log.Infof("KIND version mismatch, have %s, wanted %s...", currentVersion, version)
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

func ensureKINDCluster(banzaiCli cli.Cli, options *cpContext, listenAddress string) error {
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
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Nodes: []kind.Node{
			{
				Role: kind.ControlPlaneRole,
				ExtraPortMappings: []kind.PortMapping{
					{
						ContainerPort: 80,
						HostPort:      80,
						ListenAddress: listenAddress,
					},
					{
						ContainerPort: 443,
						HostPort:      443,
						ListenAddress: listenAddress,
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
		_, err = input.RewriteLocalhostToHostDockerInternal(kubeconfig)
		if err != nil {
			return errors.WrapIf(err, "failed to rewrite Kubernetes config")
		}
	}

	if runtime.GOOS == "darwin" {
		err = fixKind0_8_1KubeProxy(kubeconfig)
		if err != nil {
			return errors.Wrap(err, "failed to fix Kind v0.8.1 kube-proxy CrashLoopBackoff")
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

// fixKind0_8_1KubeProxy returns an error if failed, otherwise fixes the Kind
// v0.8.1 kube-proxy issue of
// https://github.com/kubernetes-sigs/kind/issues/2240 on macOS without
// requiring Kind upgrade to v0.11.1 / because that would pull in K8s client
// v0.21 upgrade which breaks bank-vaults 1.3 and would require bank-vaults 1.13
// which would pull in a lot of changes.
func fixKind0_8_1KubeProxy(kubeconfig []byte) error {
	namespace := "kube-system"
	configMapName := "kube-proxy"
	configFileName := "config.conf"

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "instantiating ClientConfig failed")
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return errors.Wrap(err, "retrieving config failed")
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "instantiating ClientSet failed")
	}

	getContext, cancelFunction := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunction()

	kubeProxyConfigMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(
		getContext,
		configMapName,
		metav1.GetOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "retrieving kube-proxy ConfigMap failed")
	}

	configConfData := kubeProxyConfigMap.Data[configFileName]

	var configConf map[string]interface{}
	err = yaml.Unmarshal([]byte(configConfData), &configConf)
	if err != nil {
		return errors.WrapWithDetails(err, "unmarshalling kube-proxy config.conf data failed", "configConf")
	}

	connTrack, isOk := configConf["conntrack"].(map[string]interface{})
	if !isOk {
		return errors.Errorf(
			"retrieving conntrack object from kube-proxy conf.conf failed, configConfData: %+v",
			configConf,
		)
	}

	connTrack["maxPerCore"] = 0
	configConf["conntrack"] = connTrack

	editedConfigConfYAML, err := yaml.Marshal(configConf)
	if err != nil {
		return errors.Wrap(err, "marshalling config.conf into YAML failed")
	}

	kubeProxyConfigMap.Data[configFileName] = string(editedConfigConfYAML)

	updateContext, updateCancelFunction := context.WithTimeout(context.Background(), 15*time.Second)
	defer updateCancelFunction()

	_, err = clientSet.CoreV1().ConfigMaps(namespace).Update(
		updateContext,
		kubeProxyConfigMap,
		metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			FieldManager: "banzai-cli",
		},
	)
	if err != nil {
		return errors.Wrap(err, "patching kube-proxy ConfigMap failed")
	}

	return nil
}
