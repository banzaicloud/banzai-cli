# \IntegratedServicesApi

All URIs are relative to *http://localhost:9090*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateFeature**](IntegratedServicesApi.md#ActivateFeature) | **Post** /api/v1/orgs/{orgId}/clusters/{id}/features/{featureName} | Activate a feature
[**ActivateIntegratedService**](IntegratedServicesApi.md#ActivateIntegratedService) | **Post** /api/v1/orgs/{orgId}/clusters/{id}/services/{serviceName} | Activate an integrated service
[**DeactivateFeature**](IntegratedServicesApi.md#DeactivateFeature) | **Delete** /api/v1/orgs/{orgId}/clusters/{id}/features/{featureName} | Deactivate a feature
[**DeactivateIntegratedService**](IntegratedServicesApi.md#DeactivateIntegratedService) | **Delete** /api/v1/orgs/{orgId}/clusters/{id}/services/{serviceName} | Deactivate an integrated service
[**FeatureDetails**](IntegratedServicesApi.md#FeatureDetails) | **Get** /api/v1/orgs/{orgId}/clusters/{id}/features/{featureName} | Get details of a feature
[**IntegratedServiceDetails**](IntegratedServicesApi.md#IntegratedServiceDetails) | **Get** /api/v1/orgs/{orgId}/clusters/{id}/services/{serviceName} | Get details of an integrated service
[**ListFeatures**](IntegratedServicesApi.md#ListFeatures) | **Get** /api/v1/orgs/{orgId}/clusters/{id}/features | List enabled features of a cluster
[**ListIntegratedServices**](IntegratedServicesApi.md#ListIntegratedServices) | **Get** /api/v1/orgs/{orgId}/clusters/{id}/services | List enabled integrated services of a cluster
[**UpdateFeature**](IntegratedServicesApi.md#UpdateFeature) | **Put** /api/v1/orgs/{orgId}/clusters/{id}/features/{featureName} | Update a feature
[**UpdateIntegratedService**](IntegratedServicesApi.md#UpdateIntegratedService) | **Put** /api/v1/orgs/{orgId}/clusters/{id}/services/{serviceName} | Update an integrated service



## ActivateFeature

> ActivateFeature(ctx, orgId, id, featureName, activateIntegratedServiceRequest)

Activate a feature

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**featureName** | **string**| Feature name | 
**activateIntegratedServiceRequest** | [**ActivateIntegratedServiceRequest**](ActivateIntegratedServiceRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ActivateIntegratedService

> ActivateIntegratedService(ctx, orgId, id, serviceName, activateIntegratedServiceRequest)

Activate an integrated service

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**serviceName** | **string**| service name | 
**activateIntegratedServiceRequest** | [**ActivateIntegratedServiceRequest**](ActivateIntegratedServiceRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeactivateFeature

> DeactivateFeature(ctx, orgId, id, featureName)

Deactivate a feature

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**featureName** | **string**| Feature name | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeactivateIntegratedService

> DeactivateIntegratedService(ctx, orgId, id, serviceName)

Deactivate an integrated service

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**serviceName** | **string**| service name | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FeatureDetails

> IntegratedServiceDetails FeatureDetails(ctx, orgId, id, featureName)

Get details of a feature

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**featureName** | **string**| Feature name | 

### Return type

[**IntegratedServiceDetails**](IntegratedServiceDetails.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json, 

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IntegratedServiceDetails

> IntegratedServiceDetails IntegratedServiceDetails(ctx, orgId, id, serviceName)

Get details of an integrated service

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**serviceName** | **string**| service name | 

### Return type

[**IntegratedServiceDetails**](IntegratedServiceDetails.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json, 

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListFeatures

> map[string]IntegratedServiceDetails ListFeatures(ctx, orgId, id)

List enabled features of a cluster

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 

### Return type

[**map[string]IntegratedServiceDetails**](IntegratedServiceDetails.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json, 

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListIntegratedServices

> map[string]IntegratedServiceDetails ListIntegratedServices(ctx, orgId, id)

List enabled integrated services of a cluster

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 

### Return type

[**map[string]IntegratedServiceDetails**](IntegratedServiceDetails.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/problem+json, 

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateFeature

> UpdateFeature(ctx, orgId, id, featureName, updateIntegratedServiceRequest)

Update a feature

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**featureName** | **string**| Feature name | 
**updateIntegratedServiceRequest** | [**UpdateIntegratedServiceRequest**](UpdateIntegratedServiceRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateIntegratedService

> UpdateIntegratedService(ctx, orgId, id, serviceName, updateIntegratedServiceRequest)

Update an integrated service

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **int32**| Organization identifier | 
**id** | **int32**| Cluster identifier | 
**serviceName** | **string**| service name | 
**updateIntegratedServiceRequest** | [**UpdateIntegratedServiceRequest**](UpdateIntegratedServiceRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

