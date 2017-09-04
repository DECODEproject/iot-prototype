package storage

import (
	"log"
	"net/http"
	"time"

	"github.com/DECODEproject/iot-prototype/storage/api"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/garyburd/redigo/redis"
)

type Options struct {
	Binding             string
	SwaggerUIPath       string
	WebServicesURL      string
	RedisNetworkAddress string
}

func Serve(options Options) error {

	// create a connction pool for the redis backend
	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", options.RedisNetworkAddress)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	restful.DefaultContainer.Add(api.NewDataService(pool).WebService())

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
