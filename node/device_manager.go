package node

import (
	"context"
	"fmt"
	"log"

	metadataclient "gogs.dyne.org/DECODE/decode-prototype-da/client/metadata"
	storageclient "gogs.dyne.org/DECODE/decode-prototype-da/client/storage"

	"github.com/kazarena/json-gold/ld"
	"gogs.dyne.org/DECODE/decode-prototype-da/node/api"
	"gogs.dyne.org/DECODE/decode-prototype-da/node/sensors"
)

type device_manager struct {
	mClient          *metadataclient.MetadataApi
	sClient          *storageclient.DataApi
	entitlementStore *api.EntitlementStore
	sensorMessages   chan sensors.SensorMessage
	ctx              context.Context
	locationToken    string
}

func NewDeviceManager(ctx context.Context,
	locationToken string,
	mClient *metadataclient.MetadataApi,
	sClient *storageclient.DataApi,
	entitlementStore *api.EntitlementStore) *device_manager {
	return &device_manager{
		ctx:              ctx,
		locationToken:    locationToken,
		mClient:          mClient,
		sClient:          sClient,
		entitlementStore: entitlementStore,
		sensorMessages:   make(chan sensors.SensorMessage),
	}
}

func (d *device_manager) Out() chan sensors.SensorMessage {
	return d.sensorMessages
}

func (d *device_manager) Start() {
	go d.loop()
}

func (d *device_manager) loop() {

	for {
		select {
		case message := <-d.sensorMessages:

			// for each bit of data
			// find an entitlement for the data
			// - if entitlement exists and IsDiscoverable() send metadata to the 'metadata' service
			// Always write data values to the 'storage' service
			for k, v := range message.Data {

				subject := buildSubjectKey(message.SensorUID, k)
				log.Println(subject)
				// find entitlement for subject
				ent, found := d.entitlementStore.Accepted.FindForSubject(subject)

				if found {

					// if the underlying data is discoverable
					// send to the metadata service
					if ent.IsDiscoverable() {
						err := d.sendDataToMetadataService(message.Schema, subject, k, v)

						if err != nil {
							log.Println(err.Error())
							continue
						}

					}
				}
				// write to the storage service
				err := d.sendDataToStorageService(subject, v)

				if err != nil {
					log.Println(err.Error())
					continue
				}
			}
		}
	}
}

func (d *device_manager) sendDataToStorageService(subject string, value interface{}) error {

	_, err := d.sClient.Append(storageclient.ApiData{Bucket: subject, Value: value})

	if err != nil {
		return fmt.Errorf("error appending to storage : %s ", err.Error())
	}

	return nil
}

func (d *device_manager) sendDataToMetadataService(schema map[string]interface{}, subject, key string, value interface{}) error {

	// we first need to use the schema for the data to 'expand' out and fully qualify the metadata
	// to do this we use the JSON-LD expand function that helpfully drops any unqualified metadata and values
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	// we need to make a copy of the schema as we don't want to mutate the existing values or schemas
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

	_, _, err = d.mClient.CatalogItem(d.locationToken, req)

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
