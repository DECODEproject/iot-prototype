package api

import (
	"net/http"
	"time"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	redis_client "github.com/garyburd/redigo/redis"
	"gogs.dyne.org/DECODE/decode-prototype-da/utils"
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
		Doc("returns all of the data stored in a logical 'bucket'.").
		Param(ws.QueryParameter("from", "return data from this ISO8601 timestamp. Defaults to 24 hours ago.").DataType("date").DataFormat(utils.ISO8601)).
		Param(ws.QueryParameter("to", "finish at this ISO8601 timestamp ").DataType("date").DataFormat(utils.ISO8601)).
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

	fromStr := request.QueryParameter("from")
	toStr := request.QueryParameter("to")
	prefix := request.QueryParameter("bucket-uid")

	timestep := time.Second
	expiry := time.Duration(0)

	ts := NewTimeSeries(prefix, timestep, expiry, e.pool)
	var from, to time.Time
	var err error

	// TODO : review should this be UTC?
	now := time.Now()

	if fromStr == "" {
		// default to 24 hours ago
		from = now.Add(-(time.Hour * 24))
	} else {

		from, err = time.Parse(utils.ISO8601, fromStr)

		if err != nil {
			response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

	}
	if toStr == "" {
		//default to now
		to = now
	} else {
		to, err = time.Parse(utils.ISO8601, toStr)

		if err != nil {
			response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	data := []*DataResponse{}

	err = ts.FetchRange(from, to, &data)

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
	expiry := time.Duration(0)

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
