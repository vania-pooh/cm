package cmd

import (
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	selenoidCleanupCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidCleanupCmd.Flags().StringVarP(&outputDir, "output-dir", "c", getSelenoidConfigDir(), "directory to remove")
}

var selenoidCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove Selenoid traces",
	Run: func(cmd *cobra.Command, args []string) {
		config := selenoid.LifecycleConfig{
			Quiet:     quiet,
			OutputDir: outputDir,
		}
		lifecycle, err := selenoid.NewLifecycle(&config)
		if err != nil {
			stderr("Failed to initialize: %v\n", err)
			os.Exit(1)
		}

		err = lifecycle.Stop()
		if err != nil {
			stderr("Failed to stop Selenoid: %v\n", err)
			os.Exit(1)
		}

		err = os.RemoveAll(outputDir)
		if err != nil {
			stderr("Failed to remove configuration directory: %v\n", err)
			os.Exit(1)
		}
		stderr("Successfully removed configuration directory\n")
		os.Exit(0)
	},
}
