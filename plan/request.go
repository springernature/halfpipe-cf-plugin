package plan

import (
	"time"
	"fmt"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/config"
)

func ErrMissingArg(command string, arg string) error {
	return errors.New(fmt.Sprintf("%s requires arg %s", command, arg))
}

type Request struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
	Space        string
	Timeout      time.Duration
}

func (r Request) Verify() (err error) {
	returnErr := func(arg string) error {
		return ErrMissingArg(r.Command, arg)
	}

	switch r.Command {
	case config.PUSH:
		if r.ManifestPath == "" {
			return returnErr("manifestPath")
		}
		if r.AppPath == "" {
			return returnErr("appPath")
		}
		if r.TestDomain == "" {
			return returnErr("testDomain")
		}
		if r.Space == "" {
			return returnErr("space")
		}
	case config.PROMOTE:
		if r.ManifestPath == "" {
			return returnErr("manifestPath")
		}
		if r.TestDomain == "" {
			return returnErr("testDomain")
		}
		if r.Space == "" {
			return returnErr("space")
		}
	case config.CLEANUP, config.DELETE:
		if r.ManifestPath == "" {
			return returnErr("manifestPath")
		}
	}

	return
}
