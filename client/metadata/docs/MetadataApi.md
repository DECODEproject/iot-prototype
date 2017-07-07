# \MetadataApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AllItems**](MetadataApi.md#AllItems) | **Get** /catalog/items/ | get all cataloged items
[**CatalogItem**](MetadataApi.md#CatalogItem) | **Put** /catalog/items | catalog an item for discovery e.g. what and where
[**MoveLocation**](MetadataApi.md#MoveLocation) | **Patch** /catalog/announce/{location-uid} | change a node&#39;s location - keeping the same location-uid
[**RegisterLocation**](MetadataApi.md#RegisterLocation) | **Put** /catalog/announce | register a node&#39;s location
[**RemoveFromCatalog**](MetadataApi.md#RemoveFromCatalog) | **Delete** /catalog/items/{catalog-uid} | delete an item from the catalog


# **AllItems**
> AllItems()

get all cataloged items

get all cataloged items


### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CatalogItem**
> ServicesItem CatalogItem($body)

catalog an item for discovery e.g. what and where

catalog an item for discovery e.g. what and where


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ServicesItem**](ServicesItem.md)|  | 

### Return type

[**ServicesItem**](services.Item.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MoveLocation**
> ServicesLocation MoveLocation($locationUid, $body)

change a node's location - keeping the same location-uid

change a node's location - keeping the same location-uid


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **locationUid** | **string**| identifier for a location | 
 **body** | [**ServicesLocationRequest**](ServicesLocationRequest.md)|  | 

### Return type

[**ServicesLocation**](services.Location.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RegisterLocation**
> ServicesLocation RegisterLocation($body)

register a node's location

register a node's location


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ServicesLocationRequest**](ServicesLocationRequest.md)|  | 

### Return type

[**ServicesLocation**](services.Location.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveFromCatalog**
> RemoveFromCatalog($catalogUid)

delete an item from the catalog

delete an item from the catalog


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **catalogUid** | **string**| identifier for a cataloged item | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

