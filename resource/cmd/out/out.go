package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/plan/plans"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"github.com/springernature/halfpipe-cf-plugin/resource/out/resource_plan"
)

func main() {
	concourseRoot := os.Args[1]

	logger := log.New(os.Stderr, "", 0)

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	request := out.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	var p plans.Plan
	switch request.Params.Command {
	case "":
		panic("params.command must not be empty")
	case halfpipe_cf_plugin.PUSH, halfpipe_cf_plugin.PROMOTE:
		p, err = resource_plan.NewPlan().Plan(request, concourseRoot)
	default:
		panic(fmt.Sprintf("Command '%s' not supported", request.Params.Command))
	}

	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	if err = p.Execute(out.NewCliExecutor(), logger); err != nil {
		os.Exit(1)
	}

	response := out.Response{
		Version: out.Version{
			Timestamp: time.Now(),
		},
		Metadata: []out.MetadataPair{
			{Name: "Api", Value: request.Source.Api},
			{Name: "Org", Value: request.Source.Org},
			{Name: "Space", Value: request.Source.Space},
		},
	}
	json.NewEncoder(os.Stdout).Encode(response)
}
