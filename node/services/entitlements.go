package services

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

// EntitlementRequest are made to request some access to a bit of data
type EntitlementRequest struct {
	Subject     string      `json:"subject" description:"path of the data e.g. data://user/email"`
	AccessLevel AccessLevel `json:"level" description:"access level requested. Valid values 'none','can-read','can-discover'"`
}

//Entitlement is returned to encapsulate the current status of the entitlement
type Entitlement struct {
	EntitlementRequest
	UID    string `json:"uid" description:"unique identifier of the entitlement request"`
	Status Status `json:"status" description:"current status of the request. Can be either 'requested', 'accepted', 'declined' or 'revoked'"`
}

type AccessLevel string

const (
	None        = AccessLevel("none")
	CanRead     = AccessLevel("can-read")
	CanDiscover = AccessLevel("can-discover")
)

type Status string

const (
	Requested = Status("requested")
	Accepted  = Status("accepted")
	Declined  = Status("declined")
	Revoked   = Status("revoked")
)

type entitlementResource struct {
	// all data held in memory
	accepted  map[string]Entitlement
	declined  map[string]Entitlement
	requested map[string]Entitlement
	revoked   map[string]Entitlement
}

func NewEntitlementService() entitlementResource {
	return entitlementResource{
		accepted:  map[string]Entitlement{},
		declined:  map[string]Entitlement{},
		requested: map[string]Entitlement{},
		revoked:   map[string]Entitlement{},
	}
}

func (e entitlementResource) createRequest(request *restful.Request, response *restful.Response) {

	req := Entitlement{}
	err := request.ReadEntity(&req)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	// TODO: Validate that I have data at that path
	// TODO : Need to validate AccessLevel
	req.UID = uuid.NewV4().String()
	req.Status = Requested

	e.requested[req.UID] = req
	response.WriteEntity(req)

}

func (e entitlementResource) findRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.requested[requestUID]

	if !found {
		response.WriteErrorString(http.StatusNotFound, "request not found")
		return
	}

	response.WriteEntity(req)

}

func (e entitlementResource) removeRequest(request *restful.Request, response *restful.Response) {
	requestUID := request.PathParameter("request-uid")
	delete(e.requested, requestUID)
}

func (e entitlementResource) allRequests(request *restful.Request, response *restful.Response) {
	list := []Entitlement{}
	for _, each := range e.requested {
		list = append(list, each)
	}
	response.WriteEntity(list)
}

func (e entitlementResource) acceptRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.requested[requestUID]

	if !found {
		response.WriteErrorString(http.StatusNotFound, "request not found")
		return
	}

	delete(e.requested, requestUID)
	req.Status = Accepted
	e.accepted[req.UID] = req

	response.WriteEntity(req)

}

func (e entitlementResource) declineRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.requested[requestUID]

	if !found {
		response.WriteErrorString(http.StatusNotFound, "request not found")
		return
	}

	delete(e.requested, requestUID)
	req.Status = Declined
	e.declined[req.UID] = req

	response.WriteEntity(req)
}

func (e entitlementResource) findEntitlements(request *restful.Request, response *restful.Response) {

	subject := request.QueryParameter("subject")

	list := []Entitlement{}
	for _, each := range e.accepted {

		if subject == "" {

			list = append(list, each)
		} else {

			if subject == each.Subject {

				list = append(list, each)

			}
		}
	}

	response.WriteEntity(list)

}

func (e entitlementResource) revokeEntitlement(request *restful.Request, response *restful.Response) {

	uid := request.PathParameter("entitlement-uid")

	req, found := e.accepted[uid]

	if !found {
		response.WriteErrorString(http.StatusNotFound, "entitlement not found")
		return
	}

	delete(e.accepted, uid)
	req.Status = Revoked
	e.revoked[req.UID] = req

	response.WriteEntity(req)

}

func (e entitlementResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/entitlements").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"entitlements"}

	requestUIDParameter := ws.PathParameter("request-uid", "identifier of the entitlement response").DataType("string")
	entitlementUIDParameter := ws.PathParameter("entitlement-uid", "identifier of an accepted entitlement").DataType("string")

	// requests
	// make a request
	ws.Route(ws.PUT("/requests/").To(e.createRequest).
		Doc("create an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(EntitlementRequest{}).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// get a request by uid
	ws.Route(ws.GET("/requests/{request-uid}").To(e.findRequest).
		Doc("get an entitlement request").
		Param(requestUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(Entitlement{}).
		Returns(200, "OK", Entitlement{}).
		Returns(404, "Not Found", nil))

	// delete a request
	// TODO : when adding in authN/R ensure only the creator can delete
	ws.Route(ws.DELETE("/request/{request-uid}").To(e.removeRequest).
		Doc("delete an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(requestUIDParameter))

	// get all requests
	// TODO : add filter on status type
	// TODO : add filter on path
	ws.Route(ws.GET("/requests/").To(e.allRequests).
		Doc("get all pending requests").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]Entitlement{}))

	// accept request
	// TODO : review GET
	ws.Route(ws.GET("/requests/{request-uid}/accept").To(e.acceptRequest).
		Param(requestUIDParameter).
		Doc("accept an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// decline request
	// TODO : review GET
	ws.Route(ws.GET("/requests/{request-uid}/decline").To(e.declineRequest).
		Param(requestUIDParameter).
		Doc("decline an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// entitlements
	// get all entitlements by path
	ws.Route(ws.GET("/accepted/").To(e.findEntitlements).
		Doc("get all accepted entitlements").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("subject", "filter by data path e.g. data://user/email").DataType("string")).
		Writes([]Entitlement{}).
		Returns(200, "OK", []Entitlement{}).
		Returns(404, "Not Found", nil))

	// revoke an entitlement
	// TODO : review GET
	ws.Route(ws.GET("/accepted/{entitlement-uid}/revoke").To(e.revokeEntitlement).
		Doc("revoke an entitlement").
		Param(entitlementUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// TODO : add isEntitled method

	return ws
}
