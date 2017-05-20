package selenoid

import (
	"encoding/json"
	"fmt"
	. "github.com/aandryashin/matchers"
	"github.com/google/go-github/github"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

//TODO: add docs for drivers, download and sync commands

const (
	previousReleaseTag = "1.2.0"
	latestReleaseTag   = "1.2.1"
	version            = "version"
)

var (
	mockDownloaderServer *httptest.Server
	releaseFileName      = fmt.Sprintf("selenoid_%s_%s", runtime.GOOS, runtime.GOARCH)
)

func init() {
	mockDownloaderServer = httptest.NewServer(downloaderMux())
}

func downloaderMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, previousReleaseTag),
		http.HandlerFunc(getReleaseHandler(previousReleaseTag)),
	)
	mux.HandleFunc(
		fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo),
		http.HandlerFunc(getReleaseHandler(latestReleaseTag)),
	)
	mux.HandleFunc("/"+releaseFileName, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			version := r.URL.Query().Get(version)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(version))
		},
	))

	return mux
}

func getReleaseHandler(v string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		releaseUrl := mockServerUrl(
			mockDownloaderServer,
			fmt.Sprintf("/%s?%s=%s", releaseFileName, version, v),
		)
		release := github.RepositoryRelease{
			Assets: []github.ReleaseAsset{
				{
					Name:               &releaseFileName,
					BrowserDownloadURL: &releaseUrl,
				},
			},
		}
		data, _ := json.Marshal(&release)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func TestDownloadLatestRelease(t *testing.T) {
	testDownloadRelease(t, Latest, latestReleaseTag)
}

func TestDownloadSpecificRelease(t *testing.T) {
	testDownloadRelease(t, previousReleaseTag, previousReleaseTag)
}

func testDownloadRelease(t *testing.T, desiredVersion string, expectedFileContents string) {
	withTmpDir(t, "downloader", func(t *testing.T, dir string) {
		downloader := NewDownloader(
			mockDownloaderServer.URL,
			dir,
			runtime.GOOS,
			runtime.GOARCH,
			desiredVersion,
			false,
		)
		err := downloader.Download()
		AssertThat(t, err, Is{nil})

		releaseOutputPath := filepath.Join(dir, "selenoid")
		_, err = os.Stat(releaseOutputPath)
		if os.IsNotExist(err) {
			t.Fatalf("release was not downloaded to %s: file does not exist\n", releaseOutputPath)
		}

		data, err := ioutil.ReadFile(releaseOutputPath)
		AssertThat(t, err, Is{nil})
		AssertThat(t, string(data), EqualTo{expectedFileContents})
	})

}

func TestUnknownRelease(t *testing.T) {
	downloadShouldFail(t, func(dir string) *Downloader {
		return NewDownloader(
			mockDownloaderServer.URL,
			dir,
			runtime.GOOS,
			runtime.GOARCH,
			"missing-version",
			false,
		)
	})
}

func downloadShouldFail(t *testing.T, fn func(string) *Downloader) {
	withTmpDir(t, "something", func(t *testing.T, dir string) {
		downloader := fn(dir)
		err := downloader.Download()
		fmt.Printf("err = %v\n", err)
		AssertThat(t, err, Is{Not{nil}})
	})
}

func TestUnavailableBinary(t *testing.T) {
	downloadShouldFail(t, func(dir string) *Downloader {
		return NewDownloader(
			mockDownloaderServer.URL,
			dir,
			"missing-os",
			"missing-arch",
			previousReleaseTag,
			false,
		)
	})
}

func TestWrongBaseUrl(t *testing.T) {
	downloadShouldFail(t, func(dir string) *Downloader {
		return NewDownloader(
			":::bad-url:::",
			dir,
			runtime.GOOS,
			runtime.GOARCH,
			Latest,
			false,
		)
	})
}
