package services

import (
	"net/http"

	validator "gopkg.in/validator.v2"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

type catalogResource struct {
	all       map[string]Item
	locations map[string]Location
}

type CatalogRequest struct {
	LocationUID string   `json:"location-uid" description:"a valid location of a node registered previously via /announce" validate:"nonzero"`
	Key         string   `json:"key" description:"path of the data item" validate:"nonzero"`
	Tags        []string `json:"tags" description:"a collection of tags probably belonging to an ontology" validate:"nonzero"`
}

type Item struct {
	CatalogRequest
	UID string `json:"uid" description:"unique identifier for a metadata item" validate:"nonzero"`
}

type LocationRequest struct {
	IPAddress string `json:"ip-address" description:"public IP address of the node" validate:"nonzero"`
	Port      int    `json:"port" description:"public port of the node" validate:"nonzero"`
}

type Location struct {
	LocationRequest
	UID string `json:"uid" description:"unique identifier for a node" validate:"nonzero"`
}

type ItemWithLocation struct {
	Item
	Location
}

func NewCatalogService() catalogResource {
	return catalogResource{
		all:       map[string]Item{},
		locations: map[string]Location{},
	}
}

func (e catalogResource) registerLocation(request *restful.Request, response *restful.Response) {

	req := LocationRequest{}

	if err := request.ReadEntity(&req); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteErrorString(http.StatusBadRequest, errs.Error())
		return
	}

	location := Location{
		LocationRequest: req,
		UID:             uuid.NewV4().String(),
	}
	e.locations[location.UID] = location

	response.WriteEntity(location)

}

func (e catalogResource) moveLocation(request *restful.Request, response *restful.Response) {

	locationUID := request.PathParameter("location-uid")

	req := LocationRequest{}
	err := request.ReadEntity(&req)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteErrorString(http.StatusBadRequest, errs.Error())
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
	response.WriteEntity(location)

}

func (e catalogResource) catalogItem(request *restful.Request, response *restful.Response) {

	req := CatalogRequest{}
	err := request.ReadEntity(&req)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteErrorString(http.StatusBadRequest, errs.Error())
		return
	}

	_, found := e.locations[req.LocationUID]

	if !found {
		response.WriteErrorString(http.StatusInternalServerError, "unknown node")
		return
	}

	item := Item{
		CatalogRequest: req,
		UID:            uuid.NewV4().String(),
	}
	e.all[item.UID] = item

	response.WriteEntity(item)
}

func (e catalogResource) removeFromCatalog(request *restful.Request, response *restful.Response) {
	uid := request.PathParameter("catalog-uid")
	delete(e.all, uid)
}

func (e catalogResource) allItems(request *restful.Request, response *restful.Response) {

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
		Returns(http.StatusBadRequest, "error validating request", nil).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// move a location
	ws.Route(ws.PATCH("/announce/{location-uid}").To(e.moveLocation).
		Param(locationUIDParameter).
		Doc("register a node's location").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(LocationRequest{}).
		Returns(http.StatusOK, "OK", Location{}).
		Returns(http.StatusBadRequest, "error validating request", nil).
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// add an item to the catalog
	ws.Route(ws.PUT("/items").To(e.catalogItem).
		Doc("catalog an item for discovery e.g. what and where").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Item{}).
		Returns(http.StatusOK, "OK", Item{}).
		Returns(http.StatusBadRequest, "error validating request", nil).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// delete an item from the catalog
	ws.Route(ws.DELETE("/items/{catalog-uid}").To(e.removeFromCatalog).
		Doc("delete an item from the catalog").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(catalogUIDParameter))

	// get all items - simple search
	ws.Route(ws.GET("/items/").To(e.allItems).
		Doc("get all cataloged items").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]ItemWithLocation{}))

	return ws
}
