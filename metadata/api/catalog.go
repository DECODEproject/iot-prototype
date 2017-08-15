package api

import (
	"log"
	"net/http"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

// catalogResource is an 'in-memory' instance of a metadata service
type catalogResource struct {
	store *MetadataStore
}

// ErrorResponse signals error messages back to the client
type ErrorResponse struct {
	Error string `json:"error" description:"error message if any"`
}

// CatalogRequest contains the information required to register some data as being available at some location
// The Tags property should contain enough information to enable a search index
// The Sample property should contain enough detail to interact with the data
type CatalogRequest struct {
	Subject string   `json:"subject" description:"path of the data item" validate:"nonzero"`
	Tags    []string `json:"tags" description:"a collection of tags probably belonging to an ontology" validate:"nonzero"`
	Sample  string   `json:"sample" description:"sample value e.g. a json object `
}

// CatalogItem contains the original request
type CatalogItem struct {
	CatalogRequest
	LocationUID string `json:"-"`
}

// ItemWithLocation contains the item metadata and its location
type ItemWithLocation struct {
	CatalogItem
	Location Location `json:"location" description:"location for the catalogued piece of data" validate:"nonzero"`
}

// LocationRequest allows a node to register its presence with the service
type LocationRequest struct {
	IPAddress string `json:"ip-address" description:"public IP address of the node" validate:"nonzero"`
	Port      int    `json:"port" description:"public port of the node" validate:"nonzero"`
	Scheme    string `json:"scheme" description:protocol to use e.g. http or https`
}

// Location contains the original request and a UID to use when interacting with the service e.g. adding Items to the catalog.
type Location struct {
	LocationRequest
	UID string `json:"uid" description:"unique identifier for a node" validate:"nonzero"`
}

func NewCatalogService(store *MetadataStore) catalogResource {
	return catalogResource{
		store: store,
	}
}

func (e catalogResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/catalog").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	catalogUIDParameter := ws.PathParameter("catalog-uid", "identifier for a cataloged item").DataType("string")
	locationUIDParameter := ws.PathParameter("location-uid", "identifier for a location").DataType("string")

	tags := []string{"metadata"}

	// register a node at a location
	ws.Route(ws.PUT("/announce").To(e.registerLocation).
		Doc("register a node's location").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(LocationRequest{}).
		Returns(http.StatusOK, "OK", Location{}).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// move a location
	ws.Route(ws.PATCH("/announce/{location-uid}").To(e.moveLocation).
		Param(locationUIDParameter).
		Doc("change a node's location - keeping the same location-uid").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(LocationRequest{}).
		Returns(http.StatusOK, "OK", Location{}).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// add an item to the catalog
	ws.Route(ws.PUT("/items/{location-uid}").To(e.catalogItem).
		Doc("catalog an item for discovery e.g. what and where").
		Param(locationUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(CatalogRequest{}).
		Returns(http.StatusOK, "OK", CatalogItem{}).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// delete an item from the catalog
	ws.Route(ws.DELETE("/items/{catalog-uid}").To(e.removeFromCatalog).
		Doc("delete an item from the catalog").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", nil).
		Param(catalogUIDParameter))

	// get all items - simple search
	ws.Route(ws.GET("/items/").To(e.allItems).
		Doc("get all cataloged items").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", []ItemWithLocation{}))

	return ws
}

func (e catalogResource) registerLocation(request *restful.Request, response *restful.Response) {

	req := LocationRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	location := Location{
		LocationRequest: req,
		UID:             uuid.NewV4().String(),
	}
	e.store.Locations.Add(location)
	log.Print("node registered :", location)

	response.WriteEntity(location)

}

func (e catalogResource) moveLocation(request *restful.Request, response *restful.Response) {

	locationUID := request.PathParameter("location-uid")

	req := LocationRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	location := Location{
		LocationRequest: req,
		UID:             locationUID,
	}

	err := e.store.Locations.Replace(locationUID, location)

	if err != nil {
		if err == ErrLocationNotExists {
			response.WriteHeader(http.StatusNotFound)
			return

		}
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return

	}

	log.Print("node moved :", location)

	response.WriteEntity(location)

}

func (e catalogResource) catalogItem(request *restful.Request, response *restful.Response) {

	req := CatalogRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	locationUID := request.PathParameter("location-uid")

	found := e.store.Locations.Exists(locationUID)

	if !found {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: "unknown node"})
		return
	}

	item := CatalogItem{
		CatalogRequest: req,
		LocationUID:    locationUID,
	}
	e.store.Items.Add(item)

	response.WriteEntity(item)
}

func (e catalogResource) removeFromCatalog(request *restful.Request, response *restful.Response) {
	uid := request.PathParameter("catalog-uid")
	e.store.Items.Delete(uid)
}

func (e catalogResource) allItems(request *restful.Request, response *restful.Response) {
	list := e.store.All()
	response.WriteEntity(list)
}
