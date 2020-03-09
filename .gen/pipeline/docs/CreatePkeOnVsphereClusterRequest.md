# CreatePkeOnVsphereClusterRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**SecretId** | **string** |  | [optional] 
**SecretName** | **string** |  | [optional] 
**SshSecretId** | **string** |  | [optional] 
**ScaleOptions** | [**ScaleOptions**](ScaleOptions.md) |  | [optional] 
**Type** | **string** |  | 
**Kubernetes** | [**CreatePkeClusterKubernetes**](CreatePKEClusterKubernetes.md) |  | 
**Proxy** | [**PkeClusterHttpProxy**](PKEClusterHTTPProxy.md) |  | [optional] 
**Folder** | **string** | Folder to create nodes in. | [optional] 
**Datastore** | **string** | Name of datastore or datastore cluster to place VM disks on. | [optional] 
**ResourcePool** | **string** | Virtual machines will be created in this resource pool. | [optional] 
**Nodepools** | [**[]PkeOnVsphereNodePool**](PKEOnVsphereNodePool.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


