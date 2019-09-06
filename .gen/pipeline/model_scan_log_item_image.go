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

type ScanLogItemImage struct {
	ImageName string `json:"imageName,omitempty"`
	ImageTag string `json:"imageTag,omitempty"`
	ImageDigest string `json:"imageDigest,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}
