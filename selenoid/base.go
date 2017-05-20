package selenoid

import (
	"log"
	"os"
)

type Configurator interface {
	Configure() *SelenoidConfig
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
