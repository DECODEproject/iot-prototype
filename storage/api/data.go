package api

import (
	"net/http"
	"time"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	redis_client "github.com/garyburd/redigo/redis"
)

// ErrorResponse signals error messages back to the client
type ErrorResponse struct {
	Error string `json:"error" description:"error message if any"`
}

// Data is a value to save to storage
type Data struct {
	Value  interface{} `json:"value" description:"data to save" validate:"nonzero" type:"object"`
	Bucket string      `json:"bucket" description:"unique bucket to save value to" validate:"nonzero"`
}

// DataResponse is the saved value with the time it was saved
type DataResponse struct {
	Value     interface{} `json:"value" description:"saved value" type:"object"`
	Timestamp time.Time   `json:"ts" description:"when the item was saved"`
}

type dataResource struct {
	pool *redis_client.Pool
}

func NewDataService(pool *redis_client.Pool) dataResource {

	return dataResource{
		pool: pool,
	}
}

func (e dataResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/data").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"data"}

	ws.Route(ws.GET("/").To(e.getAll).
		Doc("returns all of the data stored in a logical 'bucket' in the last 24 hours.").
		Param(ws.QueryParameter("bucket-uid", "name of the 'bucket' of data").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]DataResponse{}).
		Returns(http.StatusOK, "OK", []DataResponse{}).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Something went wrong", ErrorResponse{}))

	ws.Route(ws.PUT("/").To(e.append).
		Doc("append data to a bucket, will create the bucket if it does not exist.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Data{}).
		Returns(http.StatusCreated, "Data was accepted", nil).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusInternalServerError, "Something went wrong", ErrorResponse{}))

	return ws
}
func (e dataResource) getAll(request *restful.Request, response *restful.Response) {

	prefix := request.QueryParameter("bucket-uid")

	timestep := time.Second
	expiry := time.Duration(time.Hour * 24)

	ts := NewTimeSeries(prefix, timestep, expiry, e.pool)

	// TODO : review should this be UTC?
	to := time.Now()
	from := to.Add(-(time.Hour * 24))

	data := []*DataResponse{}

	err := ts.FetchRange(from, to, &data)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response.WriteEntity(data)
}

func (e dataResource) append(request *restful.Request, response *restful.Response) {

	data := Data{}

	if err := request.ReadEntity(&data); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(data); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	prefix := data.Bucket
	timestep := time.Second
	expiry := time.Duration(time.Hour * 24)

	ts := NewTimeSeries(prefix, timestep, expiry, e.pool)

	// TODO : should this be UTC
	now := time.Now()
	err := ts.Add(&DataResponse{Value: data.Value, Timestamp: now}, now)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response.WriteHeader(http.StatusCreated)

}
