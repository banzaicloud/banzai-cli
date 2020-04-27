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
// User struct for User
type User struct {
	Id int32 `json:"id,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Name string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Login string `json:"login,omitempty"`
	Image string `json:"image,omitempty"`
	Organizations map[string]interface{} `json:"organizations,omitempty"`
}
