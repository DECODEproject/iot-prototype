package api

import restful "github.com/emicklei/go-restful"

type functionResource struct {
}

func NewFunctionService() functionResource {
	return functionResource{}
}

func (e functionResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/funcs").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	return ws
}
