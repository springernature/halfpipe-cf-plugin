package plan

import "code.cloudfoundry.org/cli/plugin/models"

type mockCliConnection struct {
	apps      []plugin_models.GetAppsModel
	appsError error
	app       plugin_models.GetAppModel
	appError  error
	cliError  error
	cliOutput []string
	space plugin_models.Space
	spaceError error
}

func (m mockCliConnection) GetApps() ([]plugin_models.GetAppsModel, error) {
	return m.apps, m.appsError
}

func (m mockCliConnection) GetApp(appName string) (plugin_models.GetAppModel, error) {
	return m.app, m.appError
}

func (m mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	return m.cliOutput, m.cliError
}

func (m mockCliConnection) GetCurrentSpace() (plugin_models.Space, error) {
	return m.space, m.spaceError
}


func (m mockCliConnection) WithApps(apps []plugin_models.GetAppsModel) mockCliConnection {
	m.apps = apps
	return m
}

func (m mockCliConnection) WithGetAppsError(error error) mockCliConnection {
	m.appsError = error
	return m
}

func (m mockCliConnection) WithApp(app plugin_models.GetAppModel) mockCliConnection {
	m.app = app
	return m
}

func (m mockCliConnection) WithGetAppError(err error) mockCliConnection {
	m.appError = err
	return m
}

func (m mockCliConnection) WithCliOutput(cliOutput []string) mockCliConnection {
	m.cliOutput = cliOutput
	return m
}


func (m mockCliConnection) WithCliError(error error) mockCliConnection {
	m.cliError = error
	return m
}

func (m mockCliConnection) WithSpace(space plugin_models.Space) mockCliConnection {
	m.space = space
	return m
}

func (m mockCliConnection) WithSpaceError(err error) mockCliConnection {
	m.spaceError = err
	return m
}


func newMockCliConnection() mockCliConnection {
	return mockCliConnection{
	}
}
