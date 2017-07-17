# \DataApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Append**](DataApi.md#Append) | **Put** /data/ | append data to a bucket, will create the bucket if it does not exist.
[**GetAll**](DataApi.md#GetAll) | **Get** /data/ | returns all of the data stored in a logical &#39;bucket&#39;.


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
> []ApiDataResponse GetAll($from, $to, $bucketUid)

returns all of the data stored in a logical 'bucket'.

returns all of the data stored in a logical 'bucket'.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **from** | **string**| return data from this ISO8601 timestamp. Defaults to 24 hours ago. | [optional] [default to ]
 **to** | **string**| finish at this ISO8601 timestamp  | [optional] [default to 2017-07-17T11:36:19.547+01:00]
 **bucketUid** | **string**| name of the &#39;bucket&#39; of data | [optional] [default to ]

### Return type

[**[]ApiDataResponse**](api.DataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

