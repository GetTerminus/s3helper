package commands

import (
	"fmt"
	"os"

	"github.com/GetTerminus/s3helper/lib/aws"
	"github.com/GetTerminus/s3helper/lib/aws/s3svc"
	"github.com/GetTerminus/s3helper/lib/parser"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// EmptyBucketCommand represents the options that can be passed to the empty-bucket subcommand.
type EmptyBucketCommand struct {
	Bucket string `short:"b" long:"bucket" value-name:"bucket" description:"the bucket to empty" required:"true"`
}

func init() {
	var cmd EmptyBucketCommand

	// nolint [:errcheck]
	parser.OptParser.AddCommand(
		"empty-bucket",
		"Erase all contents in an s3 bucket",
		"Erase all objects and delete markers in an s3 bucket",
		&cmd,
	)
}

// Execute implements the interface for the go-flags subcommand.
func (cmd *EmptyBucketCommand) Execute(args []string) error {

	// nolint [:gas]
	fmt.Fprintf(os.Stdout, "Deleting contents of s3://%s\n", cmd.Bucket)

	awsSession := aws.Client.GetSession()
	s3client := s3svc.NewClient(s3.New(awsSession), parser.GlobalOpts.Verbose)

	resp, err := s3client.DeleteBucketContents(cmd.Bucket)
	if err != nil {
		return errors.Wrap(err, "Package: commands => func: Execute => method call s3svc.Client.DeleteBucketContents failed\n")
	}

	if parser.GlobalOpts.Verbose {

		// nolint [:gas]
		fmt.Fprintln(os.Stdout, resp)
	}

	return nil
}
