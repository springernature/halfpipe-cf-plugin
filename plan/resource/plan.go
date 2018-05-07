package resource

import (
	"errors"
	"fmt"
	"path"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"strings"
)

var NewErrEmptyParamValue = func(fieldName string) error {
	errorMsg := fmt.Sprintf("Field params.%s must not be empty!", fieldName)
	return errors.New(errorMsg)
}

type ErrEmptySourceValue error

var NewErrEmptySourceValue = func(fieldName string) error {
	errorMsg := fmt.Sprintf("Field source.%s must not be empty!", fieldName)
	return errors.New(errorMsg)
}

type Plan interface {
	Plan(request Request, concourseRoot string, gitRef string) (plan plan.Plan, err error)
}

type planner struct {
	manifestReader func(pathToManifest string) ([]manifest.Application, error)
	manifestWriter func(application manifest.Application, filePath string) error
}

func NewPlanner(manifestReader func(pathToManifest string) ([]manifest.Application, error), manifestWriter func(application manifest.Application, filePath string) error) Plan {
	return planner{
		manifestReader: manifestReader,
		manifestWriter: manifestWriter,
	}
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

func check(request Request) (err error) {
	if err = checkParamField("manifestPath", request.Params.ManifestPath); err != nil {
		return
	}

	if err = checkParamField("testDomain", request.Params.TestDomain); err != nil {
		return
	}

	if err = checkParamField("command", request.Params.Command); err != nil {
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

	if err = checkSourceField("api", request.Source.API); err != nil {
		return
	}

	return
}

func (p planner) Plan(request Request, concourseRoot string, gitRef string) (pl plan.Plan, err error) {
	if err = check(request); err != nil {
		return
	}

	fullManifestPath := path.Join(concourseRoot, request.Params.ManifestPath)

	if request.Params.Command == config.PUSH {
		if err = p.updateManifestWithVars(fullManifestPath, request.Params.Vars, gitRef); err != nil {
			return
		}
	}

	pl = plan.Plan{
		plan.NewCfCommand("login",
			"-a", request.Source.API,
			"-u", request.Source.Username,
			"-p", request.Source.Password,
			"-o", request.Source.Org,
			"-s", request.Source.Space),
		plan.NewCfCommand(request.Params.Command,
			"-manifestPath", fullManifestPath,
			"-appPath", path.Join(concourseRoot, request.Params.AppPath),
			"-testDomain", request.Params.TestDomain,
			"-space", request.Source.Space,
			"-env", strings.Join([]string{"GIT_REVISION", ""}, "="),
		),
	}

	return
}

func (p planner) updateManifestWithVars(manifestPath string, vars map[string]string, gitRef string) (err error) {
	apps, e := p.manifestReader(manifestPath)
	if e != nil {
		err = e
		return
	}

	// We just assume the first app in the manifest is the app under deployment.
	// We lint that this is the case in the halfpipe linter.
	app := apps[0]
	if len(app.EnvironmentVariables) == 0 {
		app.EnvironmentVariables = map[string]string{}
	}

	for key, value := range vars {
		app.EnvironmentVariables[key] = value
	}

	if gitRef != "" {
		app.EnvironmentVariables["GIT_REVISION"] = gitRef
	}

	if err = p.manifestWriter(app, manifestPath); err != nil {
		return
	}
	return
}
