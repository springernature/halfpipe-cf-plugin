package main

import (
	"os"
	"encoding/json"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"time"
	"io/ioutil"
	"syscall"
	"log"
	"github.com/springernature/halfpipe-cf-plugin/resource/out/resource_plan"
	"github.com/springernature/halfpipe-cf-plugin/color"
)

func main() {
	concourseRoot := os.Args[1]

	logger := log.New(os.Stderr, "", 0)

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Println(color.ErrColor.Sprint(err))
		syscall.Exit(1)
	}

	request := out.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		logger.Println(color.ErrColor.Sprint(err))
		syscall.Exit(1)
	}

	p, err := resource_plan.NewPush().Plan(request, concourseRoot)
	if err != nil {
		logger.Println(color.ErrColor.Sprint(err))
		syscall.Exit(1)
	}

	if err = p.Execute(out.NewCliExecutor(), logger, color.ResourcePlanColor); err != nil {
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
