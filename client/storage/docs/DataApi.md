# \DataApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Append**](DataApi.md#Append) | **Put** /data/ | append data to a bucket, will create the bucket if it does not exist.
[**GetAll**](DataApi.md#GetAll) | **Get** /data/ | returns all of the data stored in a logical &#39;bucket&#39; in the last 24 hours.


# **Append**
> Append($body)

append data to a bucket, will create the bucket if it does not exist.

append data to a bucket, will create the bucket if it does not exist.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ApiData**](ApiData.md)|  | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAll**
> []ApiDataResponse GetAll($bucketUid)

returns all of the data stored in a logical 'bucket' in the last 24 hours.

returns all of the data stored in a logical 'bucket' in the last 24 hours.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bucketUid** | **string**| name of the &#39;bucket&#39; of data | [optional] 

### Return type

[**[]ApiDataResponse**](api.DataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

