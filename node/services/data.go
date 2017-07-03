package services

import restful "github.com/emicklei/go-restful"

type dataResource struct {
}

func NewDataService() dataResource {
	return dataResource{}
}

func (e dataResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/data").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	return ws
}
