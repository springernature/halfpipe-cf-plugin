package plan

type Executor interface {
	CliCommand(args ...string) ([]string, error)
}
