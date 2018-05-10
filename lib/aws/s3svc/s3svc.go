package s3svc

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type API interface {
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
	ListObjectVersions(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)
}

type Client struct {
	s3api   API
	verbose bool
}

// NewClient returns a client struct that provides an interface to the s3 api.
func NewClient(s3 API, v bool) *Client {
	return &Client{
		s3api:   s3,
		verbose: v,
	}
}

// DeleteBucketContents will remove all object versions and delete markers from the specified s3 bucket.
func (c *Client) DeleteBucketContents(bucket string) ([]*s3.DeletedObject, error) {
	deletedObjects := make([]*s3.DeletedObject, 0)

	for {
		// 2 API calls are needed to delete things in s3, the first will get a list
		// of the objects in the bucket, and the second to delete those things
		objIDs, objIDErr := c.GetObjectIdentifiers(bucket)
		if objIDErr != nil {
			return nil, errors.Wrap(objIDErr, "package: s3svc => method: DeleteBucketContents => method call s3svc.Client.GetObjectIdentifiers failed\n")
		}

		// delete things until there's nothing left
		if len(objIDs) == 0 {
			return deletedObjects, nil
		}

		deleteIDs, deleteErr := c.DeleteObjects(bucket, objIDs)
		if deleteErr != nil {
			return nil, errors.Wrap(deleteErr, "package: s3svc => method: DeleteBucketContents => method call s3svc.Client.DeleteObjects failed\n")
		}

		deletedObjects = append(deletedObjects, deleteIDs...)
	}
}

// GetObjectIdentifiers returns object versions, delete markers, or both, from the specified s3 bucket.
func (c *Client) GetObjectIdentifiers(bucket string) ([]*s3.ObjectIdentifier, error) {
	resp, err := c.s3api.ListObjectVersions(&s3.ListObjectVersionsInput{Bucket: &bucket})
	if err != nil {
		return nil, errors.Wrap(err, "package: s3svc => method: GetObjectIdentifiers => method call s3api.ListObjectVersions failed\n")
	}

	objectIdentifiers := make([]*s3.ObjectIdentifier, 0)

	if resp.Versions != nil {
		objVersions := make([]*s3.ObjectIdentifier, len(resp.Versions))
		for i, obj := range resp.Versions {
			objVersions[i] = &s3.ObjectIdentifier{
				Key:       obj.Key,
				VersionId: obj.VersionId,
			}
		}

		objectIdentifiers = append(objectIdentifiers, objVersions...)
	}

	if resp.DeleteMarkers != nil {
		deadObjects := make([]*s3.ObjectIdentifier, len(resp.DeleteMarkers))
		for i, obj := range resp.DeleteMarkers {
			deadObjects[i] = &s3.ObjectIdentifier{
				Key:       obj.Key,
				VersionId: obj.VersionId,
			}
		}

		objectIdentifiers = append(objectIdentifiers, deadObjects...)
	}

	return objectIdentifiers, nil
}

// DeleteObjects deletes the specified objects in an s3 bucket, including delete markers.
func (c *Client) DeleteObjects(bucket string, objects []*s3.ObjectIdentifier) ([]*s3.DeletedObject, error) {
	var deleteList s3.Delete

	deleteList.SetObjects(objects)

	deleteResponse, deleteErr := c.s3api.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: &bucket,
		Delete: &deleteList,
	})
	if deleteErr != nil {
		return nil, errors.Wrap(deleteErr, "package: s3svc => method: DeleteObjects => method call s3api.DeleteObjects failed\n")
	}

	return deleteResponse.Deleted, nil
}
