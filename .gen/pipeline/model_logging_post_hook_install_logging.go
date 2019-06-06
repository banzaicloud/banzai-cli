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

type LoggingPostHookInstallLogging struct {
	BucketName string `json:"bucketName"`
	Region string `json:"region,omitempty"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
	StorageAccount string `json:"storageAccount,omitempty"`
	SecretId string `json:"secretId"`
	SecretName string `json:"secretName,omitempty"`
	Tls GenTlsForLogging `json:"tls"`
}
