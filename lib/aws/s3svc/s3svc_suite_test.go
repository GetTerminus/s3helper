package s3svc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestS3svc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3svc Suite")
}
