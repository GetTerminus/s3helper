package main

import (
	"fmt"
	"os"

	_ "github.com/GetTerminus/s3helper/commands"
	"github.com/GetTerminus/s3helper/lib/parser"
)

func main() {
	_, err := parser.OptParser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
