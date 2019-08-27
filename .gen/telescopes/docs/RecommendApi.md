# \RecommendApi

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**RecommendCluster**](RecommendApi.md#RecommendCluster) | **Post** /recommender/provider/{provider}/service/{service}/region/{region}/cluster | Provides a recommended set of node pools on a given provider in a specific region.
[**RecommendClusterScaleOut**](RecommendApi.md#RecommendClusterScaleOut) | **Put** /recommender/provider/{provider}/service/{service}/region/{region}/cluster | Provides a recommendation for a scale-out, based on a current cluster layout on a given provider in a specific region.
[**RecommendMultiCluster**](RecommendApi.md#RecommendMultiCluster) | **Post** /recommender/multicloud | Provides a recommended set of node pools on a given provider in a specific region.


# **RecommendCluster**
> RecommendationResponse RecommendCluster(ctx, provider, service, region, recommendClusterRequest)
Provides a recommended set of node pools on a given provider in a specific region.

Provides a recommended set of node pools on a given provider in a specific region.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **provider** | **string**| provider | 
  **service** | **string**| service | 
  **region** | **string**| region | 
  **recommendClusterRequest** | [**RecommendClusterRequest**](RecommendClusterRequest.md)| request params | 

### Return type

[**RecommendationResponse**](recommendationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RecommendClusterScaleOut**
> RecommendationResponse RecommendClusterScaleOut(ctx, provider, service, region, recommendClusterScaleOutRequest)
Provides a recommendation for a scale-out, based on a current cluster layout on a given provider in a specific region.

Provides a recommendation for a scale-out, based on a current cluster layout on a given provider in a specific region.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **provider** | **string**| provider | 
  **service** | **string**| service | 
  **region** | **string**| region | 
  **recommendClusterScaleOutRequest** | [**RecommendClusterScaleOutRequest**](RecommendClusterScaleOutRequest.md)| request params | 

### Return type

[**RecommendationResponse**](recommendationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RecommendMultiCluster**
> RecommendationResponse RecommendMultiCluster(ctx, recommendMultiClusterRequest)
Provides a recommended set of node pools on a given provider in a specific region.

Provides a recommended set of node pools on a given provider in a specific region.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **recommendMultiClusterRequest** | [**RecommendMultiClusterRequest**](RecommendMultiClusterRequest.md)| request params | 

### Return type

[**RecommendationResponse**](recommendationResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

