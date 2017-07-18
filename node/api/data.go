package api

import (
	"net/http"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	storageclient "gogs.dyne.org/DECODE/decode-prototype-da/client/storage"
)

type dataResource struct {
	store   *EntitlementStore
	sClient *storageclient.DataApi
}

type DataRequest struct {
	Key string `json:"key" description:"unique name for the data"`
}

type DataResponse struct {
	Data interface{} `json:"data" description:"data returned"`
}

func NewDataService(store *EntitlementStore, sClient *storageclient.DataApi) dataResource {

	return dataResource{
		store:   store,
		sClient: sClient,
	}
}

func (e dataResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/data").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"data"}

	ws.Route(ws.POST("/").To(e.getData).
		Doc("retrieve some data").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(DataRequest{}).
		Writes(DataResponse{}).
		Returns(http.StatusOK, "OK", DataResponse{}).
		Returns(http.StatusInternalServerError, "Something bad happened", ErrorResponse{}).
		Returns(http.StatusForbidden, "Forbidden", nil).
		Returns(http.StatusNotFound, "Not Found", nil))

	return ws
}

func (e dataResource) getData(request *restful.Request, response *restful.Response) {
	req := DataRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	// find entitlement for key
	ent, found := e.store.Accepted.FindForSubject(req.Key)

	if !found {
		response.WriteHeader(http.StatusForbidden)
		return
	}

	if !ent.IsAccessible() {
		response.WriteHeader(http.StatusForbidden)
		return
	}

	// go and get it
	data, _, err := e.sClient.GetAll("", "", req.Key)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := DataResponse{
		Data: data,
	}

	response.WriteEntity(resp)
}
