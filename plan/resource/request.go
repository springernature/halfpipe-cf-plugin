package resource

import (
	"errors"
	"fmt"

	"github.com/springernature/halfpipe-cf-plugin/config"
)

type Request struct {
	Source Source
	Params Params
}

type Source struct {
	API                  string
	Org                  string
	Space                string
	Username             string
	Password             string
	PrometheusGatewayURL string
}

type Params struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
	Vars         map[string]string
	GitRefPath   string
}

func SourceMissingError(field string) error {
	return errors.New(fmt.Sprintf("Source config must contain %s", field))
}

func ParamsMissingError(field string) error {
	return errors.New(fmt.Sprintf("Params config must contain %s", field))
}

func VerifyRequest(request Request) error {
	if err := VerifyRequestSource(request.Source); err != nil {
		return err
	}

	if err := VerifyRequestParams(request.Params); err != nil {
		return err
	}

	return nil
}

func VerifyRequestSource(source Source) error {
	if source.API == "" {
		return SourceMissingError("api")
	}

	if source.Space == "" {
		return SourceMissingError("space")
	}

	if source.Org == "" {
		return SourceMissingError("org")
	}

	if source.Password == "" {
		return SourceMissingError("password")
	}

	if source.Username == "" {
		return SourceMissingError("username")
	}

	return nil
}

func VerifyRequestParams(params Params) error {
	if params.Command == "" {
		return ParamsMissingError("command")
	}

	if params.ManifestPath == "" {
		return ParamsMissingError("manifestPath")
	}

	switch params.Command {
	case config.PUSH:
		if params.TestDomain == "" {
			return ParamsMissingError("testDomain")
		}

		if params.AppPath == "" {
			return ParamsMissingError("appPath")
		}

		if params.GitRefPath == "" {
			return ParamsMissingError("gitRefPath")
		}
	case config.PROMOTE:
		if params.TestDomain == "" {
			return ParamsMissingError("testDomain")
		}
	}

	return nil
}
