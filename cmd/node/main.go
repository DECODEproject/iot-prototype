package main

import (
	"os"

	"gogs.dyne.org/DECODE/decode-prototype-da/cmd"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
	Use: "node",
}

var _config = newConfig()

type config struct {
	Binding                string `envconfig:"BINDING" default:":8080"`
	SwaggerUIPath          string `envconfig:"SWAGGER_UI_PATH" default:"../../swagger-ui/"`
	WebServicesURL         string `envconfig:"WEBSERVICES_URL" default:"http://localhost:8080"`
	MetadataServiceAddress string `envconfig:"METADATA_SERVICE_ADDRESS" default:"http://localhost:8081"`
	StorageServiceAddress  string `envconfig:"STORAGE_SERVICE_ADDRESS" default:"http://localhost:8083"`
	AssetsPath             string `envconfig:"ASSETS_PATH" default:"../../ui/"`
}

func newConfig() *config {
	c := &config{}
	envconfig.Process("", c)
	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Binding, "binding", "b", o.Binding, "binding address in form of {ip}:port")
	fs.StringVarP(&o.SwaggerUIPath, "swagger-ui", "w", o.SwaggerUIPath, "path to folder to server Swagger UI")
	fs.StringVarP(&o.WebServicesURL, "url", "u", o.WebServicesURL, "public address of the API service")
	fs.StringVarP(&o.MetadataServiceAddress, "metadata", "m", o.MetadataServiceAddress, "public address of the Metadata service")
	fs.StringVarP(&o.StorageServiceAddress, "storage", "s", o.StorageServiceAddress, "public address of the Storage service")
	fs.StringVarP(&o.AssetsPath, "assets", "a", o.AssetsPath, "path to folder of static files to serve")
}

func init() {
	RootCmd.AddCommand(serverCommand)
	RootCmd.AddCommand(cmd.VersionCommand)

	_config.AddFlags(RootCmd.PersistentFlags())
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
