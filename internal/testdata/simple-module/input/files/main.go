package main

import (
	"fmt"
	"os"

	"{{ .Module }}/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
