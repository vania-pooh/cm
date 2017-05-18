package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/aerokube/cm/selenoid"
	"github.com/spf13/cobra"
	"os"
)

const (
	registryUrl = "https://registry.hub.docker.com"
)

var (
	lastVersions int
	pull         bool
	tmpfs        int
)

func init() {
	selenoidCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
	selenoidCmd.Flags().StringVarP(&registry, "registry", "r", registryUrl, "Docker registry to use")
	selenoidCmd.Flags().IntVarP(&lastVersions, "last-versions", "l", 5, "process only last N versions")
	selenoidCmd.Flags().BoolVarP(&pull, "pull", "p", false, "pull images if not present")
	selenoidCmd.Flags().IntVarP(&tmpfs, "tmpfs", "t", 0, "add tmpfs volume sized in megabytes")
}

var selenoidCmd = &cobra.Command{
	Use:   "selenoid",
	Short: "Generate JSON configuration for Selenoid",
	Run: func(cmd *cobra.Command, args []string) {

		//TODO: determine whether Docker is installed and use respective configurator
		cfg, err := selenoid.NewDockerConfigurator(registry, quiet)
		cfg.LastVersions = lastVersions
		cfg.Pull = pull
		cfg.Tmpfs = tmpfs
		if err != nil {
			fmt.Printf("Failed to initialize: %v\n", err)
			os.Exit(1)
		}
		defer cfg.Close()

		browsers := cfg.Configure()
		if err != nil {
			fmt.Printf("Failed to configure: %v", err)
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
