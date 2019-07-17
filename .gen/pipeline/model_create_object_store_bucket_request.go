/*
 * Pipeline API
 *
 * Pipeline v0.3.0 swagger
 *
 * API version: 0.26.0
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pipeline

type CreateObjectStoreBucketRequest struct {
	SecretId string `json:"secretId,omitempty"`
	SecretName string `json:"secretName,omitempty"`
	Name string `json:"name"`
	Properties CreateObjectStoreBucketProperties `json:"properties"`
}
