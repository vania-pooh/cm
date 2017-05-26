package cmd

import (
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

func init() {
	selenoidConfigureCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidConfigureCmd.Flags().StringVarP(&outputDir, "output-dir", "o", getSelenoidOutputDir(), "directory to save files")
	selenoidConfigureCmd.Flags().StringVarP(&operatingSystem, "operating-system", "o", runtime.GOOS, "target operating system (drivers only)")
	selenoidConfigureCmd.Flags().StringVarP(&arch, "architecture", "a", runtime.GOARCH, "target architecture (drivers only)")
	selenoidConfigureCmd.Flags().StringVarP(&version, "version", "v", selenoid.Latest, "desired version; empty string for latest version")
	selenoidConfigureCmd.Flags().StringVarP(&browsers, "browsers", "b", "", "comma separated list of browser names to process")
	selenoidConfigureCmd.Flags().StringVarP(&browsersJSONUrl, "browsers-json", "j", defaultBrowsersJsonURL, "browsers JSON data URL (in most cases never need to be set manually)")
	selenoidConfigureCmd.Flags().BoolVarP(&skipDownload, "no-download", "n", false, "only output config file without downloading images or drivers")
	selenoidConfigureCmd.Flags().StringVarP(&registry, "registry", "r", registryUrl, "Docker registry to use")
	selenoidConfigureCmd.Flags().IntVarP(&lastVersions, "last-versions", "l", 5, "process only last N versions")
	selenoidConfigureCmd.Flags().BoolVarP(&pull, "pull", "p", false, "pull images if not present")
	selenoidConfigureCmd.Flags().IntVarP(&tmpfs, "tmpfs", "t", 0, "add tmpfs volume sized in megabytes")
	selenoidConfigureCmd.Flags().BoolVarP(&force, "force", "f", false, "force action")
}

var selenoidConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create Selenoid configuration file and download dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		config := selenoid.LifecycleConfig{
			Quiet:     quiet,
			OutputDir: outputDir,
			Force:     force,
			Download:  !skipDownload,
			OS:        operatingSystem,
			Arch:      arch,
			Version:   version,
		}
		lifecycle, err := selenoid.NewLifecycle(&config)
		if err != nil {
			stderr("Failed to initialize: %v\n", err)
			os.Exit(1)
		}
		err = lifecycle.Configure()
		if err != nil {
			lifecycle.Printf("failed to configure Selenoid: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
