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

type BaseError struct {
	Code int32 `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Error string `json:"error,omitempty"`
}
