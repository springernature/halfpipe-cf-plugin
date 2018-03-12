package main

import (
	"fmt"
	"os"
	"io/ioutil"
)

func main() {
	// Just echo stdin to stdout.

	indata, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, string(indata))
}
