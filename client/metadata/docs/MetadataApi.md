# \MetadataApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AllItems**](MetadataApi.md#AllItems) | **Get** /catalog/items/ | get all cataloged items
[**CatalogItem**](MetadataApi.md#CatalogItem) | **Put** /catalog/items/{location-uid} | catalog an item for discovery e.g. what and where
[**MoveLocation**](MetadataApi.md#MoveLocation) | **Patch** /catalog/announce/{location-uid} | change a node&#39;s location - keeping the same location-uid
[**RegisterLocation**](MetadataApi.md#RegisterLocation) | **Put** /catalog/announce | register a node&#39;s location
[**RemoveFromCatalog**](MetadataApi.md#RemoveFromCatalog) | **Delete** /catalog/items/ | delete an item from the catalog


# **AllItems**
> []ApiItemWithLocation AllItems()

get all cataloged items

get all cataloged items


### Parameters
This endpoint does not need any parameter.

### Return type

[**[]ApiItemWithLocation**](api.ItemWithLocation.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CatalogItem**
> ApiCatalogItem CatalogItem($locationUid, $body)

catalog an item for discovery e.g. what and where

catalog an item for discovery e.g. what and where


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **locationUid** | **string**| identifier for a location | 
 **body** | [**ApiCatalogRequest**](ApiCatalogRequest.md)|  | 

### Return type

[**ApiCatalogItem**](api.CatalogItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MoveLocation**
> ApiLocation MoveLocation($locationUid, $body)

change a node's location - keeping the same location-uid

change a node's location - keeping the same location-uid


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **locationUid** | **string**| identifier for a location | 
 **body** | [**ApiLocationRequest**](ApiLocationRequest.md)|  | 

### Return type

[**ApiLocation**](api.Location.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RegisterLocation**
> ApiLocation RegisterLocation($body)

register a node's location

register a node's location


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ApiLocationRequest**](ApiLocationRequest.md)|  | 

### Return type

[**ApiLocation**](api.Location.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveFromCatalog**
> RemoveFromCatalog($subject)

delete an item from the catalog

delete an item from the catalog


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subject** | **string**| &#39;subject&#39; of the item to delete | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

