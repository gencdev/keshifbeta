package cmd

import (
	"localsd/keshif"

	"github.com/spf13/cobra"
)

var clear bool
var envoy bool

func init() {
	startCmd.PersistentFlags().BoolVarP(&clear, "clear", "c", false, "Clear vhosts after keshif closed")
	startCmd.PersistentFlags().BoolVarP(&envoy, "envoy", "e", false, "Use envoy as proxy")
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts keshif proxy.",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		keshif.Start(envoy, clear)
	},
}
