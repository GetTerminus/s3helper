package aws

import (
	"sync"

	"github.com/GetTerminus/s3helper/lib/parser"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type client struct {
	awsSession     *session.Session
	awsSessionOnce sync.Once
}

// Session returns an authenticated AWS session, used to make API calls.
func (c *client) Session() *session.Session {

	// Only ever create one session
	c.awsSessionOnce.Do(func() {
		// Create a session, session.Must should handle any errors
		// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/#Must
		c.awsSession = session.Must(
			session.NewSession(
				&aws.Config{
					Region:      aws.String(parser.GlobalOpts.Region),
					Credentials: processCredentials(parser.GlobalOpts.Profile),
				},
			),
		)
	})

	return c.awsSession
}

func processCredentials(profile string) *credentials.Credentials {
	if profile != "" {
		return credentials.NewSharedCredentials("", profile)
	}

	return credentials.NewEnvCredentials()
}

// Client is the singleton instance of the client struct.
var Client client
