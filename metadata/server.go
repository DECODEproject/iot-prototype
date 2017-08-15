package metadata

import (
	"log"
	"net/http"

	"gogs.dyne.org/DECODE/decode-prototype-da/metadata/api"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

type Options struct {
	Binding        string
	SwaggerUIPath  string
	WebServicesURL string
	AssetsPath     string
}

func Serve(options Options) error {

	store := api.NewMetadataStore()

	restful.DefaultContainer.Add(api.NewCatalogService(store).WebService())

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

	// Serve the ui
	http.Handle("/", http.FileServer(http.Dir(options.AssetsPath)))

	log.Printf("start listening on %s", options.Binding)
	return http.ListenAndServe(options.Binding, nil)
}
