package resource_plan

import (
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"fmt"
	"errors"
)

var NewErrEmptyParamValue = func(fieldName string) (error) {
	errorMsg := fmt.Sprintf("Field params.%s must not be empty!", fieldName)
	return errors.New(errorMsg)
}

type ErrEmptySourceValue error
var NewErrEmptySourceValue = func(fieldName string) (error) {
	errorMsg := fmt.Sprintf("Field source.%s must not be empty!", fieldName)
	return errors.New(errorMsg)
}


type push struct{}

func NewPush() push {
	return push{}
}


func checkParamField(field string, value string) (err error) {
	if value == "" {
		err = NewErrEmptyParamValue(field)
	}
	return
}

func checkSourceField(field string, value string) (err error) {
	if value == "" {
		err = NewErrEmptySourceValue(field)
	}
	return
}

func check(request out.Request) (err error) {
	if err = checkParamField("manifestPath", request.Params.ManifestPath); err != nil {
		return
	}

	if err = checkSourceField("space", request.Source.Space); err != nil {
		return
	}

	if err = checkSourceField("org", request.Source.Org); err != nil {
		return
	}

	if err = checkSourceField("password", request.Source.Password); err != nil {
		return
	}

	if err = checkSourceField("username", request.Source.Username); err != nil {
		return
	}

	if err = checkSourceField("api", request.Source.Api); err != nil {
		return
	}

	return
}

func (push) Plan(request out.Request) (p plan.Plan, err error) {
	if err = check(request); err != nil {
		return
	}

	p = plan.Plan{
		plan.NewCfCommand("login",
			"-a", request.Source.Api,
			"-u", request.Source.Username,
			"-p", request.Source.Password,
			"-o", request.Source.Org,
			"-s", request.Source.Space),
		plan.NewCfCommand("halfpipe-push",
			"-manifestPath", request.Params.ManifestPath,
			"-appPath", request.Params.AppPath),
	}

	return
}
