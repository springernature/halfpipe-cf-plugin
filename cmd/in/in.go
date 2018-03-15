package main

import (
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Just echo stdin to stdout.
	logger := log.New(os.Stderr, "", 0)

	inData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Println(err.Error())
		os.Exit(1)
	}

	logger.SetOutput(os.Stdout)
	logger.Println(string(inData))
}
