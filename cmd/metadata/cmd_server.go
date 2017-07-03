package main

import (
	"github.com/spf13/cobra"
	"gogs.dyne.org/DECODE/decode-prototype-da/metadata"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start metadata server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return metadata.Serve(metadata.Options{
			Binding:           _config.Binding,
			SwaggerUIPath:     _config.SwaggerUIPath,
			WebServicesURL:    _config.WebServicesURL,
			UseTLS:            _config.TLS,
			CertFilePath:      _config.CertFile,
			KeyFilePath:       _config.KeyFile,
			TrustedCAFilePath: _config.CACertFile,
			LogFile:           _config.LogFile,
			LogPath:           _config.LogPath,
			Syslog:            _config.Syslog,
		})

	},
}
