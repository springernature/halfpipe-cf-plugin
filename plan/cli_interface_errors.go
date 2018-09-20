package plan

import "fmt"

func ErrGetApps(err error) error {
	return fmt.Errorf("'cf apps' erroed with: %s", err)
}

func ErrGetApp(app string, err error) error {
	return fmt.Errorf("'cf app %s' erroed with: %s", app, err)
}

func ErrGetCurrentSpace(err error) error {
	return fmt.Errorf("'cf target' erroed with: %s", err)
}

func ErrCliCommandWithoutTerminalOutput(command string, err error) error {
	return fmt.Errorf("'%s' erroed with: %s", command, err)
}