package selenoid

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aerokube/selenoid/config"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	zipMagicHeader  = "504b"
	gzipMagicHeader = "1f8b"
	comma           = ","
)

type Browsers map[string]Browser

type Browser struct {
	Command string `json:"command"`
	Files   Files  `json:"files"`
}

type Files map[string]Architectures

type Architectures map[string]Driver

type Driver struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type downloadedDriver struct {
	BrowserName string
	Command     string
}

type DriversConfigurator struct {
	Logger
	OutputDirAware
	Browsers        string
	BrowsersJsonUrl string
	Download        bool
}

func NewDriversConfigurator(outputDir string, browsers string, browsersJsonUrl string, download bool, quiet bool) *DriversConfigurator {
	return &DriversConfigurator{
		Logger:          Logger{Quiet: quiet},
		OutputDirAware:  OutputDirAware{OutputDir: outputDir},
		Browsers:        browsers,
		BrowsersJsonUrl: browsersJsonUrl,
		Download:        download,
	}
}

func (c *DriversConfigurator) Configure() *SelenoidConfig {
	browsers := c.loadAvailableBrowsers()
	if browsers == nil {
		return nil
	}
	err := c.createOutputDir()
	if err != nil {
		c.Printf("failed to create output directory: %v\n", err)
		return nil
	}
	downloadedDrivers := c.downloadDrivers(browsers, c.OutputDir)
	cfg := generateConfig(downloadedDrivers)
	return &cfg
}

func generateConfig(downloadedDrivers []downloadedDriver) SelenoidConfig {
	browsers := make(SelenoidConfig)
	for _, dd := range downloadedDrivers {
		cmd := strings.Fields(dd.Command)
		versions := config.Versions{
			Default: Latest,
			Versions: map[string]*config.Browser{
				Latest: {
					Image: cmd,
				},
			},
		}
		browsers[dd.BrowserName] = versions
	}
	return browsers
}

func (c *DriversConfigurator) loadAvailableBrowsers() *Browsers {
	jsonUrl := c.BrowsersJsonUrl
	c.Printf("downloading browser data from: %s\n", jsonUrl)
	data, err := downloadFile(jsonUrl)
	if err != nil {
		c.Printf("browsers data download error: %v\n", err)
		return nil
	}
	var browsers Browsers
	err = json.Unmarshal(data, &browsers)
	if err != nil {
		c.Printf("browsers data read error: %v\n", err)
		return nil
	}
	return &browsers
}

func downloadFile(url string) ([]byte, error) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := downloadFileWithProgressBar(url, w)
	if err != nil {
		return nil, err
	}
	w.Flush()
	return b.Bytes(), nil
}

func downloadFileWithProgressBar(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("file download error: %v", err)
	}
	defer resp.Body.Close()

	bar := pb.New(int(resp.ContentLength)).SetUnits(pb.U_BYTES)
	bar.Output = os.Stderr
	bar.Start()
	defer bar.Finish()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	writer := io.MultiWriter(w, bar)
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	return nil
}

func (c *DriversConfigurator) downloadDriver(driver *Driver, dir string) (string, error) {
	if c.Download {
		log.Printf("Downloading driver from %s...\n", driver.URL)
		data, err := downloadFile(driver.URL)
		if err != nil {
			return "", fmt.Errorf("failed to download driver archive: %v\n", err)
		}
		return extractFile(data, driver.Filename, dir)
	}
	return driver.Filename, nil
}

func getMagicHeader(data []byte) string {
	if len(data) >= 2 {
		return hex.EncodeToString(data[:2])
	}
	return ""
}

func isZipFile(data []byte) bool {
	return getMagicHeader(data) == zipMagicHeader
}

func isTarGzFile(data []byte) bool {
	return getMagicHeader(data) == gzipMagicHeader
}

func extractFile(data []byte, filename string, outputDir string) (string, error) {
	if isZipFile(data) {
		return unzip(data, filename, outputDir)
	} else if isTarGzFile(data) {
		return untar(data, filename, outputDir)
	}
	return "", errors.New("Unknown archive type")
}

// Based on http://stackoverflow.com/questions/20357223/easy-way-to-unzip-file-with-golang
func unzip(data []byte, fileName string, outputDir string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) (string, error) {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		outputPath := filepath.Join(outputDir, f.Name)

		if f.FileInfo().IsDir() {
			return "", fmt.Errorf("can only unzip files but %s is a directory", f.Name)
		}

		err = outputFile(outputPath, f.Mode(), rc)
		if err != nil {
			return "", err
		}
		return outputPath, nil
	}

	if err == nil {
		for _, f := range zr.File {
			if f.Name == fileName {
				return extractAndWriteFile(f)
			}
		}
		err = fmt.Errorf("file %s does not exist in archive", fileName)
	}

	return "", err
}

// Based on https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untar(data []byte, fileName string, outputDir string) (string, error) {

	gzr, err := gzip.NewReader(bytes.NewReader(data))
	defer gzr.Close()

	extractAndWriteFile := func(tr *tar.Reader, header *tar.Header) (string, error) {

		outputPath := filepath.Join(outputDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			return "", fmt.Errorf("can only untar files but %s is a directory", header.Name)
		}

		err = outputFile(outputPath, os.FileMode(header.Mode), tr)
		if err != nil {
			return "", err
		}
		return outputPath, nil
	}

	if err == nil {
		tr := tar.NewReader(gzr)

		for {
			header, err := tr.Next()
			switch {
			case err == io.EOF:
				break
			case err != nil:
				return "", err
			case header == nil:
				continue
			}
			return extractAndWriteFile(tr, header)
		}
		err = fmt.Errorf("file %s does not exist in archive", fileName)
	}

	return "", err
}

func outputFile(outputPath string, mode os.FileMode, r io.Reader) error {
	os.MkdirAll(filepath.Dir(outputPath), mode)
	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	return nil
}

func (c *DriversConfigurator) downloadDrivers(browsers *Browsers, configDir string) []downloadedDriver {
	ret := []downloadedDriver{}
	browsersToIterate := *browsers
	if c.Browsers != "" {
		requestedBrowsers := strings.Split(c.Browsers, comma)
		if len(requestedBrowsers) > 0 {
			browsersToIterate = make(Browsers)
			for _, rb := range requestedBrowsers {
				if browser, ok := (*browsers)[rb]; ok {
					browsersToIterate[rb] = browser
					continue
				}
				c.Printf("unsupported browser: %s\n", rb)
			}
		}
	}

loop:
	for browserName, browser := range browsersToIterate {
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		if architectures, ok := browser.Files[goos]; ok {
			if driver, ok := architectures[goarch]; ok {
				c.Printf("Processing %s...\n", browserName)
				driverPath, err := c.downloadDriver(&driver, configDir)
				if err != nil {
					c.Printf("Failed to download %s driver: %v\n", browserName, err)
					continue loop
				}
				command := fmt.Sprintf(browser.Command, driverPath)
				ret = append(ret, downloadedDriver{
					BrowserName: browserName,
					Command:     command,
				})
			}
		}
	}
	return ret
}
