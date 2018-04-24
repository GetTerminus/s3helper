package commands

import (
	"fmt"
	"os"

	"github.com/GetTerminus/s3helper/lib/parser"
	"github.com/GetTerminus/s3helper/lib/provider"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// EmptyBucketCommand represents the options that can be passed to the empty-bucket subcommand.
type EmptyBucketCommand struct {
	Bucket string `short:"b" long:"bucket" value-name:"bucket" description:"the bucket to empty" required:"true"`
}

func init() {
	var cmd EmptyBucketCommand
	parser.OptParser.AddCommand(
		"empty-bucket",
		"Erase all contents in an s3 bucket",
		"Erase all objects and delete markers in an s3 bucket",
		&cmd,
	)
}

// Execute implements the interface for the go-flags subcommand.
func (cmd *EmptyBucketCommand) Execute(args []string) error {
	fmt.Fprintf(os.Stdout, "Deleting contents of s3://%s\n", cmd.Bucket)

	// Create a session and get an s3 handle
	session, err := provider.AWSSession()
	if err != nil {
		return errors.Wrap(err, "error getting aws session")
	}

	svc := s3.New(session)

	totalDeleted := 0

	for {
		// 2 API calls are needed to delete things in s3, the first will get a list
		// of the objects in the bucket, and the second to delete those things
		resp, err := svc.ListObjectVersions(&s3.ListObjectVersionsInput{
			Bucket: aws.String(cmd.Bucket),
		})
		if err != nil {
			return errors.Wrap(err, "error listing object versions on bucket")
		}

		// svc.ListObjectVersions will only return 1000 results at a time
		numObjects := len(resp.DeleteMarkers) + len(resp.Versions)
		totalDeleted += numObjects

		// delete things until there's nothing left
		if numObjects == 0 {
			fmt.Fprintf(os.Stdout, "Total number of objects deleted: %v\n", totalDeleted)
			return nil
		}

		// We're emptying the bucket, no need to discrimiate the things to delete
		var items s3.Delete
		var objectIdentifiers = make([]*s3.ObjectIdentifier, numObjects)

		for i, obj := range resp.Versions {
			objectIdentifiers[i] = &s3.ObjectIdentifier{
				Key:       obj.Key,
				VersionId: obj.VersionId,
			}
		}

		for i, obj := range resp.DeleteMarkers {
			objectIdentifiers[i+len(resp.Versions)] = &s3.ObjectIdentifier{
				Key:       aws.String(*obj.Key),
				VersionId: aws.String(*obj.VersionId),
			}
		}

		items.SetObjects(objectIdentifiers)

		fmt.Fprintf(os.Stdout, "Deleting %v objects\n", numObjects)

		if parser.GlobalOpts.Verbose {
			fmt.Fprintln(os.Stdout, items)
		}

		// DeleteObjects can only delete 1000 things at a time
		_, err = svc.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(cmd.Bucket),
			Delete: &items,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting objects")
		}
	}
}
