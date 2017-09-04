package cmd

import (
	"fmt"

	prototype "github.com/DECODEproject/iot-prototype"
	"github.com/spf13/cobra"
)

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(prototype.VersionString())
		return nil
	},
}
