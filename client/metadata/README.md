# Go API client for swagger


## Overview
This API client was generated by the [swagger-codegen](https://github.com/swagger-api/swagger-codegen) project.  By using the [swagger-spec](https://github.com/swagger-api/swagger-spec) from a remote server, you can easily generate an API client.

- API version: 
- Package version: 1.0.0
- Build date: 2017-07-07T12:59:26.469+01:00
- Build package: class io.swagger.codegen.languages.GoClientCodegen

## Installation
Put the package under your project folder and add the following in import:
```
    "./swagger"
```

## Documentation for API Endpoints

All URIs are relative to *https://localhost*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*MetadataApi* | [**AllItems**](docs/MetadataApi.md#allitems) | **Get** /catalog/items/ | get all cataloged items
*MetadataApi* | [**CatalogItem**](docs/MetadataApi.md#catalogitem) | **Put** /catalog/items | catalog an item for discovery e.g. what and where
*MetadataApi* | [**MoveLocation**](docs/MetadataApi.md#movelocation) | **Patch** /catalog/announce/{location-uid} | change a node&#39;s location - keeping the same location-uid
*MetadataApi* | [**RegisterLocation**](docs/MetadataApi.md#registerlocation) | **Put** /catalog/announce | register a node&#39;s location
*MetadataApi* | [**RemoveFromCatalog**](docs/MetadataApi.md#removefromcatalog) | **Delete** /catalog/items/{catalog-uid} | delete an item from the catalog


## Documentation For Models

 - [ServicesItem](docs/ServicesItem.md)
 - [ServicesItemWithLocation](docs/ServicesItemWithLocation.md)
 - [ServicesLocation](docs/ServicesLocation.md)
 - [ServicesLocationRequest](docs/ServicesLocationRequest.md)


## Documentation For Authorization

 All endpoints do not require authorization.


## Author


