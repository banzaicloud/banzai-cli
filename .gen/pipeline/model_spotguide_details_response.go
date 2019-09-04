/*
 * Pipeline API
 *
 * Pipeline v0.3.0 swagger
 *
 * API version: 0.3.0
 * Contact: info@banzaicloud.com
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pipeline

type SpotguideDetailsResponse struct {
	Name string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Readme string `json:"readme,omitempty"`
	Version string `json:"version,omitempty"`
	Tags []string `json:"tags,omitempty"`
	Resources RequestedResources `json:"resources,omitempty"`
	Questions []SpotguideOption `json:"questions,omitempty"`
}
