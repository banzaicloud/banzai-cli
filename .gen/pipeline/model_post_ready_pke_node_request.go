/*
 * Pipeline API
 *
 * Pipeline is a feature rich application platform, built for containers on top of Kubernetes to automate the DevOps experience, continuous application development and the lifecycle of deployments. 
 *
 * API version: latest
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pipeline

type PostReadyPkeNodeRequest struct {
	// kubeconfig in base64 or empty if not a master
	Config string `json:"config,omitempty"`
	// name of node
	Name string `json:"name,omitempty"`
	// name of nodepool
	NodePool string `json:"nodePool,omitempty"`
	// ip address of node (where the other nodes can reach it)
	Ip string `json:"ip,omitempty"`
	// if this node is a master node
	Master bool `json:"master,omitempty"`
	// if this node is a worker node
	Worker bool `json:"worker,omitempty"`
}
