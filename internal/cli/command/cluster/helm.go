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
	"strings"

	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"k8s.io/client-go/tools/clientcmd"
)

func getK8sClientSets(kubeconfig string) (*clientappsv1.AppsV1Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, emperror.WrapWith(err, "failed to get kubernetes config", "kubeconfig", kubeconfig)
	}

	appsClientSet, err := clientappsv1.NewForConfig(config)
	if err != nil {
		return nil, emperror.Wrap(err, "cannot create new core clientSet")
	}

	return appsClientSet, nil
}

func getClusterTillerVersion(kubeconfig string) (string, error) {
	cli, _ := getK8sClientSets(kubeconfig)
	deployment, err := cli.Deployments("kube-system").Get("tiller-deploy", metav1.GetOptions{})
	if err != nil {
		return "", emperror.Wrap(err, "cannot get tiller-deploy deployment")
	}
	version := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, ":")[1]
	return version, nil
}

func setClusterHelm(kubeconfig string) error {
	version, err := getClusterTillerVersion(kubeconfig)
	if err != nil {
		return err
	}
	log.Printf("version %s", version)

	return nil
}

func downloadClusterHelm() error {
	return nil
}
