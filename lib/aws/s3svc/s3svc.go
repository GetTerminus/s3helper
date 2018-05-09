package s3svc

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type api interface {
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
	ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
	ListObjectVersions(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)
}

type client struct {
	s3api   api
	verbose bool
}

// NewClient returns a client struct that provides an interface to the s3 api.
func NewClient(s3 api, v bool) *client {
	c := new(client)
	c.s3api = s3
	c.verbose = v
	return c
}

func (c *client) DeleteBucketContents(bucket *string) ([]*s3.DeletedObject, error) {
	deletedObjects := make([]*s3.DeletedObject, 0)

	for {
		// 2 API calls are needed to delete things in s3, the first will get a list
		// of the objects in the bucket, and the second to delete those things
		listResponse, listErr := c.s3api.ListObjectVersions(&s3.ListObjectVersionsInput{
			Bucket: bucket,
		})
		if listErr != nil {
			return nil, errors.Wrap(listErr, "error listing object versions on bucket")
		}

		// svc.ListObjectVersions will only return 1000 results at a time
		numObjects := len(listResponse.DeleteMarkers) + len(listResponse.Versions)

		// delete things until there's nothing left
		if numObjects == 0 {
			return deletedObjects, nil
		}

		// We're emptying the bucket, no need to discrimiate the things to delete
		var items s3.Delete
		var objectIdentifiers = make([]*s3.ObjectIdentifier, numObjects)

		for i, obj := range listResponse.Versions {
			objectIdentifiers[i] = &s3.ObjectIdentifier{
				Key:       obj.Key,
				VersionId: obj.VersionId,
			}
		}

		for i, obj := range listResponse.DeleteMarkers {
			objectIdentifiers[i+len(listResponse.Versions)] = &s3.ObjectIdentifier{
				Key:       obj.Key,
				VersionId: obj.VersionId,
			}
		}

		items.SetObjects(objectIdentifiers)

		// DeleteObjects can only delete 1000 things at a time
		deleteResponse, deleteErr := c.s3api.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: bucket,
			Delete: &items,
		})
		if deleteErr != nil {
			return nil, errors.Wrap(deleteErr, "error deleting objects")
		}

		deletedObjects = append(deletedObjects, deleteResponse.Deleted...)
	}
}

func (c *client) ListBucketContents(bucket *string) error {
	resp, _ := c.s3api.ListObjects(&s3.ListObjectsInput{
		Bucket: bucket,
	})

	fmt.Println(resp.Contents)

	return nil
}
