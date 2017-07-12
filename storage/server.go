package storage

import (
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/storage/services"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/garyburd/redigo/redis"
)

type Options struct {
	Binding             string
	SwaggerUIPath       string
	WebServicesURL      string
	RedisNetworkAddress string

	UseTLS            bool
	CertFilePath      string
	KeyFilePath       string
	TrustedCAFilePath string
	LogFile           bool
	LogPath           string
	Syslog            bool
}

func Serve(options Options) error {
	log.Print(options)

	redisConnection, err := redis.Dial("tcp", options.RedisNetworkAddress)

	if err != nil {
		return err
	}

	defer redisConnection.Close()

	restful.DefaultContainer.Add(services.NewDataService(redisConnection).WebService())

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