package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	cfPlugin "code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/plan/plugin"
)

type Halfpipe struct{}

func parseArgs(args []string) (manifestPath string, appPath string, testDomain string, space string) {
	flagSet := flag.NewFlagSet("halfpipe", flag.ExitOnError)
	mP := flagSet.String("manifestPath", "", "Path to the manifest")
	aP := flagSet.String("appPath", "", "Path to the app")
	tD := flagSet.String("testDomain", "", "Domain to push the app to during the candidate stage")
	s := flagSet.String("space", "", "Space which we are currently operating in, this is used to build candidate route")
	if err := flagSet.Parse(args[1:]); err != nil {
		panic(err)
	}

	return *mP, *aP, *tD, *s
}

func (Halfpipe) Run(cliConnection cfPlugin.CliConnection, args []string) {
	logger := log.New(os.Stdout, "", 0)
	logger.Printf("CF plugin built from git revision '%s'\n", config.SHA)

	command := args[0]
	if command == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	manifestPath, appPath, testDomain, space := parseArgs(args)
	pluginRequest := plugin.Request{
		Command:      command,
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
		Space:        space,
	}

	planner := plugin.NewPlanner(
		plugin.NewPushPlanner(plugin.NewCheck(cliConnection)),
		plugin.NewPromotePlanner(cliConnection),
		plugin.NewDeletePlanner(cliConnection),
		manifest.ReadAndMergeManifests,
	)

	p, err := planner.GetPlan(pluginRequest)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	logger.Println(p)
	if err = p.Execute(cliConnection, logger); err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}
}

func (Halfpipe) GetMetadata() cfPlugin.PluginMetadata {
	return cfPlugin.PluginMetadata{
		Name: "halfpipe",
		Commands: []cfPlugin.Command{
			{
				Name: config.PUSH,
			},
			{
				Name: config.PROMOTE,
			},
			{
				Name: config.DELETE,
			},
		},
	}
}

func main() {
	cfPlugin.Start(new(Halfpipe))
}
