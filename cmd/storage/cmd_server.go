package main

import (
	"github.com/DECODEproject/iot-prototype/storage"
	"github.com/spf13/cobra"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start storage server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return storage.Serve(storage.Options{
			Binding:             _config.Binding,
			SwaggerUIPath:       _config.SwaggerUIPath,
			WebServicesURL:      _config.WebServicesURL,
			RedisNetworkAddress: _config.RedisNetworkAddress,
		})

	},
}
