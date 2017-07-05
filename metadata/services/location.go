package services

import restful "github.com/emicklei/go-restful"

type locationResource struct {
}

type Location struct {
	UID       string
	IPAddress string
	Port      int
}

func NewLocationService() locationResource {
	return locationResource{}
}

func (e locationResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/location").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

		// announce

		// move e.g. another IPAddress

	return ws
}
