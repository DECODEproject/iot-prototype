package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
	Use: "node",
}

var _config = newConfig()

type config struct {
	Binding    string `envconfig:"BINDING" default:":8089"`
	TLS        bool   `envconfig:"TLS"`
	ServerName string `envconfig:"TLS_SERVER_NAME"`
	CACertFile string `envconfig:"TLS_CA_CERT_FILE"`
	CertFile   string `envconfig:"TLS_CERT_FILE"`
	KeyFile    string `envconfig:"TLS_KEY_FILE"`
	DataDir    string `envconfig:"DATA_DIR" default:"."`
	DataImpl   string `envconfig:"DATA_IMPL" default:"boltdb"`
	LogFile    bool   `envconfig:"LOG_FILE"`
	LogPath    string `envconfig:"LOG_PATH" default:"./decode_node.log"`
	Syslog     bool   `envconfig:"LOG_SYSLOG"`
}

func newConfig() *config {
	c := &config{}
	envconfig.Process("", c)
	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Binding, "binding", "b", o.Binding, "binding address in form of {ip}:port")
	fs.BoolVar(&o.TLS, "tls", o.TLS, "enable tls")
	fs.StringVar(&o.CACertFile, "tls-ca-cert-file", o.CACertFile, "ca certificate file")
	fs.StringVar(&o.CertFile, "tls-cert-file", o.CertFile, "client certificate file")
	fs.StringVar(&o.KeyFile, "tls-key-file", o.KeyFile, "client key file")
	fs.StringVar(&o.DataDir, "data-dir", o.DataDir, "path to db folder, defaults to current directory")
	fs.StringVar(&o.DataImpl, "data-impl", o.DataImpl, "datastore to use, valid values are 'boltdb' or 'filestore', defaults to boltdb")
	fs.BoolVar(&o.LogFile, "log-file", o.LogFile, "enable log to file")
	fs.StringVar(&o.LogPath, "log-path", o.LogPath, "path to log file, defaults to ./device-hub.log")
	fs.BoolVar(&o.Syslog, "log-syslog", o.Syslog, "enable log to local SYSLOG")
}

func init() {
	RootCmd.AddCommand(serverCommand)
	_config.AddFlags(RootCmd.PersistentFlags())
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
