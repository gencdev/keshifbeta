package cmd

import (
	"localsd/keshif"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear generated vhosts",
	Run: func(cmd *cobra.Command, args []string) {
		keshif.Clear()
	},
}
