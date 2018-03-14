package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/springernature/halfpipe-cf-plugin/controller"
	"github.com/springernature/halfpipe-cf-plugin/color"
	"github.com/springernature/halfpipe-cf-plugin"
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
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	logger := log.New(os.Stdout, "", 0)

	manifestPath, appPath, testDomain := parseArgs(args)

	p, err := controller.NewController(args[0], manifestPath, appPath, testDomain, cliConnection).GetPlan()
	if err != nil {
		logger.Println(color.ErrColor.Sprint(err))
		syscall.Exit(1)
	}

	logger.Println(color.PlanColor.Sprint(p))
	if err = p.Execute(cliConnection, logger, color.PlanColor); err != nil {
		logger.Println(color.ErrColor.Sprint(err))
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
