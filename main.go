package main

import (
	"fmt"
	"os"

	"github.com/dyl0115/speakr/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
