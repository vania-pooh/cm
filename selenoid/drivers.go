package selenoid

import (
	"os/exec"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
	"runtime"
	"github.com/emicklei/go-restful/log"
)

const (
	DefaultBrowsersXmlURL = "https://raw.githubusercontent.com/aerokube/cm/master/browsers.json"
	Firefox = "firefox"
	Chrome = "chrome"
	Opera = "opera"
	InternetExplorer = "internet_explorer"
	BrowserNames = []string{
		Firefox,
		Chrome,
		Opera,
		InternetExplorer,
	}
	Commands = map[string] string {
		Firefox: "firefox --version",
		Chrome: "google-chrome --version",
		Opera: "opera --version",
		InternetExplorer: `reg query "HKEY_LOCAL_MACHINE\Software\Microsoft\Internet Explorer" /v svcVersion`,
	}
)

type Browsers struct {
	Browsers map[string] Versions `json:"browsers"`
	Drivers map[string] Platforms `json:"drivers"`
}

type Versions map[string] string

type Platforms map[string] Architectures

type Architectures map[string] string

func Configure() error {
	browsers, err := loadAvailableBrowsers()
	if (err != nil) {
		return fmt.Errorf("failed to configure: %v\n", err)
	}
	for browserName, url := range determineDriverUrls(browsers) {
		log.Printf("Downloading %s driver from %s...\n", browserName, url)
		err := downloadDriver(url)
		if (err != nil) {
			return fmt.Errorf("failed to download driver: %v\n", err)
		}
	}
	//TODO: generate config file here!
	return nil
}

func loadAvailableBrowsers() (*Browsers, error) {
	data, err := downloadFile(DefaultBrowsersXmlURL)
	if (err != nil) {
		return nil, fmt.Errorf("browsers data download error: %v\n", err)
	}
	var browsers Browsers
	err = json.Unmarshal(data, browsers)
	if (err != nil) {
		return nil, fmt.Errorf("browsers data read error: %v\n", err)
	}
	return browsers, nil
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if (err != nil) {
		return nil, fmt.Errorf("file download error: %v\n", err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	return data, nil
}

func downloadDriver(url string) error {
	//TODO: to be implemented, use downloadFile() inside
	//TODO: includes extracting driver archive...
	return nil
}

func determineDriverUrls(browsers *Browsers) map[string]string {
	ret := make(map[string] string)
	for _, browserName := range BrowserNames {
		if driverUrl := getDriverUrl(browserName, browsers); driverUrl != "" {
			ret[browserName] = driverUrl
		}
	}
	return ret
}

func getDriverUrl(browserName string, browsers *Browsers) string {
	version := getBrowserVersion(browserName)
	if (version != "") {
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		versions, _ := browsers.Browsers[browserName]
		for v, driverKey := range versions {
			if (strings.Contains(version, v)) {
				if platforms, ok := browsers.Drivers[driverKey]; ok {
					if architectures, ok := platforms[goos]; ok {
						if url, ok := architectures[goarch]; ok {
							return url
						}
					}
				} else {
					log.Printf("Unsupported driver key: %s. This is probably a bug.\n", driverKey)
				}
			}
		}
		log.Printf("Skipping unsupported browser: %s %s %s %s\n", browserName, version, goos, goarch)
	}
	return ""
}

func getBrowserVersion(browserName string) string {
	return runCommand(Commands[browserName])
}

func runCommand(command string) string {
	output, err := exec.Command(command).Output()
	if (err != nil) {
		return ""
	}
	return output
}