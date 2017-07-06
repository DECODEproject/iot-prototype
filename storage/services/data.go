package services

import (
	"net/http"
	"time"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	redis_client "github.com/garyburd/redigo/redis"
	"gogs.dyne.org/DECODE/decode-prototype-da/storage/redis"
	"gogs.dyne.org/DECODE/decode-prototype-da/utils"
)

type Data struct {
	Value string `json:"value" description:"encoded contents to save" validate:"nonzero"`
}

type dataResource struct {
	db redis_client.Conn
}

func NewDataService(db redis_client.Conn) dataResource {
	return dataResource{
		db: db,
	}
}

func (e dataResource) getAll(request *restful.Request, response *restful.Response) {

	prefix := request.PathParameter("bucket-uid")
	timestep := time.Second
	expiry := time.Duration(0)

	ts := redis.NewTimeSeries(prefix, timestep, expiry, e.db)

	fromStr := request.QueryParameter("from")
	toStr := request.QueryParameter("to")

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
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}

	}
	if toStr == "" {
		//default to now
		to = now
	} else {
		to, err = time.Parse(utils.ISO8601, toStr)

		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
	}

	data := []*Data{}
	err = ts.FetchRange(from, to, &data)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(data)
	}
}

func (e dataResource) append(request *restful.Request, response *restful.Response) {

	prefix := request.PathParameter("bucket-uid")
	timestep := time.Second
	expiry := time.Duration(0)

	data := Data{}
	err := request.ReadEntity(&data)

	if err := request.ReadEntity(&data); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if errs := validator.Validate(data); errs != nil {
		response.WriteErrorString(http.StatusBadRequest, errs.Error())
		return
	}

	ts := redis.NewTimeSeries(prefix, timestep, expiry, e.db)
	err = ts.Add(&data)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteHeader(http.StatusCreated)
	}
}

func (e dataResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/data").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"data"}

	bucketUID := ws.PathParameter("bucket-uid", "name of the 'bucket' of data").DataType("string")
	now := time.Now()

	ws.Route(ws.GET("/{bucket-uid}").To(e.getAll).
		Doc("returns all of the data stored in a logical 'bucket'.").
		Param(bucketUID).
		Param(ws.QueryParameter("from", "return data from this ISO8601 timestamp. Defaults to 24 hours ago.").DataType("date").DataFormat(utils.ISO8601)).
		Param(ws.QueryParameter("to", "finish at this ISO8601 timestamp ").DataType("date").DataFormat(utils.ISO8601).DefaultValue(utils.ISO8601Time{now}.String())).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]Data{}). // on the response
		Returns(http.StatusOK, "OK", []Data{}).
		Returns(http.StatusNotFound, "Not Found", nil))

	ws.Route(ws.PUT("/{bucket-uid}").To(e.append).
		Doc("append data to a bucket, will create the bucket if it does not exist.").
		Param(bucketUID).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Data{}). // from the request
		Returns(http.StatusCreated, "Data was accepted.", nil).
		Returns(http.StatusBadRequest, "error validating request", nil).
		Returns(http.StatusInternalServerError, "Something went wrong", nil))

	return ws
}
