package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/plan/resource"
	"github.com/springernature/halfpipe-cf-plugin/config"
)

func main() {
	concourseRoot := os.Args[1]

	started := time.Now()

	logger := log.New(os.Stderr, "", 0)

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	request := resource.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	var p plan.Plan
	switch request.Params.Command {
	case "":
		panic("params.command must not be empty")
	case config.PUSH, config.PROMOTE, config.DELETE, config.CLEANUP:
		p, err = resource.NewPlanner(manifest.ReadAndMergeManifests, manifest.WriteApplicationManifest).Plan(request, concourseRoot)
	default:
		panic(fmt.Sprintf("Command '%s' not supported", request.Params.Command))
	}

	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	if err = p.Execute(resource.NewCFCliExecutor(), logger); err != nil {
		os.Exit(1)
	}

	finished := time.Now()

	response := resource.Response{
		Version: resource.Version{
			Timestamp: finished,
		},
		Metadata: []resource.MetadataPair{
			{Name: "Api", Value: request.Source.API},
			{Name: "Org", Value: request.Source.Org},
			{Name: "Space", Value: request.Source.Space},
			{Name: "Duration", Value: finished.Sub(started).String()},
		},
	}
	if err = json.NewEncoder(os.Stdout).Encode(response); err != nil {
		panic(err)
	}
}
