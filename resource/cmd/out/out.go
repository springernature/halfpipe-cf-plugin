package main

import (
	"io/ioutil"
	"os"
	"encoding/json"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"syscall"
	"log"
	"github.com/springernature/halfpipe-cf-plugin/resource/out/resource_plan"
)

func main() {

	concourseRoot := os.Args[0]

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

	p, err := resource_plan.NewPush().Plan(request, concourseRoot)
	if err != nil {
		logger.Println(err)
		syscall.Exit(1)
	}

	logger.Println(p)
	if err = p.Execute(out.NewCliExecutor(logger), logger); err != nil {
		syscall.Exit(1)
	}
}
