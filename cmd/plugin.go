package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/springernature/halfpipe-cf-plugin/another_package"
)


type Halfpipe struct{}

func (Halfpipe) Run(cliConnection plugin.CliConnection, args []string) {
	another_package.WillThisWork()
}

func (Halfpipe) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "halfpipe",
		Commands: []plugin.Command{
			{
				Name: "halfpipe-push",
			},
		},

	}
}

func main() {
	plugin.Start(new(Halfpipe))
}
