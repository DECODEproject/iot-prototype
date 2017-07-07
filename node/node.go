package node

import (
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/services"

	metadataclient "gogs.dyne.org/DECODE/decode-prototype-da/client/metadata"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

type Options struct {
	Binding                string
	SwaggerUIPath          string
	WebServicesURL         string
	MetadataServiceAddress string

	UseTLS            bool
	CertFilePath      string
	KeyFilePath       string
	TrustedCAFilePath string
	LogFile           bool
	LogPath           string
	Syslog            bool
}

func Serve(options Options) error {

	log.Printf("registering %s with metadata service %s", options.WebServicesURL, options.MetadataServiceAddress)

	// TODO : check for existing token
	// If found then update the location by telling the metadata service where I am
	// The path for the eprototype is to just register again
	token, err := registerWithMetadataService(options.MetadataServiceAddress, options.WebServicesURL)

	if err != nil {
		return err
	}

	log.Printf("registered with metadata service : %s", token)

	// TODO : add service to receive data from the device hub and/or any other service
	restful.DefaultContainer.Add(services.NewEntitlementService().WebService())
	restful.DefaultContainer.Add(services.NewFunctionService().WebService())

	config := restfulspec.Config{
		WebServices:    restful.RegisteredWebServices(),
		WebServicesURL: options.WebServicesURL,
		APIPath:        "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	http.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir(options.SwaggerUIPath))))

	log.Printf("start listening on %s", options.Binding)
	return http.ListenAndServe(options.Binding, nil)
}

func registerWithMetadataService(metadataServiceAddress, nodePublicAddress string) (string, error) {

	log.Printf("registering %s with metadata service %s", nodePublicAddress, metadataServiceAddress)

	// TODO : need to retry with backoff

	api := metadataclient.NewMetadataApiWithBasePath(metadataServiceAddress)
	response, _, err := api.RegisterLocation(metadataclient.ServicesLocationRequest{
		IpAddress: "localhost",
		Port:      8080,
	})

	if err != nil {
		return "", err
	}

	return response.Uid, nil
}
