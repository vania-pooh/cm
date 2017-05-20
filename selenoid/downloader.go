package selenoid

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/google/go-github/github"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	owner = "aerokube"
	repo  = "selenoid"
)

type Downloader struct {
	Logger
	GithubBaseUrl string
	OutputDir     string
	OS            string
	Arch          string
	Version       string
}

func NewDownloader(githubBaseUrl string, outputDir string, os string, arch string, version string, quiet bool) *Downloader {
	return &Downloader{
		Logger:        Logger{Quiet: quiet},
		GithubBaseUrl: githubBaseUrl,
		OutputDir:     outputDir,
		OS:            os,
		Arch:          arch,
		Version:       version,
	}
}

func (d *Downloader) Download() error {
	u, err := d.getUrl()
	if err != nil {
		return fmt.Errorf("failed to get download URL for arch = %s and version = %s: %v\n", d.Arch, d.Version, err)
	}
	outputFile, err := d.downloadFile(u)
	if err != nil {
		return fmt.Errorf("failed to download Selenoid for arch = %s and version = %s: %v\n", d.Arch, d.Version, err)
	}
	d.Printf("successfully downloaded Selenoid to %s\n", outputFile)
	return nil
}

func (d *Downloader) getUrl() (string, error) {
	d.Printf("getting Selenoid release information for version: %s\n", d.Version)
	ctx := context.Background()
	client := github.NewClient(nil)
	if d.GithubBaseUrl != "" {
		u, err := url.Parse(d.GithubBaseUrl)
		if err != nil {
			return "", fmt.Errorf("invalid Github base url [%s]: %v\n", d.GithubBaseUrl, err)
		}
		client.BaseURL = u
	}
	var release *github.RepositoryRelease
	var err error
	if d.Version != Latest {
		release, _, err = client.Repositories.GetReleaseByTag(ctx, owner, repo, d.Version)
	} else {
		release, _, err = client.Repositories.GetLatestRelease(ctx, owner, repo)
	}

	if err != nil {
		return "", err
	}

	if release == nil {
		return "", fmt.Errorf("unknown release: %s\n", d.Version)
	}

	for _, asset := range release.Assets {
		assetName := *(asset.Name)
		if strings.Contains(assetName, d.OS) && strings.Contains(assetName, d.Arch) {
			return *(asset.URL), nil
		}
	}
	return "", fmt.Errorf("Selenoid binary for %s %s is not available for specified release: %s\n", strings.Title(d.OS), d.Arch, d.Version)
}

func (d *Downloader) downloadFile(url string) (string, error) {
	d.Printf("downloading Selenoid release from %s\n", url)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	outputPath := filepath.Join(d.OutputDir, "selenoid")
	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return "", err
	}
	d.Printf("Selenoid binary saved to %s. Don't forget to add %s to PATH, e.g.:\n $ export PATH=\"$PATH:%s\".\n", outputPath, d.OutputDir, d.OutputDir)
	return outputPath, nil
}
