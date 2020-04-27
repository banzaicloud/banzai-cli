# CreatePkeOnVsphereClusterRequestAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StorageSecretId** | **string** | Secret ID used to setup VSphere storage classes. Overrides the default settings in main cluster secret. | [optional] 
**StorageSecretName** | **string** | Secret name used to setup VSphere storage classes. Overrides default value from the main cluster secret. | [optional] 
**Folder** | **string** | Folder to create nodes in. Overrides default value from the main cluster secret. | [optional] 
**Datastore** | **string** | Name of datastore or datastore cluster to place VM disks on. Overrides default value from the main cluster secret. | [optional] 
**ResourcePool** | **string** | Virtual machines will be created in this resource pool. Overrides default value from the main cluster secret. | [optional] 
**NodePools** | [**[]PkeOnVsphereNodePool**](PKEOnVsphereNodePool.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


