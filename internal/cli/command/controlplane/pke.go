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
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
)

func isPKEInstalled(banzaiCli cli.Cli) bool {
	path, err := findPKEPath(banzaiCli)
	return path != "" && err == nil
}

func findPKEPath(banzaiCli cli.Cli) (string, error) {
	path := filepath.Join(banzaiCli.Home(), "bin", "pke")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	} else {
		return path, nil
	}

	path = "/opt/banzaicloud/bin/pke"
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	} else {
		return path, nil
	}

	if path, err := exec.LookPath("pke"); err != nil {
		return "", nil
	} else {
		return path, nil
	}
}

func downloadPKE(banzaiCli cli.Cli) error {

	const src = "https://banzaicloud.com/downloads/pke/latest"

	binDir := filepath.Join(banzaiCli.Home(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return errors.WrapIff(err, "failed to create %q directory", binDir)
	}
	pkePath := filepath.Join(binDir, "pke")

	resp, err := http.Get(src)
	if err != nil {
		return errors.WrapIf(err, "failed to HTTP GET pke binary")
	}

	defer resp.Body.Close()

	tempName := pkePath + "~"
	f, err := os.OpenFile(tempName, (os.O_WRONLY | os.O_CREATE | os.O_EXCL), 0700)
	if err != nil {
		return errors.WrapIf(err, "failed to create temporary file for pke binary")
	}

	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return errors.WrapIf(err, "failed to write pke binary")
	}

	return errors.WrapIf(os.Rename(tempName, pkePath), "failed to move pke binary to its final place")
}

func ensurePKECluster(banzaiCli cli.Cli, options *cpContext) error {
	if options.kubeconfigExists() {
		return nil
	}

	if !isPKEInstalled(banzaiCli) {
		log.Info("PKE binary (pke) is not available, downloading...")
		err := downloadPKE(banzaiCli)
		if err != nil {
			return errors.WrapIf(err, "failed to download pke binary")
		}
		log.Info("PKE downloaded")
	}

	pkePath, err := findPKEPath(banzaiCli)
	if err != nil {
		return err
	}

	log.Info("Installing single-node PKE cluster...")
	cmd := exec.Command(pkePath, "install", "single")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to install PKE cluster")
	}

	bytes, err := ioutil.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		return errors.WrapIf(err, "can't read Kubernetes secret")
	}

	return options.writeKubeconfig(bytes)
}

func checkPKESupported() error {
	cmd := exec.Command("rpm", "--query", "centos-release")
	err := cmd.Run()
	if err != nil {
		cmd := exec.Command("rpm", "--query", "redhat-release")
		err = cmd.Run()
	}

	return errors.Wrap(err, "unsupported OS")
}

func checkRoot() error {
	if os.Getuid() != 0 {
		return errors.Errorf("this command must be run as root")
	}
	return nil
}

func guessExternalAddr() string {
	host, err := os.Hostname()
	if err == nil && !strings.HasSuffix(host, ".internal") {
		if strings.Contains(host, ".") {
			log.Debugf("using hostname which is an fqdn: %q", host)
			return host
		}
		addrs, err := net.LookupIP(host)
		if err == nil {
			for _, addr := range addrs {
				if ipv4 := addr.To4(); ipv4 != nil && ipv4.IsGlobalUnicast() {
					hosts, err := net.LookupAddr(string(ipv4))
					if err == nil || len(hosts) > 0 {
						fqdn := strings.TrimSuffix(hosts[0], ".")
						log.Debugf("using reverse record of IP the hostname resolves to: %q", fqdn)
						return fqdn
					}
				}
			}
		}
	}

	client := http.Client{Timeout: time.Second}
	if response, err := client.Get("http://169.254.169.254/latest/meta-data/public-hostname"); err == nil {
		if response.StatusCode == http.StatusOK {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err == nil && len(body) > 0 {
				log.Debugf("using public hostname from metadata: %q", body)
				return string(body)
			}
		}
	}

	cmd := exec.Command("ip", "ro", "get", "93.184.216.34") // example.com, as a random internet address
	out, err := cmd.Output()
	if err == nil {
		last := ""
		for _, word := range strings.Fields(string(out)) {
			if last == "src" {
				log.Debugf("using default source address: %q", word)
				return word
			}
			last = word
		}
	}
	log.Debugf("using hostname: %q", host)
	return host
}
