package cmd

import (
	"fmt"
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	selenoidStopCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
}

var selenoidStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Selenoid",
	Run: func(cmd *cobra.Command, args []string) {
		config := selenoid.LifecycleConfig{
			Quiet: quiet,
		}
		lifecycle, err := selenoid.NewLifecycle(&config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize: %v\n", err)
			os.Exit(1)
		}
		err = lifecycle.Stop()
		if err != nil {
			lifecycle.Printf("failed to stop Selenoid: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
