package main

import (
	"fmt"
	"os"

	"github.com/OneOf-Inc/firefly-tezosconnect/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
