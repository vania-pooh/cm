package selenoid

import (
	"log"
	"os"
)

type Configurator interface {
	Configure() *SelenoidConfig
}

type Downloadable interface {
	IsDownloaded() bool
	Download() (string, error)
}

type Configurable interface {
	IsConfigured() bool
	Configure() (*SelenoidConfig, error)
}

type Runnable interface {
	IsRunning() bool
	Start() error
	Stop() error
}

type Logger struct {
	Quiet bool
}

func (c *Logger) Printf(format string, v ...interface{}) {
	if !c.Quiet {
		log.Printf(format, v...)
	}
}

type ConfigDirAware struct {
	ConfigDir string
}

func (c *ConfigDirAware) createConfigDir() error {
	err := os.MkdirAll(c.ConfigDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

type Forceable struct {
	Force bool
}

type VersionAware struct {
	Version string
}

type DownloadAware struct {
	DownloadNeeded bool
}

type RequestedBrowsersAware struct {
	Browsers string
}
