package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/springernature/halfpipe-cf-plugin/controller"
	"github.com/springernature/halfpipe-cf-plugin/color"
)

type Halfpipe struct{}

type Options struct {
	ManifestPath string
}

func parseArgs(args []string) (manifestPath string, appPath string) {
	flagSet := flag.NewFlagSet("halfpipe", flag.ExitOnError)
	mP := flagSet.String("manifestPath", "", "Path to the manifest")
	aP := flagSet.String("appPath", "", "Path to the app")
	if err := flagSet.Parse(args[1:]); err != nil {
		panic(err)
	}

	return *mP, *aP
}

func (Halfpipe) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	logger := log.New(os.Stdout, "", 0)

	manifestPath, appPath := parseArgs(args)

	p, err := controller.NewController(args[0], manifestPath, appPath).GetPlan()
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
				Name: "halfpipe-push",
			},
		},
	}
}

func main() {
	plugin.Start(new(Halfpipe))
}
