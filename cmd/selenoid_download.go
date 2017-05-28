package cmd

import (
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

func init() {
	selenoidDownloadCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidDownloadCmd.Flags().StringVarP(&outputDir, "config-dir", "c", getSelenoidOutputDir(), "directory to save files")
	selenoidDownloadCmd.Flags().StringVarP(&operatingSystem, "operating-system", "o", runtime.GOOS, "target operating system (drivers only)")
	selenoidDownloadCmd.Flags().StringVarP(&arch, "architecture", "a", runtime.GOARCH, "target architecture (drivers only)")
	selenoidDownloadCmd.Flags().StringVarP(&registry, "registry", "r", registryUrl, "Docker registry to use")
	selenoidDownloadCmd.Flags().StringVarP(&version, "version", "v", selenoid.Latest, "desired version; empty string for latest version")
	selenoidDownloadCmd.Flags().BoolVarP(&force, "force", "f", false, "force action")
}

var selenoidDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download Selenoid latest or specified release",
	Run: func(cmd *cobra.Command, args []string) {
		config := selenoid.LifecycleConfig{
			Quiet:       quiet,
			OutputDir:   outputDir,
			RegistryUrl: registry,
			OS:          operatingSystem,
			Arch:        arch,
			Version:     version,
		}
		lifecycle, err := selenoid.NewLifecycle(&config)
		if err != nil {
			stderr("Failed to initialize: %v\n", err)
			os.Exit(1)
		}
		err = lifecycle.Download()
		if err != nil {
			lifecycle.Printf("failed to download Selenoid release: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
