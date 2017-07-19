package node

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/node/api"
	"gogs.dyne.org/DECODE/decode-prototype-da/node/sensors"
	"gogs.dyne.org/DECODE/decode-prototype-da/utils"

	metadataclient "gogs.dyne.org/DECODE/decode-prototype-da/client/metadata"
	storageclient "gogs.dyne.org/DECODE/decode-prototype-da/client/storage"

	"github.com/cenkalti/backoff"
	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/kazarena/json-gold/ld"
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

	// TODO : reuse existing token
	// If found then update the location by telling the metadata service where I am
	// The prototype will just register again
	token, err := registerWithMetadataService(metadataClient, options.WebServicesURL)

	if err != nil {
		return err
	}

	log.Printf("registered with metadata service : %s", token)

	store := api.NewEntitlementStore()

	// TODO : add service to receive data from the device hub and/or any other service
	restful.DefaultContainer.Add(api.NewEntitlementService(store).WebService())
	restful.DefaultContainer.Add(api.NewDataService(store, storageClient).WebService())
	//	restful.DefaultContainer.Add(api.NewFunctionService().WebService())

	config := restfulspec.Config{
		WebServices:    restful.RegisteredWebServices(),
		WebServicesURL: options.WebServicesURL,
		APIPath:        "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
	}

	// install the cors filter
	restful.DefaultContainer.Filter(cors.Filter)

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

func pretendToBeADeviceHubEndpoint(locationToken string, mClient *metadataclient.MetadataApi, sClient *storageclient.DataApi, entitlements *api.EntitlementStore) {

	sensorMessages := make(chan sensors.SensorMessage)
	ctx := context.Background()

	// start up a few temp, humdity sensors
	one := sensors.NewTemperatureHumiditySensor(ctx, "sensor-1", sensorMessages)
	two := sensors.NewTemperatureHumiditySensor(ctx, "sensor-2", sensorMessages)

	// and a sine wave generator
	three := sensors.NewSineCurveEmitterSensor(ctx, "sine", sensorMessages)

	one.Start()
	two.Start()
	three.Start()

	// set up the entitlements we will use in the hard coded example
	entitlements.Accepted.Add(api.Entitlement{
		EntitlementRequest: api.EntitlementRequest{
			Subject:     buildSubjectKey("sensor-1", "temp"),
			AccessLevel: api.CanDiscover},
		UID: "abc",
	})
	entitlements.Accepted.Add(api.Entitlement{
		EntitlementRequest: api.EntitlementRequest{
			Subject:     buildSubjectKey("sensor-1", "humidity"),
			AccessLevel: api.CanAccess},
		UID: "def",
	})

	entitlements.Accepted.Add(api.Entitlement{
		EntitlementRequest: api.EntitlementRequest{
			Subject:     buildSubjectKey("sensor-2", "temp"),
			AccessLevel: api.CanDiscover},
		UID: "ghi",
	})
	entitlements.Accepted.Add(api.Entitlement{
		EntitlementRequest: api.EntitlementRequest{
			Subject:     buildSubjectKey("sensor-2", "humidity"),
			AccessLevel: api.CanAccess},
		UID: "klm",
	})

	entitlements.Accepted.Add(api.Entitlement{
		EntitlementRequest: api.EntitlementRequest{
			Subject:     buildSubjectKey("sine", "value"),
			AccessLevel: api.CanDiscover},
		UID: "nop",
	})

	for {
		select {
		case message := <-sensorMessages:

			// for each bit of data
			// find an entitlement for the data
			// - if entitlement exists and IsAccessible() send metadata to the 'metadata' service
			// Write data values to the 'storage' service
			for k, v := range message.Data {

				subject := buildSubjectKey(message.SensorUID, k)
				log.Println(subject)
				// find entitlement for subject
				ent, found := entitlements.Accepted.FindForSubject(subject)

				if found {

					// if the underlying data is discoverable
					// send to the metadata service
					if ent.IsDiscoverable() {
						err := sendDataToMetadataService(mClient, locationToken, message.Schema, subject, k, v)

						if err != nil {
							log.Println(err.Error())
							continue
						}

					}
				}
				// write to the storage service
				err := sendDataToStorageService(sClient, subject, v)

				if err != nil {
					log.Println(err.Error())
					continue
				}
			}

		}
	}
}

func sendDataToStorageService(sClient *storageclient.DataApi, subject string, value interface{}) error {

	_, err := sClient.Append(storageclient.ApiData{Bucket: subject, Value: value})

	if err != nil {
		return fmt.Errorf("error appending to storage : %s ", err.Error())
	}

	return nil
}

func sendDataToMetadataService(mClient *metadataclient.MetadataApi, locationToken string, schema map[string]interface{}, subject, key string, value interface{}) error {

	// we first need to use the schema for the data to 'expand' out and fully qualify the metadata
	// to do this we use the JSON-LD expand function that helpfully drops any unqualified metadata and values
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	// we need to make a copy of the schema as we don't want to mutate the existing vaules or schemas
	s := map[string]interface{}{}
	// copy schema
	for k2, v2 := range schema {
		s[k2] = v2
	}
	// add in the original value
	s[key] = value

	// and 'expand' out the data, schema
	expanded, err := proc.Expand(s, options)

	if err != nil {
		return err
	}

	// create our metadata request
	req := metadataclient.ApiCatalogRequest{
		Sample: fmt.Sprintf("%v", value), // TODO : respect the confidentiality
		Key:    subject,
		Tags:   harvestTagData("", expanded),
	}

	_, _, err = mClient.CatalogItem(locationToken, req)

	if err != nil {
		return fmt.Errorf("error updating metadata : %s", err.Error())
	}

	return nil
}

func buildSubjectKey(sensor, key string) string {
	return fmt.Sprintf("data://%s/%s", sensor, key)
}

func harvestTagData(parent string, v []interface{}) []string {

	r := []string{}

	for i, _ := range v {

		maybeMap := v[i]

		m, isMap := maybeMap.(map[string]interface{})
		if isMap {

			for k, v := range m {

				// does it have a '@type' annotation
				// if it does we add it
				if k == "@type" {

					v2, ok := v.([]interface{})

					if ok {
						r = append(r, fmt.Sprintf("%v", v2[0]))
					}
				} else {

					v2, ok := v.([]interface{})

					if ok {
						// down the rabbit hole we go...
						r = append(r, harvestTagData(k, v2)...)

						// if it is a child node with a value add the value's key
					} else if k == "@value" {
						r = append(r, parent)

					}
				}
			}
		}
	}

	return r
}
