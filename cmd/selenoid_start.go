package cmd

import (
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	selenoidStartCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidStartCmd.Flags().StringVarP(&outputDir, "config-dir", "c", getSelenoidOutputDir(), "directory to save files")
	selenoidStartCmd.Flags().BoolVarP(&force, "force", "f", false, "force action")
}

var selenoidStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Selenoid",
	Run: func(cmd *cobra.Command, args []string) {
		startImpl(force)
	},
}

func startImpl(force bool) {
	config := selenoid.LifecycleConfig{
		Quiet:     quiet,
		OutputDir: outputDir,
		Force:     force,
	}
	lifecycle, err := selenoid.NewLifecycle(&config)
	if err != nil {
		stderr("Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	err = lifecycle.Start()
	if err != nil {
		lifecycle.Printf("failed to start Selenoid: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
