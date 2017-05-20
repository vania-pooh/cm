package cmd

import (
	"fmt"
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	operatingSystem string
	arch            string
	version         string
)

func init() {
	selenoidDownloadCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Failed to determine current user")
		os.Exit(1)
	}
	selenoidDownloadCmd.Flags().StringVarP(&configDir, "config-dir", "c", filepath.Join(usr.HomeDir, ".aerokube", "selenoid"), "directory to save Selenoid binary")
	selenoidDownloadCmd.Flags().StringVarP(&operatingSystem, "operatingSystem", "o", runtime.GOOS, "target operating system")
	selenoidDownloadCmd.Flags().StringVarP(&arch, "architecture", "a", runtime.GOARCH, "target architecture")
	selenoidDownloadCmd.Flags().StringVarP(&version, "version", "v", selenoid.Latest, "desired version; empty string for latest version")
}

var selenoidDownloadCmd = &cobra.Command{
	Use:   "donwload",
	Short: "Download Selenoid latest or specified release",
	Run: func(cmd *cobra.Command, args []string) {
		downloader := selenoid.NewDownloader("", configDir, operatingSystem, arch, version, quiet)
		err := downloader.Download()
		if err != nil {
			downloader.Printf("failed to download Selenoid release: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}
