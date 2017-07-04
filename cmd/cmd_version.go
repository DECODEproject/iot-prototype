package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	prototype "gogs.dyne.org/DECODE/decode-prototype-da"
)

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(prototype.VersionString())
		return nil
	},
}
