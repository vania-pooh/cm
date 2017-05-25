package selenoid

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io"
)

type LifecycleConfig struct {
	Quiet     bool
	Force     bool
	OutputDir string
	Browsers  string
	Download  bool

	// Docker specific
	LastVersions int
	Pull         bool
	RegistryUrl  string
	Tmpfs        int

	// Drivers specific
	BrowsersJsonUrl string
	GithubBaseUrl   string
	OS              string
	Arch            string
	Version         string
}

type Lifecycle struct {
	Logger
	Forceable
	downloadable Downloadable
	configurable Configurable
	runnable     Runnable
	closer       io.Closer
}

func NewLifecycle(config *LifecycleConfig) (*Lifecycle, error) {
	lc := Lifecycle{
		Logger:    Logger{Quiet: config.Quiet},
		Forceable: Forceable{Force: config.Force},
	}
	if isDockerAvailable() {
		dockerCfg, err := NewDockerConfigurator(config)
		if err != nil {
			return nil, err
		}
		lc.downloadable = dockerCfg
		lc.configurable = dockerCfg
		lc.runnable = dockerCfg
		lc.closer = dockerCfg
	} else {
		driversCfg := NewDriversConfigurator(config)
		lc.downloadable = driversCfg
		lc.configurable = driversCfg
		lc.runnable = driversCfg
		lc.closer = driversCfg
	}
	return &lc, nil
}

func (l *Lifecycle) Close() {
	if l.closer != nil {
		l.closer.Close()
	}
}

func (l *Lifecycle) Download() error {
	if l.downloadable.IsDownloaded() && !l.Force {
		l.Printf("Selenoid is already downloaded")
		return nil
	} else {
		l.Printf("downloading Selenoid")
		return l.downloadable.Download()
	}
}

func (l *Lifecycle) Configure() error {
	return chain([]func() error{
		func() error {
			return l.Download()
		},
		func() error {
			if l.configurable.IsConfigured() && !l.Force {
				l.Printf("Selenoid is already configured")
				return nil
			}
			l.Printf("starting Selenoid configuration\n")
			return l.configurable.Configure()
		},
	})
}

func (l *Lifecycle) Start() error {
	return chain([]func() error{
		func() error {
			return l.Configure()
		},
		func() error {
			if l.runnable.IsRunning() {
				if l.Force {
					l.Printf("stopping previous Selenoid process\n")
					err := l.Stop()
					if err != nil {
						return fmt.Errorf("failed to stop previous Selenoid process: %v\n", err)
					}
				} else {
					l.Printf("Selenoid is already running\n")
				}
				return nil
			}
			l.Printf("starting Selenoid\n")
			return l.runnable.Start()
		},
	})
}

func (l *Lifecycle) Stop() error {
	return chain([]func() error{
		func() error {
			return l.Configure()
		},
		func() error {
			if !l.runnable.IsRunning() {
				l.Printf("Selenoid is not running\n")
				return nil
			}
			l.Printf("stopping Selenoid\n")
			return l.runnable.Stop()
		},
	})
	return nil
}

func isDockerAvailable() bool {
	cl, err := client.NewEnvClient()
	if err != nil {
		return false
	}
	_, err = cl.Ping(context.Background())
	return err == nil
}

func chain(steps []func() error) error {
	for _, step := range steps {
		err := step()
		if err != nil {
			return err
		}
	}
	return nil
}