package halfpipe_cf_plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/springernature/halfpipe-cf-plugin/config"
)

func ErrMissingArg(command string, arg string) error {
	return errors.New(fmt.Sprintf("%s requires arg %s", command, arg))
}

func ErrInvalidPreStartCommand(preStartCommand string) error {
	return errors.New(fmt.Sprintf("invalid pre-start command - only cf commands are allowed: '%s'", preStartCommand))
}

type Request struct {
	Command         string
	ManifestPath    string
	AppPath         string
	TestDomain      string
	Timeout         time.Duration
	PreStartCommand string
}

func (r Request) Verify() (err error) {
	missingArg := func(arg string) error {
		return ErrMissingArg(r.Command, arg)
	}

	switch r.Command {
	case config.PUSH:
		if r.ManifestPath == "" {
			return missingArg("manifestPath")
		}
		if r.AppPath == "" {
			return missingArg("appPath")
		}
		if r.TestDomain == "" {
			return missingArg("testDomain")
		}
		for _, preStartCommand := range strings.Split(r.PreStartCommand, ";") {
			trimmedCommand := strings.TrimSpace(preStartCommand)
			if !strings.HasPrefix(trimmedCommand, "cf ") {
				return ErrInvalidPreStartCommand(trimmedCommand)
			}
		}

	case config.PROMOTE:
		if r.ManifestPath == "" {
			return missingArg("manifestPath")
		}
		if r.TestDomain == "" {
			return missingArg("testDomain")
		}
	case config.CHECK, config.CLEANUP, config.DELETE:
		if r.ManifestPath == "" {
			return missingArg("manifestPath")
		}
	}

	return
}
