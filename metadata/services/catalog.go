package services

import restful "github.com/emicklei/go-restful"

type catalogResource struct {
}

type Item struct {
	Location string
	Key      string
	Schema   []string
}

func NewCatalogService() catalogResource {
	return catalogResource{}
}

func (e catalogResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/catalog").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

		// put an item

		// delete an item

	return ws
}
