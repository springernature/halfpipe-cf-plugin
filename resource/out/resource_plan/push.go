package resource_plan

import (
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"fmt"
	"errors"
)

var ErrEmptyParamValue = func(fieldName string) (error) {
	errorMsg := fmt.Sprintf("Field params.%s must not be empty!", fieldName)
	return errors.New(errorMsg)
}

type push struct{}

func NewPush() push {
	return push{}
}

func checkField(field string, value string) (err error) {
	if value == "" {
		err = ErrEmptyParamValue(field)
	}
	return
}

func check(request out.Request) (err error) {
	if err = checkField("manifestPath", request.Params.ManifestPath); err != nil {
		return
	}

	if err = checkField("space", request.Params.Space); err != nil {
		return
	}

	if err = checkField("org", request.Params.Org); err != nil {
		return
	}

	if err = checkField("password", request.Params.Password); err != nil {
		return
	}

	if err = checkField("username", request.Params.Username); err != nil {
		return
	}

	if err = checkField("api", request.Params.Api); err != nil {
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
			"-a", request.Params.Api,
			"-u", request.Params.Username,
			"-p", request.Params.Password,
			"-o", request.Params.Org,
			"-s", request.Params.Space),
		plan.NewCfCommand("halfpipe-push",
			"-manifestPath", request.Params.ManifestPath,
			"-appPath", request.Params.AppPath),
	}

	return
}
