package selenoid

import (
	"encoding/json"
	"fmt"
	. "github.com/aandryashin/matchers"
	"github.com/aerokube/selenoid/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
)

var (
	mockDriverServer *httptest.Server
)

func init() {
	mockDriverServer = httptest.NewServer(driversMux())
}

func driversMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/browsers.json", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			goos := runtime.GOOS
			goarch := runtime.GOARCH
			browsers := Browsers{
				"first": Browser{
					Command: "%s",
					Files: Files{
						goos: {
							goarch: Driver{
								URL:      mockServerUrl("/testfile.zip"),
								Filename: "zip-testfile",
							},
						},
					},
				},
				"second": Browser{
					Command: "%s",
					Files: Files{
						goos: {
							goarch: Driver{
								URL:      mockServerUrl("/testfile.tar.gz"),
								Filename: "gzip-testfile",
							},
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&browsers)
		},
	))

	//Serving static files from current directory
	mux.Handle("/", http.FileServer(http.Dir("")))

	return mux
}

func TestAllUrlsAreValid(t *testing.T) {

	dir, err := os.Getwd()
	AssertThat(t, err, Is{nil})

	data := readFile(t, path.Join(dir, "..", "browsers.json"))

	var browsers Browsers
	err = json.Unmarshal(data, &browsers)
	AssertThat(t, err, Is{nil})

	//Loops are ugly but we need to check all urls in one test...
	for _, browser := range browsers {
		for _, architectures := range browser.Files {
			for _, driver := range architectures {
				u := driver.URL
				fmt.Printf("Checking URL: %s\n", u)
				req, err := http.NewRequest(http.MethodHead, u, nil)
				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						/*
							Do not follow redirects in order to avoid 403 Forbidden responses from S3 when checking Github releases links
						*/
						return http.ErrUseLastResponse
					},
				}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("failed to request url %s: %v\n", u, err)
				}
				if resp.StatusCode != 200 && resp.StatusCode != 301 && resp.StatusCode != 302 {
					t.Fatalf("broken url %s: %d", u, resp.StatusCode)
				}
			}
		}
	}
}

func TestConfigureDrivers(t *testing.T) {

	dir, err := ioutil.TempDir("", "test-download")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	browsersJsonUrl := mockServerUrl("/browsers.json")
	configurator := NewDriversConfigurator(dir, "", browsersJsonUrl, true, false)
	cfg := *configurator.Configure()
	AssertThat(t, len(cfg), EqualTo{2})

	unpackedFirstFile := path.Join(dir, "zip-testfile")
	unpackedSecondFile := path.Join(dir, "gzip-testfile")
	correctConfig := SelenoidConfig{
		"first": config.Versions{
			Default: latest,
			Versions: map[string]*config.Browser{
				latest: {
					Image: []string{unpackedFirstFile},
				},
			},
		},
		"second": config.Versions{
			Default: latest,
			Versions: map[string]*config.Browser{
				latest: {
					Image: []string{unpackedSecondFile},
				},
			},
		},
	}

	if !reflect.DeepEqual(cfg, correctConfig) {
		cfgData, _ := json.MarshalIndent(cfg, "", "    ")
		correctConfigData, _ := json.MarshalIndent(correctConfig, "", "    ")
		t.Fatalf("Incorrect config. Expected:\n %+v\n Actual: %+v\n", string(correctConfigData), string(cfgData))
	}

	for _, unpackedFile := range []string{unpackedFirstFile, unpackedSecondFile} {
		_, err = os.Stat(unpackedFile)
		if os.IsNotExist(err) {
			t.Fatalf("file %s does not exist\n", unpackedFile)
		}
	}

}

func TestUnzip(t *testing.T) {
	data := readFile(t, "testfile.zip")
	AssertThat(t, isZipFile(data), Is{true})
	AssertThat(t, isTarGzFile(data), Is{false})
	testUnpack(t, data, "zip-testfile", func(data []byte, filePath string, outputDir string) (string, error) {
		return unzip(data, filePath, outputDir)
	}, "zip\n")
}

func TestUntar(t *testing.T) {
	data := readFile(t, "testfile.tar.gz")
	AssertThat(t, isTarGzFile(data), Is{true})
	AssertThat(t, isZipFile(data), Is{false})
	testUnpack(t, data, "gzip-testfile", func(data []byte, filePath string, outputDir string) (string, error) {
		return untar(data, filePath, outputDir)
	}, "gzip\n")
}

func testUnpack(t *testing.T, data []byte, fileName string, fn func([]byte, string, string) (string, error), correctContents string) {

	dir, err := ioutil.TempDir("", "test-unpack")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	unpackedFile, err := fn(data, fileName, dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(unpackedFile)
	if os.IsNotExist(err) {
		t.Fatalf("file %s does not exist\n", unpackedFile)
	}

	unpackedFileContents := string(readFile(t, unpackedFile))
	if unpackedFileContents != correctContents {
		t.Fatalf("incorrect unpacked file contents; expected: '%s', actual: '%s'\n", correctContents, unpackedFileContents)
	}

}

func readFile(t *testing.T, fileName string) []byte {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestDownloadFile(t *testing.T) {
	fileUrl := mockServerUrl("/testfile")
	data, err := downloadFile(fileUrl)
	if err != nil {
		t.Fatalf("failed to download file: %v\n", err)
	}
	AssertThat(t, string(data), EqualTo{"test-data"})
}

func mockServerUrl(relativeUrl string) string {
	base, _ := url.Parse(mockDriverServer.URL)
	relative, _ := url.Parse(relativeUrl)
	return base.ResolveReference(relative).String()
}
