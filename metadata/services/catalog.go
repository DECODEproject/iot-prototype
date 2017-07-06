package services

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	uuid "github.com/satori/go.uuid"
)

type catalogResource struct {
	all       map[string]Item
	locations map[string]Location
}

type CatalogRequest struct {
	LocationUID string
	Key         string
	Tags        []string
}

type Item struct {
	CatalogRequest
	UID string
}

type LocationRequest struct {
	UID       string
	IPAddress string
	Port      int
}

type Location struct {
	Location
	UID string
}

type ItemWithLocation struct {
	Item
	Location
}

func NewCatalogService() catalogResource {
	return catalogResource{
		all:      map[string]Item{},
		location: map[string]Location{},
	}
}

func (e catalogResource) registerLocation(request *restful.Request, response *restful.Response) {

	req := LocationRequest{}

}

func (e catalogResource) moveLocation(request *restful.Request, response *restful.Response) {
}

func (e catalogResource) catalogItem(request *restful.Request, response *restful.Response) {
	req := CatalogRequest{}
	err := request.ReadEntity(&req)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
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
}

func (e catalogResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/catalog").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	catalogUIDParameter := ws.PathParameter("catalog-uid", "identifier of a cataloged item").DataType("string")

	// register a node at a location
	ws.Route(ws.PUT("/announce").To(e.registerLocation).
		Doc("register a node's location").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(LocationRequest{}).
		Returns(http.StatusOK, "OK", Location{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// move a location

	// add an item to the catalog
	ws.Route(ws.PUT("/items").To(e.catalogItem).
		Doc("catalog an item for discovery e.g. what and where").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Item{}).
		Returns(http.StatusOK, "OK", Item{}).
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
