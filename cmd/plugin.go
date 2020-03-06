package main

import (
	"flag"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/check"
	"github.com/springernature/halfpipe-cf-plugin/cleanup"
	"github.com/springernature/halfpipe-cf-plugin/executor"
	"github.com/springernature/halfpipe-cf-plugin/promote"
	"github.com/springernature/halfpipe-cf-plugin/push"
	"log"
	"os"
	"strings"
	"syscall"

	cfPlugin "code.cloudfoundry.org/cli/plugin"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"time"
)

type Halfpipe struct{}

func parseArgs(args []string) (manifestPath string, appPath string, testDomain string, timeout time.Duration, preStartCommand string, instances int) {
	flagSet := flag.NewFlagSet("halfpipe", flag.ExitOnError)
	mP := flagSet.String("manifestPath", "", "Path to the manifest")
	aP := flagSet.String("appPath", "", "Path to the app")
	tD := flagSet.String("testDomain", "", "Domain to push the app to during the candidate stage")
	tO := flagSet.Duration("timeout", 5*time.Minute, "Timeout for each command")
	pS := flagSet.String("preStartCommand", "", "cf command to run before the application is started. Supports multiple commands semi-colon delimited.")
	iC := flagSet.Int("instances", 0, "Instances to deploy. Overrides value in manifest")
	if err := flagSet.Parse(args[1:]); err != nil {
		panic(err)
	}

	return *mP, *aP, *tD, *tO, *pS, *iC
}

func (Halfpipe) Run(cliConnection cfPlugin.CliConnection, args []string) {
	logger := log.New(os.Stdout, "", 0)
	logger.Printf("# CF plugin built from git revision '%s'\n", config.SHA)

	command := args[0]
	if command == "CLI-MESSAGE-UNINSTALL" {
		syscall.Exit(0)
	}

	manifestPath, appPath, testDomain, timeout, preStartCommand, instances := parseArgs(args)

	// not sure if this will ever happen in reality, but in the integration tests
	// we are given the string in quotes `"<value>"`
	if strings.HasPrefix(preStartCommand, `"`) && strings.HasSuffix(preStartCommand, `"`) {
		preStartCommand = preStartCommand[1 : len(preStartCommand)-1]
	}

	pluginRequest := halfpipe_cf_plugin.Request{
		Command:         command,
		ManifestPath:    manifestPath,
		AppPath:         appPath,
		TestDomain:      testDomain,
		Timeout:         timeout,
		PreStartCommand: preStartCommand,
		Instances:       instances,
	}

	if err := pluginRequest.Verify(); err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	planner := plan.NewPlanner(
		push.NewPushPlanner(cliConnection),
		check.NewCheckPlanner(10*time.Second),
		promote.NewPromotePlanner(cliConnection),
		cleanup.NewCleanupPlanner(cliConnection),
		manifest.NewManifestReadWrite(afero.Afero{Fs: afero.NewOsFs()}),
	)

	p, err := planner.GetPlan(pluginRequest)
	if err != nil {
		logger.Println("Failed to create execution plan!")
		logger.Println(err)
		syscall.Exit(1)
	}

	logger.Println(p)
	if err = executor.NewExecutor(executor.NewShellExecutor(logger), executor.NewCliExecutor(cliConnection, logger), pluginRequest.Timeout, logger).Execute(p); err != nil {
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
					Usage: "cf halfpipe-push [-manifestPath PATH] [-appPath PATH] [-testDomain DOMAIN] [-timeout TIMEOUT] [-preStartCommand CF COMMAND]",
					Options: map[string]string{
						"manifestPath":    "Relative or absolute path to cf manifest",
						"appPath":         "Relative or absolute path to the app bits you wish to deploy",
						"testDomain":      "Domain that will be used when constructing the candidate route for the app",
						"timeout":         "Timeout for all the sub commands, example 10s, 2m37s",
						"preStartCommand": "cf command to run before the application is started. Supports multiple commands semi-colon delimited.",
						"instances":       "how many instances to push the app with. If not specified it will read the instance count from the manifest",
					},
				},
			},
			{
				Name:     config.CHECK,
				HelpText: "Checks that all instances are in running state",
				UsageDetails: cfPlugin.Usage{
					Usage: "cf halfpipe-check [-manifestPath PATH] [-timeout TIMEOUT]",
					Options: map[string]string{
						"-manifestPath": "Relative or absolute path to cf manifest",
						"-timeout":      "Timeout for all the sub commands, example 10s, 2m37s",
					},
				},
			},
			{
				Name:     config.PROMOTE,
				HelpText: "Promotes the app from `my-app-candidate` to `my-app`, binds all the production routes, removes the test route and stops old instances of the app",
				UsageDetails: cfPlugin.Usage{
					Usage: "cf halfpipe-push [-manifestPath PATH] [-testDomain DOMAIN] [-timeout TIMEOUT]",
					Options: map[string]string{
						"-manifestPath": "Relative or absolute path to cf manifest",
						"-testDomain":   "Domain that will be used when constructing the candidate route for the app",
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
					Usage: "cf halfpipe-cleanup [-manifestPath PATH] [-timeout TIMEOUT]",
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
