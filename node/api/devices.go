package api

import (
	"context"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/sensors"
	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

type deviceResource struct {
	entitlementStore *EntitlementStore
	metaStore        *MetadataStore
	deviceStore      map[string]DeviceResponse
	out              chan<- sensors.SensorMessage
	ctx              context.Context
}

func NewDeviceService(ctx context.Context, entitlementStore *EntitlementStore, metaStore *MetadataStore, out chan<- sensors.SensorMessage) deviceResource {
	return deviceResource{
		entitlementStore: entitlementStore,
		metaStore:        metaStore,
		deviceStore:      map[string]DeviceResponse{},
		out:              out,
		ctx:              ctx,
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
		Reads(DeviceRequest{Description: "a demonstration sine wave generator.", Type: "sine"}).
		Writes(DeviceResponse{}).
		Returns(http.StatusOK, "OK", DeviceResponse{}).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
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

	req := DeviceRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	uid := uuid.NewV4().String()

	// create an entitlement
	// TODO : need to deal with devices that output multiple keys
	// of data
	e.entitlementStore.Accepted.Add(
		Entitlement{
			EntitlementRequest: EntitlementRequest{
				Subject:     uid,
				AccessLevel: OwnerOnly,
			},
			UID:    uuid.NewV4().String(),
			Status: Accepted,
		},
	)

	// add the metadata to the catalog
	e.metaStore.Add(Metadata{Subject: uid, Description: req.Description})

	// start accepting from the sensor
	if req.Type == "sine" {
		sensor := sensors.NewSineCurveEmitterSensor(e.ctx, uid, e.out)
		go sensor.Start()
	}

	resp := DeviceResponse{
		DeviceRequest: req,
		UID:           uid,
	}

	// TODO : add lock
	e.deviceStore[uid] = resp

	response.WriteEntity(resp)

}

func (e deviceResource) allDevices(request *restful.Request, response *restful.Response) {

	resp := []DeviceResponse{}

	for _, v := range e.deviceStore {
		resp = append(resp, v)
	}

	response.WriteEntity(resp)

}
