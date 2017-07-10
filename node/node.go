package node

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/services"
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

	UseTLS            bool
	CertFilePath      string
	KeyFilePath       string
	TrustedCAFilePath string
	LogFile           bool
	LogPath           string
	Syslog            bool
}

func Serve(options Options) error {

	metadataClient := metadataclient.NewMetadataApiWithBasePath(options.MetadataServiceAddress)
	storageClient := storageclient.NewDataApiWithBasePath(options.StorageServiceAddress)

	log.Print(options.StorageServiceAddress)
	log.Printf("registering %s with metadata service %s", options.WebServicesURL, options.MetadataServiceAddress)

	// TODO : check for existing token
	// If found then update the location by telling the metadata service where I am
	// The prototype will just register again
	token, err := registerWithMetadataService(metadataClient, options.WebServicesURL)

	if err != nil {
		return err
	}

	log.Printf("registered with metadata service : %s", token)

	store := services.NewEntitlementStore()

	// TODO : add service to receive data from the device hub and/or any other service
	restful.DefaultContainer.Add(services.NewEntitlementService(store).WebService())
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

	// start up a pretend device-hub input
	func() {

		go pretendToBeADeviceHubEndpoint(token, metadataClient, storageClient, store)

	}()

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
		response, _, err := client.RegisterLocation(metadataclient.ServicesLocationRequest{
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

func pretendToBeADeviceHubEndpoint(locationToken string, mClient *metadataclient.MetadataApi, sClient *storageclient.DataApi, entitlements *services.EntitlementStore) {

	// for every item from device-hub
	values := map[string]interface{}{
		"temp":     23.3,
		"humidity": 34,
	}
	/*
		schema := map[string]interface{}{
			"@context": map[string]interface{}{
				"decode":   "http://decode.eu#",
				"m3-lite":  "http://purl.org/iot/vocab/m3-lite#",
				"humidity": "m3-lite:AirHumidity",
				"temp":     "m3-lite:AirTemperature",
				"domain":   "decode:hasDomain",
			},
			"@type": "m3-lite:Sensor",
			"domain": map[string]interface{}{
				"@type": "m3-lite:Environment",
			},
		}
	*/
	// qualify values to unique paths maybe including json-ld plus something else...
	values = map[string]interface{}{
		"data://private/sensor-1/temp":     23.3,
		"data://private/sensor-1/humidity": 34,
	}

	// set up the entitlements we will use in the hard coded example
	entitlements.Accepted.Add(services.Entitlement{
		EntitlementRequest: services.EntitlementRequest{
			Subject:     "data://private/sensor-1/temp",
			AccessLevel: services.CanDiscover},
		UID: "abc",
	})
	entitlements.Accepted.Add(services.Entitlement{
		EntitlementRequest: services.EntitlementRequest{
			Subject:     "data://private/sensor-1/humidity",
			AccessLevel: services.CanDiscover},
		UID: "def",
	})

	log.Print("entitlements", entitlements.Accepted)

	// break down to individual key value pairs
	for k, _ := range values {

		// find entitlement for subject
		ent, found := entitlements.Accepted.FindForSubject(k)
		fmt.Println(ent, found)

		if found {

			// if the underlying data is accessible
			// send to the metadata service
			if ent.IsAccessible() {
				r, _, err := mClient.CatalogItem(metadataclient.ServicesItem{
					Example:     "some example",
					Key:         k,
					LocationUid: locationToken,
					// Is this the expanded view???
					Tags: []string{
						"one", "two", "three",
					},
				})

				log.Print(r, err)

				if err != nil {
					log.Print("error updating metadata : ", err.Error())
				}
			}
		}

		// write to the storage service
		_, err := sClient.Append(storageclient.ServicesData{Bucket: k, Value: "TODO"})

		if err != nil {
			log.Print("error appending to storage : ", err.Error())
		}
	}
}
