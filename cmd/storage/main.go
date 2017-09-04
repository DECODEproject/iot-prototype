package main

import (
	"os"

	"github.com/DECODEproject/iot-prototype/cmd"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
	Use: "storage",
}

var _config = newConfig()

type config struct {
	Binding             string `envconfig:"BINDING" default:":8083"`
	SwaggerUIPath       string `envconfig:"SWAGGER_UI_PATH" default:"../../swagger-ui/"`
	WebServicesURL      string `envconfig:"WEBSERVICES_URL" default:"http://localhost:8083"`
	RedisNetworkAddress string `envconfig:"REDIS_SERVICE_ADDRESS" default:":6379"`
}

func newConfig() *config {
	c := &config{}
	envconfig.Process("", c)
	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Binding, "binding", "b", o.Binding, "binding address in form of {ip}:port")
	fs.StringVarP(&o.SwaggerUIPath, "swagger-ui", "s", o.SwaggerUIPath, "path to folder to server Swagger UI")
	fs.StringVarP(&o.WebServicesURL, "url", "u", o.WebServicesURL, "external address of the API service")
	fs.StringVarP(&o.RedisNetworkAddress, "redis-address", "r", o.RedisNetworkAddress, "network address of the redis server")
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
