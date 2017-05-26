package selenoid

import (
	"log"
	"os"
)

type Configurator interface {
	Configure() *SelenoidConfig
}

type Closer interface {
	Close()
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

type OutputDirAware struct {
	OutputDir string
}

func (o *OutputDirAware) createOutputDir() error {
	err := os.MkdirAll(o.OutputDir, os.ModePerm)
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
