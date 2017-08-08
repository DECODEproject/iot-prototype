package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

type deviceResource struct {
	entitlementStore *EntitlementStore
	metaStore        *MetadataStore
	deviceStore      map[string]DeviceResponse
}

func NewDeviceService(entitlementStore *EntitlementStore, metaStore *MetadataStore) deviceResource {
	return deviceResource{
		entitlementStore: entitlementStore,
		metaStore:        metaStore,
		deviceStore:      map[string]DeviceResponse{},
	}
}

type DeviceRequest struct {
	Type        string `json:"type" description:"type of device"`
	Description string `json:"description" description:"description of device"`
}

type DeviceResponse struct {
	DeviceRequest
	UID string `json:"uid" description:"unique identifier for the device"`
}

func (e deviceResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/devices").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"devices"}

	ws.Route(ws.POST("/").To(e.newDevice).
		Doc("add a new device").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(DeviceRequest{}).
		Writes(DeviceResponse{}).
		Returns(http.StatusOK, "OK", DeviceResponse{}).
		Returns(http.StatusInternalServerError, "Something bad happened", ErrorResponse{}))

	ws.Route(ws.GET("/").To(e.allDevices).
		Doc("retrieve configured devices").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]DeviceResponse{}).
		Returns(http.StatusOK, "OK", []DeviceResponse{}).
		Returns(http.StatusInternalServerError, "Something bad happened", ErrorResponse{}))

	return ws
}

func (e deviceResource) newDevice(request *restful.Request, response *restful.Response) {
}

func (e deviceResource) allDevices(request *restful.Request, response *restful.Response) {

	resp := []DeviceResponse{}

	for _, v := range e.deviceStore {
		resp = append(resp, v)
	}

	response.WriteEntity(resp)

}
