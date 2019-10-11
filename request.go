package halfpipe_cf_plugin

import (
	"errors"
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"strings"
	"time"
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
		if len(r.PreStartCommand) > 0 && !strings.HasPrefix(r.PreStartCommand, "cf ") {
			return ErrInvalidPreStartCommand(r.PreStartCommand)
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
