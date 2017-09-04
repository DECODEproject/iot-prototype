package api

import (
	"log"
	"net/http"

	validator "gopkg.in/validator.v2"

	metadataclient "github.com/DECODEproject/iot-prototype/client/metadata"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

// ErrorResponse signals error messages back to the client
type ErrorResponse struct {
	Error string `json:"error" description:"error message if any"`
}

// EntitlementRequest are made to request some access to a bit of data
type EntitlementRequest struct {
	Subject     string      `json:"subject" description:"path of the data e.g. data://user/email" validate:"nonzero"`
	AccessLevel AccessLevel `json:"level" description:"access level requested. Valid values 'owner-only', 'can-read','can-discover'" validate:"nonzero"`
}

// Entitlement is returned to encapsulate the current status of the entitlement
type Entitlement struct {
	EntitlementRequest
	UID    string `json:"uid" description:"unique identifier of the entitlement request" validate:"nonzero"`
	Status Status `json:"status" description:"current status of the request. Can be either 'requested', 'accepted', 'declined' or 'revoked'" validate:"nonzero"`
}

// IsDiscoverable returns true if the presence of the data can be discovered
// For the purposes of the prototype this means the data will be sent to the metadata service
func (e Entitlement) IsDiscoverable() bool {
	return e.AccessLevel == CanAccess || e.AccessLevel == CanDiscover
}

// IsAccessible returns true if the data can be accessed e.g. viewed, collated etc
func (e Entitlement) IsAccessible() bool {
	return e.AccessLevel == CanAccess
}

type AccessLevel string

const (
	OwnerOnly   = AccessLevel("owner-only")
	CanAccess   = AccessLevel("can-access")
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
	store   *EntitlementStore
	mClient *metadataclient.MetadataApi
}

func NewEntitlementService(store *EntitlementStore, mClient *metadataclient.MetadataApi) entitlementResource {
	return entitlementResource{
		store:   store,
		mClient: mClient,
	}
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
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// get a request by uid
	ws.Route(ws.GET("/requests/{request-uid}").To(e.findRequest).
		Doc("get an entitlement request").
		Param(requestUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(Entitlement{}).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusNotFound, "Not Found", nil))

	// delete a request
	// TODO : when adding in authN/R ensure only the creator can delete
	ws.Route(ws.DELETE("/requests/{request-uid}").To(e.removeRequest).
		Doc("delete an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(requestUIDParameter))

	// get all requests
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
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// decline request
	// TODO : review GET
	ws.Route(ws.GET("/requests/{request-uid}/decline").To(e.declineRequest).
		Param(requestUIDParameter).
		Doc("decline an entitlement request").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// entitlements
	// get all entitlements by path
	ws.Route(ws.GET("/accepted/").To(e.findEntitlements).
		Doc("get all accepted entitlements").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("subject", "filter by data path e.g. data://user/email").DataType("string")).
		Writes([]Entitlement{}).
		Returns(200, "OK", []Entitlement{}).
		Returns(404, "Not Found", nil))

	// update an existing accepted entitlement
	// TODO : when adding in authN/R ensure only the creator can modify
	ws.Route(ws.POST("/accepted/{entitlement-uid}").
		To(e.amendAcceptedEntitlement).
		Param(entitlementUIDParameter).
		Doc("amend an accepted entitlement.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Entitlement{}).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusBadRequest, "error validating request", ErrorResponse{}).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	// revoke an entitlement
	// NOT useful at the moment but when we get users it will revoke a specific users entitlement
	// TODO : review GET
	ws.Route(ws.GET("/accepted/{entitlement-uid}/revoke").
		To(e.revokeEntitlement).
		Doc("revoke an entitlement").
		Param(entitlementUIDParameter).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", Entitlement{}).
		Returns(http.StatusNotFound, "Not found", nil).
		Returns(http.StatusInternalServerError, "something went wrong", ErrorResponse{}))

	return ws
}

func (e entitlementResource) createRequest(request *restful.Request, response *restful.Response) {

	req := EntitlementRequest{}
	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	// TODO: Validate that I have data at that path
	// TODO : Need to validate AccessLevel
	entitlement := Entitlement{
		EntitlementRequest: req,
		UID:                uuid.NewV4().String(),
		Status:             Requested,
	}

	e.store.Requested.AppendOrReplaceOnSubject(entitlement)
	response.WriteEntity(entitlement)

}

func (e entitlementResource) findRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.store.Requested.Get(requestUID)

	if !found {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	response.WriteEntity(req)

}

func (e entitlementResource) removeRequest(request *restful.Request, response *restful.Response) {
	requestUID := request.PathParameter("request-uid")
	e.store.Requested.Delete(requestUID)
}

func (e entitlementResource) allRequests(request *restful.Request, response *restful.Response) {
	response.WriteEntity(e.store.Requested.All())
}

func (e entitlementResource) acceptRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.store.Requested.Get(requestUID)
	if !found {
		response.WriteHeader(http.StatusNotFound)
		return

	}

	e.store.Requested.Delete(requestUID)
	req.Status = Accepted
	// replace current one with the accepted entitlement
	e.store.Accepted.AppendOrReplaceOnSubject(req)

	response.WriteEntity(req)
}

func (e entitlementResource) declineRequest(request *restful.Request, response *restful.Response) {

	requestUID := request.PathParameter("request-uid")

	req, found := e.store.Requested.Get(requestUID)

	if !found {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	e.store.Requested.Delete(requestUID)
	req.Status = Declined
	// add to the list of declined entitlements
	e.store.Declined.Add(req)

	response.WriteEntity(req)
}

func (e entitlementResource) findEntitlements(request *restful.Request, response *restful.Response) {

	subject := request.QueryParameter("subject")

	list := []Entitlement{}
	for _, each := range e.store.Accepted.All() {

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

func (e entitlementResource) amendAcceptedEntitlement(request *restful.Request, response *restful.Response) {

	req := Entitlement{}
	if err := request.ReadEntity(&req); err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{Error: errs.Error()})
		return
	}

	uid := request.PathParameter("entitlement-uid")

	// we only allow updating the access level at the moment
	err := e.store.Accepted.Update(uid, func(e *Entitlement) error {
		e.AccessLevel = req.AccessLevel
		return nil
	})

	if err != nil {
		if err == ErrEntitlementNotFound {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// if the entitlement is now not discoverable we will need to
	// remove the data from the metadata service
	if !req.IsDiscoverable() {

		_, err = e.mClient.RemoveFromCatalog(req.Subject)

		if err != nil {
			log.Println("error removing from catalog : ", err.Error(), req.Subject)
			response.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})

			return
		}
	}

	response.WriteEntity(req)

}

func (e entitlementResource) revokeEntitlement(request *restful.Request, response *restful.Response) {

	uid := request.PathParameter("entitlement-uid")

	req, found := e.store.Accepted.Get(uid)

	if !found {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	e.store.Accepted.Delete(uid)
	req.Status = Revoked
	// add to the list of revoked entitlements
	e.store.Revoked.Add(req)

	response.WriteEntity(req)

}
