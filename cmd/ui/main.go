package main

import (
	"os"

	"gogs.dyne.org/DECODE/decode-prototype-da/cmd"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
	Use: "ui",
}

var _config = newConfig()

type config struct {
	Binding    string `envconfig:"BINDING" default:":8085"`
	AssetsPath string `envconfig:"ASSETS_PATH" default:"../../ui/"`
}

func newConfig() *config {
	c := &config{}
	envconfig.Process("", c)
	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Binding, "binding", "b", o.Binding, "binding address in form of {ip}:port")
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