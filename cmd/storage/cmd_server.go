package main

import (
	"github.com/spf13/cobra"
	"gogs.dyne.org/DECODE/decode-prototype-da/storage"
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
			UseTLS:              _config.TLS,
			CertFilePath:        _config.CertFile,
			KeyFilePath:         _config.KeyFile,
			TrustedCAFilePath:   _config.CACertFile,
			LogFile:             _config.LogFile,
			LogPath:             _config.LogPath,
			Syslog:              _config.Syslog,
		})

	},
}
