package node

import (
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/services"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
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

func Serve(options Options) error {

	restful.DefaultContainer.Add(services.NewEntitlementService().WebService())
	restful.DefaultContainer.Add(services.NewDataService().WebService())
	restful.DefaultContainer.Add(services.NewFunctionService().WebService())

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
