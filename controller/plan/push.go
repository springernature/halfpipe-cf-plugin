package plan

type push struct {
	manifestPath string
	appPath      string
}

func (p push) Commands() (plan Plan, err error) {
	command := NewCfCommand("push", "-f", p.manifestPath)

	if p.appPath != "" {
		command.args = append(command.args, "-p", p.appPath)
	}

	plan = append(plan, command)
	return
}

func NewPush(manifestPath string, appPath string) push {
	return push{
		manifestPath,
		appPath,
	}
}