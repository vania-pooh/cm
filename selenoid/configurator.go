package selenoid

import "log"

type Configurator interface {
	Configure() *SelenoidConfig
}

type BaseConfigurator struct {
	Quiet bool
}

func (c *BaseConfigurator) Printf(format string, v ...interface{}) {
	if !c.Quiet {
		log.Printf(format, v...)
	}
}
