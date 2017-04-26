package selenoid

import (
	"os/exec"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
	"runtime"
	"encoding/hex"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"errors"
	"path"
	"github.com/aerokube/selenoid/config"
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
		
		//TODO: use golang registry and _windows.go file https://godoc.org/golang.org/x/sys/windows/registry
		InternetExplorer: `reg query "HKEY_LOCAL_MACHINE\Software\Microsoft\Internet Explorer" /v svcVersion`,
	}
	
	ZipMagicHeader = "504b"
	GzipMagicHeader = "1f8b"
	
	ConfigDir = "~/.aerokube/selenoid"
)

type Browsers struct {
	Browsers map[string] Versions `json:"browsers"`
	Drivers map[string] Platforms `json:"drivers"`
}

type Versions map[string] string

type Platforms map[string] Architectures

type Architectures map[string] Driver

type Driver struct {
	URL string `json:"url"`
	Filename string `json:"filename"`
	Command string `json:"command"`
}

type existingDriver struct {
	BrowserName string
	Version string
	DriverPath  string
	Driver      *Driver
}

func Configure() (string, error) {
	browsers, err := loadAvailableBrowsers()
	if (err != nil) {
		return "", fmt.Errorf("failed to configure: %v\n", err)
	}
	configDir, err := prepareConfigDir()
	if (err != nil) {
		return fmt.Errorf("failed to prepare config dir: %v\n", err)
	}
	existingDrivers := getDrivers(browsers)
	for _, ed := range existingDrivers {
		log.Printf("Downloading %s %s driver from %s...\n", ed.BrowserName, ed.Version, ed.Driver.URL)
		driverPath, err := downloadDriver(ed.Driver, configDir)
		if (err != nil) {
			return "", fmt.Errorf("failed to download driver: %v\n", err)
		}
		ed.DriverPath = driverPath
	}
	return generateConfig(existingDrivers)
}

func generateConfig(existingDrivers []*existingDriver) (string, error) {
	browsers := make(map[string]config.Versions)
	for _, ed := range existingDrivers {
		driver := ed.Driver
		cmd := strings.Split(driver.Command, " ")
		versions := config.Versions{
			Default: ed.Version,
			Versions: map[string]*config.Browser{
				ed.Version: {
					Image: cmd,
					Port: 4444, //That's a convention
				},
			},
		}
		browsers[ed.BrowserName] = versions
	}
	data, err := json.Marshal(browsers)
	if (err != nil) {
		return "", fmt.Errorf("failed to generate config json: %v\n", err)
	}
	return string(data), nil
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

func prepareConfigDir() (string, error) {
	homeDir, err := homedir.Expand(ConfigDir)
	if (err != nil) {
		return "", fmt.Errorf("failed to determine config directory: %v\n", err)
	}
	err = os.MkdirAll(homeDir, os.ModePerm)
	if (err != nil) {
		return "", fmt.Errorf("failed to create config directory: %v\n", err)
	}
	return homeDir, nil
}

func downloadDriver(driver *Driver, dir string) (string, error) {
	data, err := downloadFile(driver.URL)
	if (err != nil) {
		return "", fmt.Errorf("failed to download driver archive: %v\n", err)
	}
	return extractFile(data, driver.Filename, dir)
}

func getMagicHeader(data []byte) string {
	if (len(data) >= 2) {
		return hex.EncodeToString(data[:2])
	}
	return ""
}

func isZipFile(data []byte) bool {
	return getMagicHeader(data) == ZipMagicHeader
}

func isGzipFile(data []byte) bool {
	return getMagicHeader(data) == GzipMagicHeader
}

func extractFile(data []byte, filename string, outputDir string) error {
	if isZipFile(data) {
		return unzip(data, filename, outputDir)
	} else if isGzipFile(data) {
		return gunzip(data, filename, outputDir)
	}
	return errors.New("Unknown archive type")
}

func unzip(data []byte, filename string, outputDir string) (string, error) {
	//TODO: to be implemented!
	return path.Join(outputDir, filename), nil
}

func gunzip(data []byte, filename string, outputDir string) (string, error) {
	//TODO: to be implemented!
	return path.Join(outputDir, filename), nil
}

func getDrivers(browsers *Browsers) []*existingDriver {
	ret := []*existingDriver{}
	for _, browserName := range BrowserNames {
		if existingDriver := getDriver(browserName, browsers); existingDriver != nil {
			ret = append(ret, existingDriver)
		}
	}
	return ret
}

func getDriver(browserName string, browsers *Browsers) *existingDriver {
	version := getBrowserVersion(browserName)
	if (version != "") {
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		versions, _ := browsers.Browsers[browserName]
		for v, driverKey := range versions {
			if (strings.Contains(version, v)) {
				if platforms, ok := browsers.Drivers[driverKey]; ok {
					if architectures, ok := platforms[goos]; ok {
						if driver, ok := architectures[goarch]; ok {
							return &existingDriver{
								BrowserName: browserName,
								Version: version,
								Driver: &driver,
							}
						}
					}
				} else {
					log.Printf("Unsupported driver key: %s. This is probably a bug.\n", driverKey)
				}
			}
		}
		log.Printf("Skipping unsupported browser: %s %s %s %s\n", browserName, version, goos, goarch)
	}
	return nil
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