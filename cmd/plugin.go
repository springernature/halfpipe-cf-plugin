package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type Halfpipe struct{}

type Options struct {
	ManifestPath string
}

func parseArgs(args []string) (manifestPath string, appPath string, testDomain string) {
	flagSet := flag.NewFlagSet("halfpipe", flag.ExitOnError)
	mP := flagSet.String("manifestPath", "", "Path to the manifest")
	aP := flagSet.String("appPath", "", "Path to the app")
	tD := flagSet.String("testDomain", "", "Domain to push the app to during the candidate stage")
	if err := flagSet.Parse(args[1:]); err != nil {
		panic(err)
	}

	return *mP, *aP, *tD
}

func (Halfpipe) Run(cliConnection plugin.CliConnection, args []string) {
	command := args[0]
	if command == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	logger := log.New(os.Stdout, "", 0)

	manifestPath, appPath, testDomain := parseArgs(args)

	p, err := plan.NewPlanner(manifestPath, appPath, testDomain, cliConnection).GetPlan(command)
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

func (Halfpipe) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "halfpipe",
		Commands: []plugin.Command{
			{
				Name: halfpipe_cf_plugin.PUSH,
			},
			{
				Name: halfpipe_cf_plugin.PROMOTE,
			},
		},
	}
}

func main() {
	plugin.Start(new(Halfpipe))
}
