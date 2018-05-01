package provider

import (
	"github.com/GetTerminus/s3helper/lib/parser"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
)

// AWSSession returns an authenticated AWS session, used to make API calls.
func AWSSession() (*session.Session, error) {
	// Get credentials
	var creds *credentials.Credentials

	if parser.GlobalOpts.Profile != "" {
		creds = credentials.NewSharedCredentials("", parser.GlobalOpts.Profile)
	} else {
		creds = credentials.NewEnvCredentials()
	}

	sessionHandle, err := session.NewSession(&aws.Config{
		Region:      aws.String(parser.GlobalOpts.Region),
		Credentials: creds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating session with the specified credentials")
	}

	return sessionHandle, nil
}
