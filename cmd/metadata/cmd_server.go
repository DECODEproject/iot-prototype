package main

import (
	"github.com/DECODEproject/iot-prototype/metadata"
	"github.com/spf13/cobra"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start metadata server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return metadata.Serve(metadata.Options{
			Binding:        _config.Binding,
			SwaggerUIPath:  _config.SwaggerUIPath,
			WebServicesURL: _config.WebServicesURL,
			AssetsPath:     _config.AssetsPath,
		})

	},
}
