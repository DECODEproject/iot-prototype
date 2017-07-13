package services

import (
	"log"
	"net/http"
	"sync"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

// catalogResource is an 'in-memory' instance of a metadata service
type catalogResource struct {
	lock      sync.RWMutex
	all       map[string]Item
	locations map[string]Location
}

// ErrorResponse signals error messages back to the client
type ErrorResponse struct {
	Error string `json:"error" description:"error message if any"`
}

// ItemRequest contains the information required to register some data as being available at some location
// The Tags property should contain enough information to enable a search index
// The Sample property should contain enough detail to interact with the data
type ItemRequest struct {
	LocationUID string   `json:"location-uid" description:"a valid location of a node registered previously" validate:"nonzero"`
	Key         string   `json:"key" description:"path of the data item" validate:"nonzero"`
	Tags        []string `json:"tags" description:"a collection of tags probably belonging to an ontology" validate:"nonzero"`
	Sample      string   `json:"sample" description:"sample value e.g. a json object `
}

// Item contains the original request
type Item struct {
	ItemRequest
}

// ItemWithLocation contains the item metadata and its location
type ItemWithLocation struct {
	Item
	Location
}

// LocationRequest allows a node to register its presence with the service
type LocationRequest struct {
	IPAddress string `json:"ip-address" description:"public IP address of the node" validate:"nonzero"`
	Port      int    `json:"port" description:"public port of the node" validate:"nonzero"`
}

// Location contains the original request and a UID to use when interacting with the service e.g. adding Items to the catalog.
type Location struct {
	LocationRequest
	UID string `json:"uid" description:"unique identifier for a node" validate:"nonzero"`
}

func NewCatalogService() catalogResource {
	return catalogResource{
		all:       map[string]Item{},
		locations: map[string]Location{},
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
	ws.Route(ws.PUT("/items").To(e.catalogItem).
		Doc("catalog an item for discovery e.g. what and where").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Item{}).
		Returns(http.StatusOK, "OK", Item{}).
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
		Writes([]ItemWithLocation{}))

	return ws
}

func (e catalogResource) registerLocation(request *restful.Request, response *restful.Response) {

	e.lock.Lock()
	defer e.lock.Unlock()

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
	e.locations[location.UID] = location

	log.Print("node registered :", location)

	response.WriteEntity(location)

}

func (e catalogResource) moveLocation(request *restful.Request, response *restful.Response) {

	e.lock.Lock()
	defer e.lock.Unlock()

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

	_, found := e.locations[locationUID]
	if !found {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	location := Location{
		LocationRequest: req,
		UID:             locationUID,
	}

	e.locations[location.UID] = location

	log.Print("node moved :", location)

	response.WriteEntity(location)

}

func (e catalogResource) catalogItem(request *restful.Request, response *restful.Response) {

	e.lock.Lock()
	defer e.lock.Unlock()

	req := ItemRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	_, found := e.locations[req.LocationUID]

	if !found {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: "unknown node"})
		return
	}

	item := Item{
		ItemRequest: req,
	}
	e.all[item.Key] = item

	log.Print(e.all)

	response.WriteEntity(item)
}

func (e catalogResource) removeFromCatalog(request *restful.Request, response *restful.Response) {

	e.lock.Lock()
	defer e.lock.Unlock()

	uid := request.PathParameter("catalog-uid")
	delete(e.all, uid)
}

func (e catalogResource) allItems(request *restful.Request, response *restful.Response) {

	e.lock.Lock()
	defer e.lock.Unlock()

	list := []ItemWithLocation{}

	for _, item := range e.all {

		location := e.locations[item.LocationUID]
		list = append(list, ItemWithLocation{
			item,
			location,
		})
	}

	response.WriteEntity(list)

}
