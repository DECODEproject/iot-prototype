package main

import (
	"github.com/spf13/cobra"
	"gogs.dyne.org/DECODE/decode-prototype-da/node"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start node.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return node.Serve(node.Options{
			Binding:                _config.Binding,
			SwaggerUIPath:          _config.SwaggerUIPath,
			WebServicesURL:         _config.WebServicesURL,
			MetadataServiceAddress: _config.MetadataServiceAddress,
			StorageServiceAddress:  _config.StorageServiceAddress,
			AssetsPath:             _config.AssetsPath,
		})

	},
}
