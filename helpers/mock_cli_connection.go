package helpers

import "code.cloudfoundry.org/cli/plugin/models"

type mockCliConnection struct {
	apps       []plugin_models.GetAppsModel
	appsError  error
	appFun     func() (plugin_models.GetAppModel, error)
	cliError   error
	cliOutput  []string
	space      plugin_models.Space
	spaceError error
}

func (m mockCliConnection) GetApps() ([]plugin_models.GetAppsModel, error) {
	return m.apps, m.appsError
}

func (m mockCliConnection) GetApp(appName string) (plugin_models.GetAppModel, error) {
	return m.appFun()
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
	m.appFun = func() (model plugin_models.GetAppModel, e error) {
		return app, nil
	}
	return m
}

func (m mockCliConnection) WithAppFun(appFun func() (model plugin_models.GetAppModel, e error)) mockCliConnection {
	m.appFun = appFun
	return m
}

func (m mockCliConnection) WithGetAppError(err error) mockCliConnection {
	m.appFun = func() (model plugin_models.GetAppModel, e error) {
		return plugin_models.GetAppModel{}, err
	}
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

func (m mockCliConnection) WithSpace(space string) mockCliConnection {
	m.space = plugin_models.Space{
		SpaceFields: plugin_models.SpaceFields{
			Guid: "wakawaka",
			Name: space,
		},
	}
	return m
}

func (m mockCliConnection) WithSpaceError(err error) mockCliConnection {
	m.spaceError = err
	return m
}

func NewMockCliConnection() mockCliConnection {
	return mockCliConnection{
	}
}
