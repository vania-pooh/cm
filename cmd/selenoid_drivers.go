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
	skipDownload bool
)

func init() {
	selenoidDriversCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidDriversCmd.Flags().StringVarP(&browsers, "browsers", "b", "", "comma separated list of browser names to process")
	selenoidDriversCmd.Flags().StringVarP(&browsersJSONUrl, "browsers-json", "j", defaultBrowsersJsonURL, "browsers JSON data URL (in most cases never need to be set manually)")
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Failed to determine current user")
		os.Exit(1)
	}
	selenoidDriversCmd.Flags().StringVarP(&configDir, "config-dir", "c", filepath.Join(usr.HomeDir, ".aerokube", "selenoid"), "directory to save configuration and driver binaries")
	selenoidDriversCmd.Flags().BoolVarP(&skipDownload, "no-download", "n", false, "whether to skip downloading drivers for installed browsers")
}

var selenoidDriversCmd = &cobra.Command{
	Use:   "drivers",
	Short: "Download drivers and generate JSON configuration for Selenoid without Docker",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := selenoid.NewDriversConfigurator(configDir, browsers, browsersJSONUrl, !skipDownload, quiet)

		browsers := cfg.Configure()

		if browsers == nil {
			os.Exit(1)
		}

		data, err := json.MarshalIndent(*browsers, "", "    ")
		if err != nil {
			cfg.Printf("failed to output Selenoid config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		os.Exit(0)
	},
}
