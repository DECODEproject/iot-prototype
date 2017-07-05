package services

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

// EntitlementRequest are made to request some access to a bit of data
type EntitlementRequest struct {
	Path        string      `json:"path" description:"path of the data e.g. data://user/email"`
	AccessLevel AccessLevel `json:"level" description:"access level requested. Valid values 'none','can-read','can-discover'"`
}

//EntitlementResponse is returned to encapsulate the current status of the request
type EntitlementResponse struct {
	UID         string      `json:"uid" description:"unique identifier of the entitlement request"`
	Path        string      `json:"path" description:"path of the data e.g. data://user/email"`
	AccessLevel AccessLevel `json:"level" description:"access level requested. Valid values 'none','can-read','can-discover'"`
	Status      Status      `json:"status" description:"current status of the request. Can be either 'requested', 'accepted', 'declined' or 'revoked'"`
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
	accepted  map[string]EntitlementResponse
	declined  map[string]EntitlementResponse
	requested map[string]EntitlementResponse
	revoked   map[string]EntitlementResponse
}

func NewEntitlementService() entitlementResource {
	return entitlementResource{
		accepted:  map[string]EntitlementResponse{},
		declined:  map[string]EntitlementResponse{},
		requested: map[string]EntitlementResponse{},
		revoked:   map[string]EntitlementResponse{},
	}
}

func (e entitlementResource) createRequest(request *restful.Request, response *restful.Response) {

	req := EntitlementRequest{}
	err := request.ReadEntity(&req)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	// TODO: Validate that I have data at that path
	// TODO : Need to validate AccessLevel
	resp := EntitlementResponse{
		UID:         uuid.NewV4().String(),
		Path:        req.Path,
		AccessLevel: req.AccessLevel,
		Status:      Requested,
	}

	e.requested[resp.UID] = resp
	response.WriteEntity(resp)

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
	list := []EntitlementResponse{}
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

	path := request.QueryParameter("path")

	list := []EntitlementResponse{}
	for _, each := range e.accepted {

		if path != "" {

			list = append(list, each)
		} else {

			if path == each.Path {

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
		Returns(http.StatusOK, "OK", EntitlementResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// get a request by uid
	ws.Route(ws.GET("/requests/{request-uid}").To(e.findRequest).
		Doc("get an entitlement request").
		Param(requestUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(EntitlementResponse{}).
		Returns(200, "OK", EntitlementResponse{}).
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
		Writes([]EntitlementResponse{}))

	// accept request
	ws.Route(ws.PUT("/requests/{request-uid}/accept").To(e.acceptRequest).
		Doc("accept an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", EntitlementResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// decline request
	ws.Route(ws.PUT("/requests/{request-uid}/decline").To(e.declineRequest).
		Doc("decline an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", EntitlementResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// entitlements
	// get all entitlements by path
	ws.Route(ws.GET("/accepted/").To(e.findEntitlements).
		Doc("get all accepted entitlements").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("path", "filter by data path e.g. data://user/email").DataType("string")).
		Writes([]EntitlementResponse{}).
		Returns(200, "OK", []EntitlementResponse{}).
		Returns(404, "Not Found", nil))

	// revoke an entitlement
	ws.Route(ws.PUT("/accepted/{entitlement-uid}/revoke").To(e.revokeEntitlement).
		Doc("revoke an entitlement").
		Param(entitlementUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", EntitlementResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", nil))

	// TODO : add isEntitled method

	return ws
}
