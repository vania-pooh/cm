package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	selenoidUpdateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidUpdateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", getSelenoidOutputDir(), "directory to save files")
}

var selenoidUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Selenoid (download latest Selenoid, configure and start)",
	Run: func(cmd *cobra.Command, args []string) {
		stderr("Updating Selenoid...\n")
		startImpl(true)
	},
}
