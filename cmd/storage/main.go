package main

import (
	"os"

	"gogs.dyne.org/DECODE/decode-prototype-da/cmd"

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
	WebServicesURL      string `envconfig:"WEBSERVICES_URL" default:"http://localhost:8081"`
	RedisNetworkAddress string `envconfig:"REDIS_SERVICE_ADDRESS" default:":6379"`

	TLS        bool   `envconfig:"TLS"`
	ServerName string `envconfig:"TLS_SERVER_NAME"`
	CACertFile string `envconfig:"TLS_CA_CERT_FILE"`
	CertFile   string `envconfig:"TLS_CERT_FILE"`
	KeyFile    string `envconfig:"TLS_KEY_FILE"`

	DataImpl string `envconfig:"DATA_IMPL" default:"redis"`

	LogFile bool   `envconfig:"LOG_FILE"`
	LogPath string `envconfig:"LOG_PATH" default:"./decode_storage.log"`
	Syslog  bool   `envconfig:"LOG_SYSLOG"`
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

	fs.BoolVar(&o.TLS, "tls", o.TLS, "enable tls")
	fs.StringVar(&o.CACertFile, "tls-ca-cert-file", o.CACertFile, "ca certificate file")
	fs.StringVar(&o.CertFile, "tls-cert-file", o.CertFile, "client certificate file")
	fs.StringVar(&o.KeyFile, "tls-key-file", o.KeyFile, "client key file")

	fs.StringVar(&o.DataImpl, "data-impl", o.DataImpl, "datastore to use, valid values are 'boltdb' or 'filestore', defaults to boltdb")

	fs.BoolVar(&o.LogFile, "log-file", o.LogFile, "enable log to file")
	fs.StringVar(&o.LogPath, "log-path", o.LogPath, "path to log file, defaults to ./device-hub.log")
	fs.BoolVar(&o.Syslog, "log-syslog", o.Syslog, "enable log to local SYSLOG")
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
