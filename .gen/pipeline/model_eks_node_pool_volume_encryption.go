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
// EksNodePoolVolumeEncryption Encryption details of the instance volumes in an EKS node pool (default null -> control plane configuration -> AWS account default).
type EksNodePoolVolumeEncryption struct {
	// Indicator of encrypted node pool node volumes.
	Enabled bool `json:"enabled"`
	// KMS key ARN to use for node volume encryption.
	EncryptionKeyARN string `json:"encryptionKeyARN,omitempty"`
}
