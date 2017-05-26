package selenoid

import (
	"fmt"
	. "github.com/aandryashin/matchers"
	"github.com/aerokube/selenoid/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var (
	mockDockerServer *httptest.Server
)

func init() {
	mockDockerServer = httptest.NewServer(mux())
	os.Setenv("DOCKER_HOST", "tcp://"+hostPort(mockDockerServer.URL))
	os.Setenv("DOCKER_API_VERSION", "1.29")
}

func mux() http.Handler {
	mux := http.NewServeMux()

	//Docker Registry API mock
	mux.HandleFunc("/v2/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	mux.HandleFunc("/v2/selenoid/firefox/tags/list", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintln(w, `{"name":"firefox", "tags": ["46.0", "45.0", "7.0", "latest"]}`)
		},
	))

	mux.HandleFunc("/v2/selenoid/opera/tags/list", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintln(w, `{"name":"opera", "tags": ["44.0", "latest"]}`)
		},
	))

	//Docker API mock
	mux.HandleFunc("/v1.29/images/create", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			output := `{"id": "a86cd3433934", "status": "Downloading layer"}`
			w.Write([]byte(output))
		},
	))
	return mux
}

func hostPort(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	return u.Host
}

func TestImageWithTag(t *testing.T) {
	AssertThat(t, imageWithTag("selenoid/firefox", "tag"), EqualTo{"selenoid/firefox:tag"})
}

func TestFetchImageTags(t *testing.T) {
	lcConfig := LifecycleConfig{
		RegistryUrl: mockDockerServer.URL,
		Quiet:       false,
	}
	c, err := NewDockerConfigurator(&lcConfig)
	AssertThat(t, err, Is{nil})
	defer c.Close()
	tags := c.fetchImageTags("selenoid/firefox")
	AssertThat(t, len(tags), EqualTo{3})
	AssertThat(t, tags[0], EqualTo{"46.0"})
	AssertThat(t, tags[1], EqualTo{"45.0"})
	AssertThat(t, tags[2], EqualTo{"7.0"})
}

func TestPullImages(t *testing.T) {
	lcConfig := LifecycleConfig{
		RegistryUrl: mockDockerServer.URL,
		Quiet:       false,
	}
	c, err := NewDockerConfigurator(&lcConfig)
	AssertThat(t, err, Is{nil})
	defer c.Close()
	tags := c.pullImages("selenoid/firefox", []string{"46.0", "45.0"})
	AssertThat(t, len(tags), EqualTo{2})
	AssertThat(t, tags[0], EqualTo{"46.0"})
	AssertThat(t, tags[1], EqualTo{"45.0"})
}

func TestConfigureDocker(t *testing.T) {
	testConfigure(t, true)
}

func TestLimitNoPull(t *testing.T) {
	testConfigure(t, false)
}

func testConfigure(t *testing.T, pull bool) {
	lcConfig := LifecycleConfig{
		RegistryUrl: mockDockerServer.URL,
		Quiet:       false,
	}
	c, err := NewDockerConfigurator(&lcConfig)
	AssertThat(t, err, Is{nil})
	c.LastVersions = 2
	c.Pull = pull
	c.Tmpfs = 512
	defer c.Close()
	cfgPointer, err := (*c).Configure()
	AssertThat(t, err, Is{nil})
	AssertThat(t, cfgPointer, Is{Not{nil}})

	cfg := *cfgPointer
	AssertThat(t, len(cfg), EqualTo{2})

	firefoxVersions, hasFirefoxKey := cfg["firefox"]
	AssertThat(t, hasFirefoxKey, Is{true})
	AssertThat(t, firefoxVersions, Is{Not{nil}})

	tmpfsMap := make(map[string]string)
	tmpfsMap["/tmp"] = "size=512m"

	correctFFBrowsers := make(map[string]*config.Browser)
	correctFFBrowsers["46.0"] = &config.Browser{
		Image: "selenoid/firefox:46.0",
		Port:  "4444",
		Path:  "/wd/hub",
		Tmpfs: tmpfsMap,
	}
	correctFFBrowsers["45.0"] = &config.Browser{
		Image: "selenoid/firefox:45.0",
		Port:  "4444",
		Path:  "/wd/hub",
		Tmpfs: tmpfsMap,
	}
	AssertThat(t, firefoxVersions, EqualTo{config.Versions{
		Default:  "46.0",
		Versions: correctFFBrowsers,
	}})

	operaVersions, hasPhantomjsKey := cfg["opera"]
	AssertThat(t, hasPhantomjsKey, Is{true})
	AssertThat(t, operaVersions, Is{Not{nil}})
	AssertThat(t, operaVersions.Default, EqualTo{"44.0"})

	correctPhantomjsBrowsers := make(map[string]*config.Browser)
	correctPhantomjsBrowsers["2.1.1"] = &config.Browser{
		Image: "selenoid/opera:44.0",
		Port:  "4444",
		Tmpfs: tmpfsMap,
	}
}
