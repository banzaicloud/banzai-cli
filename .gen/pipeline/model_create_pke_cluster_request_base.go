/*
 * Pipeline API
 *
 * Pipeline v0.3.0 swagger
 *
 * API version: 0.21.2
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pipeline

type CreatePkeClusterRequestBase struct {
	Name string `json:"name"`
	Features []Feature `json:"features,omitempty"`
	SecretId string `json:"secretId,omitempty"`
	SecretName string `json:"secretName,omitempty"`
	SshSecretId string `json:"sshSecretId,omitempty"`
	ScaleOptions ScaleOptions `json:"scaleOptions,omitempty"`
	Type string `json:"type"`
	Kubernetes CreatePkeClusterKubernetes `json:"kubernetes"`
}
