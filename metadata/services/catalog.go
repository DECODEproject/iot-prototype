package services

import restful "github.com/emicklei/go-restful"

type catalogResource struct {
}

func NewCatalogService() catalogResource {
	return catalogResource{}
}

func (e dataResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/catalog").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	return ws
}
