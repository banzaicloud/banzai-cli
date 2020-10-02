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

package cluster

import (
	"encoding/json"
	"fmt"

	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
)

type TemplateNotFoundError struct {
	Name string
}

func (te TemplateNotFoundError) Error() string {
	return fmt.Sprintf("Template not found with name %s", te.Name)
}

var (
	templates = map[string]interface{}{
		pkeOnAws: pipeline.CreateClusterRequest{
			Cloud:    "amazon",
			Location: "us-east-2",
			Properties: map[string]interface{}{
				"pke": pipeline.CreatePkeProperties{
					NodePools: []pipeline.NodePoolsPke{
						{
							Name:     "master",
							Roles:    []string{"master", "worker"},
							Provider: "amazon",
							ProviderConfig: map[string]interface{}{
								"autoScalingGroup": map[string]interface{}{
									"instanceType": "c5.large",
									"zones":        []string{"us-east-2a"},
									"spotPrice":    "",
									"size": map[string]interface{}{
										"desired": 1,
										"min":     1,
										"max":     1,
									},
								},
							},
						},
					},
					Kubernetes: pipeline.CreatePkePropertiesKubernetes{
						Version: "v1.17.9",
						Rbac: pipeline.CreatePkePropertiesKubernetesRbac{
							Enabled: true,
						},
					},
					Cri: pipeline.CreatePkePropertiesCri{
						Runtime: "containerd",
					},
				},
			},
		},
		pkeOnAzure: pipeline.CreatePkeOnAzureClusterRequest{
			Type:     pkeOnAzure,
			Location: "westus2",
			Nodepools: []pipeline.PkeOnAzureNodePool{
				{
					Name:         "master",
					Roles:        []string{"master", "worker"},
					Autoscaling:  false,
					MinCount:     1,
					MaxCount:     1,
					Count:        1,
					InstanceType: "Standard_D2s_v3",
				},
			},
			Kubernetes: pipeline.CreatePkeClusterKubernetes{
				Version: "1.17.9",
				Rbac:    true,
			},
		},
		pkeOnVsphere: pipeline.CreatePkeOnVsphereClusterRequest{
			Type: pkeOnVsphere,
			Kubernetes: pipeline.CreatePkeClusterKubernetes{
				Version: "1.15.3",
				Rbac:    true,
			},
			Folder:       "folder",
			Datastore:    "DatastoreCluster",
			ResourcePool: "resource-pool",
			Nodepools: []pipeline.PkeOnVsphereNodePool{
				{
					Name:          "master",
					Roles:         []string{"master", "worker"},
					Size:          1,
					Vcpu:          2,
					Ram:           1024,
					Template:      "pke-template",
					AdminUsername: "root",
				},
			},
		},
		"ack": pipeline.CreateClusterRequest{
			Cloud: "alibaba",
			Properties: map[string]interface{}{
				"ack": pipeline.CreateAckPropertiesAck{},
			},
		},
		"aks": pipeline.CreateClusterRequest{
			Cloud:    "azure",
			Location: "westus2",
			Properties: map[string]interface{}{
				"aks": pipeline.CreateAksPropertiesAks{
					KubernetesVersion: "1.17.11",
					NodePools: map[string]pipeline.NodePoolsAzure{
						"pool1": {
							Count:        3,
							Autoscaling:  true,
							InstanceType: "Standard_B2s",
							MinCount:     3,
							MaxCount:     4,
						},
					},
				},
			},
		},
		"eks": pipeline.CreateClusterRequest{
			Cloud:    "amazon",
			Location: "us-east-2",
			Properties: map[string]interface{}{
				"eks": pipeline.CreateEksPropertiesEks{
					Version: "1.16.13",
					NodePools: map[string]pipeline.EksNodePool{
						"pool1": {
							Count:        3,
							MinCount:     3,
							MaxCount:     4,
							Autoscaling:  true,
							InstanceType: "t2.medium",
						},
					},
				},
			},
		},
		"gke": pipeline.CreateClusterRequest{
			Cloud:    "google",
			Location: "us-west1-b",
			Properties: map[string]interface{}{
				"gke": pipeline.CreateGkePropertiesGke{
					Master: pipeline.CreateGkePropertiesGkeMaster{
						Version: "1.15.12-gke.20",
					},
					NodeVersion: "1.15.12-gke.20",
					NodePools: map[string]pipeline.NodePoolsGoogle{
						"pool1": {
							Count: 3, Autoscaling: true,
							InstanceType: "e2-medium",
							Preemptible:  true,
							MinCount:     3,
							MaxCount:     4,
						},
					},
				},
			},
		},
		"oke": pipeline.CreateClusterRequest{
			Cloud:    "oracle",
			Location: "eu-frankfurt-1",
			Properties: map[string]interface{}{
				"oke": pipeline.CreateUpdateOkePropertiesOke{
					Version: "v1.14.8",
					NodePools: map[string]pipeline.NodePoolsOracle{
						"pool1": {
							Count:   3,
							Shape:   "VM.Standard.E2.2",
							Version: "v1.14.8",
						},
					},
				},
			},
		},
	}
)

func convertCreateTemplate(in interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	b, _ := json.Marshal(in)
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, errors.WrapIf(err, "failed to convert create template")
	}

	return out, nil
}
