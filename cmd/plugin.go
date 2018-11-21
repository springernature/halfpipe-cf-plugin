package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	cfPlugin "code.cloudfoundry.org/cli/plugin"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"time"
)

type Halfpipe struct{}

func parseArgs(args []string) (manifestPath string, appPath string, testDomain string, timeout time.Duration) {
	flagSet := flag.NewFlagSet("halfpipe", flag.ExitOnError)
	mP := flagSet.String("manifestPath", "", "Path to the manifest")
	aP := flagSet.String("appPath", "", "Path to the app")
	tD := flagSet.String("testDomain", "", "Domain to push the app to during the candidate stage")
	tO := flagSet.Duration("timeout", 5*time.Minute, "Timeout for each command")
	if err := flagSet.Parse(args[1:]); err != nil {
		panic(err)
	}

	return *mP, *aP, *tD, *tO
}

func (Halfpipe) Run(cliConnection cfPlugin.CliConnection, args []string) {
	logger := log.New(os.Stdout, "", 0)
	logger.Printf("# CF plugin built from git revision '%s'\n", config.SHA)

	command := args[0]
	if command == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	manifestPath, appPath, testDomain, timeout := parseArgs(args)

	pluginRequest := plan.Request{
		Command:      command,
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
		Timeout:      timeout,
	}

	if err := pluginRequest.Verify(); err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	planner := plan.NewPlanner(
		plan.NewPushPlanner(cliConnection),
		plan.NewPromotePlanner(cliConnection),
		plan.NewCleanupPlanner(cliConnection),
		manifest.NewManifestReadWrite(afero.Afero{Fs: afero.NewOsFs()}),
	)

	p, err := planner.GetPlan(pluginRequest)
	if err != nil {
		logger.Println("Failed to create execution plan!")
		logger.Println(err)
		syscall.Exit(1)
	}

	logger.Println(p)
	if err = p.Execute(plan.NewShellExecutor(logger), pluginRequest.Timeout, logger); err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}
}

func (Halfpipe) GetMetadata() cfPlugin.PluginMetadata {
	return cfPlugin.PluginMetadata{
		Name: "halfpipe",
		Commands: []cfPlugin.Command{
			{
				Name:     config.PUSH,
				HelpText: "Pushes the app with name `my-app` as a `my-app-candidate` and binds a single temporary route to it for testing",
				UsageDetails: cfPlugin.Usage{
					Usage: "cf halfpipe-push [-manifestPath PATH] [-appPath PATH] [-testDomain DOMAIN] [-space DOMAIN]",
					Options: map[string]string{
						"-manifestPath": "Relative or absolute path to cf manifest",
						"-appPath":      "Relative or absolute path to the app bits you wish to deploy",
						"-testDomain":   "Domain that will be used when constructing the candidate route for the app",
						"-space":        "Space will be used when constructing the candidate test route",
						"-timeout":      "Timeout for all the sub commands, example 10s, 2m37s",
					},
				},
			},
			{
				Name:     config.PROMOTE,
				HelpText: "Promotes the app from `my-app-candidate` to `my-app`, binds all the production routes, removes the test route and stops old instances of the app",
				UsageDetails: cfPlugin.Usage{
					Usage: "cf halfpipe-push [-manifestPath PATH] [-testDomain DOMAIN] [-space DOMAIN]",
					Options: map[string]string{
						"-manifestPath": "Relative or absolute path to cf manifest",
						"-testDomain":   "Domain that will be used when constructing the candidate route for the app",
						"-space":        "Space will be used when constructing the candidate test route",
						"-timeout":      "Timeout for all the sub commands, example 10s, 2m37s",
					},
				},
			},
			{
				Name:     config.DELETE,
				HelpText: "Deprecated please use halfpipe-cleanup instead!",
			},
			{
				Name:     config.CLEANUP,
				HelpText: "Cleanups all apps that has the -DELETE postfix",
				UsageDetails: cfPlugin.Usage{
					Usage: "cf halfpipe-cleanup [-manifestPath PATH]",
					Options: map[string]string{
						"-manifestPath": "Relative or absolute path to cf manifest",
						"-timeout":      "Timeout for all the sub commands, example 10s, 2m37s",
					},
				},
			},
		},
	}
}

func main() {
	cfPlugin.Start(new(Halfpipe))
}
