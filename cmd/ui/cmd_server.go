package main

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start ui server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		http.Handle("/", http.FileServer(http.Dir(_config.AssetsPath)))

		log.Printf("Serving %s on HTTP port: %s\n", _config.AssetsPath, _config.Binding)
		return http.ListenAndServe(_config.Binding, nil)
	},
}
