package plan

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"strings"
)

type push struct {
	manifestPath string
	appPath      string
	testDomain   string
}

func (p push) GetPlan(application manifest.Application) (plan Plan, err error) {
	candidateName := strings.Join([]string{application.Name, "CANDIDATE"}, "-")

	command := NewCfCommand(
		"push",
		candidateName,
		"-f", p.manifestPath,
		"-p", p.appPath,
		"-n", candidateName,
		"-d", p.testDomain,
	)
	plan = append(plan, command)
	return
}

func NewPush(manifestPath string, appPath string, testDomain string) push {
	return push{
		manifestPath,
		appPath,
		testDomain,
	}
}
