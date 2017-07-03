package node

import (
	"log"
	"net/http"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	uuid "github.com/satori/go.uuid"
)

type Options struct {
	Binding           string
	UseTLS            bool
	CertFilePath      string
	KeyFilePath       string
	TrustedCAFilePath string
	LogFile           bool
	LogPath           string
	Syslog            bool
}

type Entitlement struct {
	UID  string `json:"uid" description:"unique identifier of the entitlement"`
	Path string `json:"path" description:"path of the data e.g. data://user/email"`
	Key  string `json:"key" description:"public key of the requester"`
}

type EntitlementResource struct {
	entitlements map[string]Entitlement
}

func (e EntitlementResource) find(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("entitlement-id")
	ent := e.entitlements[id]

	if len(ent.UID) == 0 { // bleurgg
		response.WriteErrorString(http.StatusNotFound, "Entitlement could not be found.")
	} else {
		response.WriteEntity(ent)
	}
}

func (e EntitlementResource) update(request *restful.Request, response *restful.Response) {

	ent := new(Entitlement)

	err := request.ReadEntity(&ent)

	if err == nil {
		e.entitlements[ent.UID] = *ent /// WTF
		response.WriteEntity(ent)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func (e EntitlementResource) create(request *restful.Request, response *restful.Response) {

	ent := Entitlement{UID: uuid.NewV4().String()}
	err := request.ReadEntity(&ent)
	if err == nil {
		e.entitlements[ent.UID] = ent
		response.WriteHeaderAndEntity(http.StatusCreated, ent)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}
func (e EntitlementResource) remove(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("entitlement-id")
	delete(e.entitlements, id)

}

func (e EntitlementResource) findAll(request *restful.Request, response *restful.Response) {
	list := []Entitlement{}
	for _, each := range e.entitlements {
		list = append(list, each)
	}
	response.WriteEntity(list)
}

func (e EntitlementResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/entitlements").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"entitlements"}

	ws.Route(ws.GET("/").To(e.findAll).
		// docs
		Doc("get all entitlements").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]Entitlement{}).
		Returns(200, "OK", []Entitlement{}))

	ws.Route(ws.GET("/{entitlement-uid}").To(e.find).
		// docs
		Doc("get an entitlement").
		Param(ws.PathParameter("entitlement-uid", "identifier of the entitlement").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(Entitlement{}). // on the response
		Returns(200, "OK", Entitlement{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PUT("/{entitlement-uid}").To(e.update).
		// docs
		Doc("update an entitlement").
		Param(ws.PathParameter("entitlement-uid", "identifier of the entitlement").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Entitlement{})) // from the request

	ws.Route(ws.PUT("").To(e.create).
		// docs
		Doc("create an entitlement").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(Entitlement{})) // from the request

	ws.Route(ws.DELETE("/{entitlement-uid}").To(e.remove).
		// docs
		Doc("delete an entitlement").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("entitlement-uid", "identifier of the entitlement").DataType("string")))

	return ws
}

func Serve(options Options) error {

	e := EntitlementResource{
		entitlements: map[string]Entitlement{},
	}

	restful.DefaultContainer.Add(e.WebService())

	config := restfulspec.Config{
		WebServices:    restful.RegisteredWebServices(),
		WebServicesURL: "http://localhost:8080",
		APIPath:        "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	http.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("./swagger-ui/"))))

	log.Printf("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	return nil
}
