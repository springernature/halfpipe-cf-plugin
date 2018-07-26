package resource

import (
	"path"

	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/spf13/afero"
	"strings"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type Plan interface {
	Plan(request Request, concourseRoot string) (plan plan.Plan, err error)
}

type planner struct {
	manifestReaderWrite manifest.ManifestReaderWriter
	fs                  afero.Afero
}

func NewPlanner(manifestReaderWrite manifest.ManifestReaderWriter, fs afero.Afero) Plan {
	return planner{
		manifestReaderWrite: manifestReaderWrite,
		fs:                  fs,
	}
}

func (p planner) Plan(request Request, concourseRoot string) (pl plan.Plan, err error) {
	// Here we assume that the request is complete.
	// It has already been verified in out.go with the help of requests.VerifyRequest.

	fullManifestPath := path.Join(concourseRoot, request.Params.ManifestPath)

	if request.Params.Command == config.PUSH {
		fullGitRefPath := ""
		if request.Params.GitRefPath != "" {
			fullGitRefPath = path.Join(concourseRoot, request.Params.GitRefPath)
		}

		if err = p.updateManifestWithVars(fullManifestPath, fullGitRefPath, request.Params.Vars); err != nil {
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
	}

	switch request.Params.Command {
	case config.PUSH:
		pl = append(pl, plan.NewCfCommand(request.Params.Command,
			"-manifestPath", fullManifestPath,
			"-appPath", path.Join(concourseRoot, request.Params.AppPath),
			"-testDomain", request.Params.TestDomain,
			"-space", request.Source.Space,
		))
	case config.PROMOTE:
		pl = append(pl, plan.NewCfCommand(request.Params.Command,
			"-manifestPath", fullManifestPath,
			"-testDomain", request.Params.TestDomain,
			"-space", request.Source.Space,
		))
	case config.CLEANUP, config.DELETE:
		pl = append(pl, plan.NewCfCommand(request.Params.Command,
			"-manifestPath", fullManifestPath,
		))
	}

	return
}

func (p planner) updateManifestWithVars(manifestPath string, gitRefPath string, vars map[string]string) (err error) {
	if len(vars) > 0 || gitRefPath != "" {
		apps, e := p.manifestReaderWrite.ReadManifest(manifestPath)
		if e != nil {
			err = e
			return
		}

		// We just assume the first app in the manifest is the app under deployment.
		// We lint that this is the case in the halfpipe linter.
		app := apps.Applications[0]
		if len(app.EnvironmentVariables) == 0 {
			app.EnvironmentVariables = map[string]string{}
		}

		for key, value := range vars {
			app.EnvironmentVariables[key] = value
		}

		if gitRefPath != "" {
			ref, errRead := p.readGitRef(gitRefPath)
			if errRead != nil {
				err = errRead
				return
			}
			app.EnvironmentVariables["GIT_REVISION"] = ref
		}

		if err = p.manifestReaderWrite.WriteManifest(manifestPath, app); err != nil {
			return
		}
	}
	return
}

func (p planner) readGitRef(gitRefPath string) (ref string, err error) {
	bytes, err := p.fs.ReadFile(gitRefPath)
	if err != nil {
		return
	}
	ref = strings.TrimSpace(string(bytes))
	return
}
