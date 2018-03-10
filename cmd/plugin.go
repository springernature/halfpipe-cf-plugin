package main

import (
	"code.cloudfoundry.org/cli/plugin"
)


type Halfpipe struct{}

func (Halfpipe) Run(cliConnection plugin.CliConnection, args []string) {
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
