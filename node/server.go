package node

import (
	"context"
	"errors"
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/api"
	"gogs.dyne.org/DECODE/decode-prototype-da/utils"

	metadataclient "gogs.dyne.org/DECODE/decode-prototype-da/client/metadata"
	storageclient "gogs.dyne.org/DECODE/decode-prototype-da/client/storage"

	"github.com/cenkalti/backoff"
	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

type Options struct {
	Binding                string
	SwaggerUIPath          string
	WebServicesURL         string
	MetadataServiceAddress string
	StorageServiceAddress  string
}

func Serve(options Options) error {

	metadataClient := metadataclient.NewMetadataApiWithBasePath(options.MetadataServiceAddress)
	storageClient := storageclient.NewDataApiWithBasePath(options.StorageServiceAddress)

	log.Printf("registering %s with metadata service %s", options.WebServicesURL, options.MetadataServiceAddress)

	// TODO : reuse existing token
	// If found then update the location by telling the metadata service where I am
	// The prototype will just register again
	token, err := registerWithMetadataService(metadataClient, options.WebServicesURL)

	if err != nil {
		return err
	}

	log.Printf("registered with metadata service : %s", token)

	// entitlementStore holds an in-memory cache of entitlement data
	entitlementStore := api.NewEntitlementStore()
	// metaStore holds additional information about the data stored
	metaStore := api.NewMetadataStore()

	ctx := context.Background()
	deviceManager := NewDeviceManager(ctx, token, metadataClient, storageClient, entitlementStore, metaStore)
	deviceManager.Start()

	// wire up the json apis
	restful.DefaultContainer.Add(api.NewEntitlementService(entitlementStore, metaStore, metadataClient).WebService())
	restful.DefaultContainer.Add(api.NewDataService(entitlementStore, storageClient, metaStore).WebService())
	restful.DefaultContainer.Add(api.NewDeviceService(ctx, entitlementStore, metaStore, deviceManager.Out()).WebService())

	config := restfulspec.Config{
		WebServices:    restful.RegisteredWebServices(),
		WebServicesURL: options.WebServicesURL,
		APIPath:        "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
	}

	// install the cors filter
	restful.DefaultContainer.Filter(cors.Filter)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	http.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir(options.SwaggerUIPath))))
	log.Printf("start listening on %s", options.Binding)

	return http.ListenAndServe(options.Binding, nil)
}

// registerWithMetadataService returns the 'announce' token from the metadata service
func registerWithMetadataService(client *metadataclient.MetadataApi, nodePublicAddress string) (string, error) {

	// parse the nods public address into its component parts
	ok, host, port := utils.HostAndIpToBits(nodePublicAddress)

	if !ok {
		return "", errors.New("unable to parse WEBSERVICES_URL or flag -u. Expected value : http[s]://host:port")
	}

	// register with the metadata service using an exponential backoff
	var token string

	f := func() error {

		log.Printf(".")
		response, _, err := client.RegisterLocation(metadataclient.ApiLocationRequest{
			IpAddress: host,
			Port:      int32(port),
		})

		if err != nil {
			return err
		}
		token = response.Uid
		return nil
	}

	err := backoff.Retry(f, backoff.NewExponentialBackOff())
	return token, err
}
