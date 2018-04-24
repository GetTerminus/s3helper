package parser

import (
	"github.com/jessevdk/go-flags"
)

var GlobalOpts struct {
	Region  string `short:"r" long:"region" description:"The region the s3 bucket resides in" required:"false" default:"us-east-1"`
	Profile string `short:"p" long:"profile" description:"AWS Credential profile" required:"false"`
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output" required:"false"`
}

var OptParser = flags.NewParser(&GlobalOpts, flags.HelpFlag)
