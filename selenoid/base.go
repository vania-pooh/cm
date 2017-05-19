package selenoid

import "log"

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
