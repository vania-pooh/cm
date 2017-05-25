package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"path/filepath"
)

const (
	registryUrl            = "https://registry.hub.docker.com"
	defaultBrowsersJsonURL = "https://raw.githubusercontent.com/aerokube/cm/master/browsers.json"
)

var (
	lastVersions    int
	pull            bool
	tmpfs           int
	operatingSystem string
	arch            string
	version         string
	browsers        string
	browsersJSONUrl string
	outputDir       string
	skipDownload    bool
	force           bool
)

func init() {
	selenoidCmd.AddCommand(selenoidDownloadCmd)
	selenoidCmd.AddCommand(selenoidConfigureCmd)
	selenoidCmd.AddCommand(selenoidStartCmd)
	selenoidCmd.AddCommand(selenoidStopCmd)
	selenoidCmd.AddCommand(selenoidUpdateCmd)
	selenoidCmd.AddCommand(selenoidCleanupCmd)
}

var selenoidCmd = &cobra.Command{
	Use:   "selenoid",
	Short: "Download, configure and run Selenoid",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func getConfigDir(elem ...string) string {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to determine current user: writing data to current directory")
		return filepath.Join(append([]string{usr.HomeDir}, elem...)...)
	}
	return filepath.Join(elem...)
}

func getSelenoidConfigDir() string {
	return getConfigDir(".aerokube", "selenoid")
}

func stderr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a)
}
