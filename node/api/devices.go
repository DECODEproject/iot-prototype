package api

import (
	"context"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/sensors"
	"gogs.dyne.org/DECODE/decode-prototype-da/utils"
	validator "gopkg.in/validator.v2"

	randomdata "github.com/Pallinder/go-randomdata"
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
	Type string `json:"type" description:"type of device" validate:"nonzero"`
}

type DeviceResponse struct {
	DeviceRequest
	UID         string `json:"uid" description:"unique identifier for the device"`
	Name        string `json:"name" description:"unique name for the device" `
	Description string `json:"description" description:"information about the device" `
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
	subject := utils.BuildSubjectKey(uid)
	e.entitlementStore.Accepted.Add(
		Entitlement{
			EntitlementRequest: EntitlementRequest{
				Subject:     subject.String(),
				AccessLevel: OwnerOnly,
			},
			UID:    uuid.NewV4().String(),
			Status: Accepted,
		},
	)
	var description string

	switch req.Type {
	case "fake-sine":
		description = "fake device producing a sine curve"
		sensor := sensors.NewSineCurveEmitterSensor(e.ctx, uid, e.out)
		go sensor.Start()
	case "fake-temp-humidity":
		description = "fake device producing temperature and humidity values"
		sensor := sensors.NewTemperatureHumiditySensor(e.ctx, uid, e.out)
		go sensor.Start()

	default:
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: "unknown device type"})
		return
	}

	name := randomdata.SillyName()
	// add the metadata to the catalog
	e.metaStore.Add(Metadata{Subject: subject.String(), Description: description, Name: name})

	resp := DeviceResponse{
		DeviceRequest: req,
		UID:           uid,
		Description:   description,
		Name:          name,
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
