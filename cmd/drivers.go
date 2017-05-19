package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"path/filepath"
)

const (
	defaultBrowsersJsonURL = "https://raw.githubusercontent.com/aerokube/cm/master/browsers.json"
)

var (
	browsers        string
	browsersJSONUrl string
	configDir       string
	download        bool
)

func init() {
	driversCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	driversCmd.Flags().StringVarP(&browsers, "browsers", "b", "", "comma separated list of browser names to process")
	driversCmd.Flags().StringVarP(&browsersJSONUrl, "browsers-json", "j", defaultBrowsersJsonURL, "browsers JSON data URL (in most cases never need to be set manually)")
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Failed to determine current user")
		os.Exit(1)
	}
	driversCmd.Flags().StringVarP(&configDir, "config-dir", "c", filepath.Join(usr.HomeDir, ".aerokube", "selenoid"), "directory to save configuration and driver binaries")
	driversCmd.Flags().BoolVarP(&download, "download", "d", true, "whether to download drivers for installed browsers")
}

var driversCmd = &cobra.Command{
	Use:   "drivers",
	Short: "Download drivers and generate JSON configuration for Selenoid without Docker",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := selenoid.NewDriversConfigurator(configDir, browsers, browsersJSONUrl, download, quiet)

		browsers := cfg.Configure()

		if browsers == nil {
			os.Exit(1)
		}

		data, err := json.MarshalIndent(*browsers, "", "    ")
		if err != nil {
			fmt.Printf("Failed to output Selenoid config: %v", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		os.Exit(0)
	},
}
