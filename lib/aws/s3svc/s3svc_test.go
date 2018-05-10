package s3svc_test

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/GetTerminus/s3helper/lib/aws/s3svc"
	"github.com/GetTerminus/s3helper/lib/aws/s3svc/s3svcfakes"
)

var _ = Describe("S3svc", func() {
	var (
		bucket   string
		fakeS3   *s3svcfakes.FakeAPI
		s3Client *s3svc.Client
	)

	BeforeEach(func() {
		bucket = "fake_bucket"
		fakeS3 = &s3svcfakes.FakeAPI{}
		s3Client = s3svc.NewClient(fakeS3, false)
	})

	Describe("GetObjectIdentifiers", func() {
		var (
			actualErr                 error
			actualResp                []*s3.ObjectIdentifier
			expectedResp              []*s3.ObjectIdentifier
			listObjVersionOutput      *s3.ListObjectVersionsOutput
			listObjVersionOutputError error
		)

		JustBeforeEach(func() {
			fakeS3.ListObjectVersionsReturns(listObjVersionOutput, listObjVersionOutputError)
			actualResp, actualErr = s3Client.GetObjectIdentifiers(bucket)
		})

		Context("when everything is fine", func() {
			BeforeEach(func() {
				listObjVersionOutput = &s3.ListObjectVersionsOutput{
					IsTruncated:     aws.Bool(false),
					KeyMarker:       aws.String(""),
					MaxKeys:         aws.Int64(1000),
					Name:            aws.String(bucket),
					Prefix:          aws.String(""),
					VersionIdMarker: aws.String(""),
				}
				listObjVersionOutputError = nil
			})

			It("should not produce an error", func() {
				Expect(actualErr).To(BeNil())
			})

			Context("when there are no objects in the bucket", func() {
				BeforeEach(func() {
					expectedResp = []*s3.ObjectIdentifier{}
				})

				It("should return an empty slice", func() {
					Expect(actualResp).NotTo(BeNil())
					Expect(actualResp).To(Equal(expectedResp))
				})
			})

			Context("when there are only object versions", func() {
				BeforeEach(func() {
					listObjVersionOutput.Versions = []*s3.ObjectVersion{
						&s3.ObjectVersion{
							ETag:         aws.String("\"4dc5bf55892540efa31afa0463667a82\""),
							Key:          aws.String("index"),
							LastModified: aws.Time(time.Now()),
							Owner: &s3.Owner{
								DisplayName: aws.String("GetTerminus"),
								ID:          aws.String("7c799faaab6148a7b7104bc9a3835f65"),
							},
							Size:         aws.Int64(1024),
							StorageClass: aws.String("STANDARD"),
							VersionId:    aws.String("null"),
						},
					}

					expectedResp = []*s3.ObjectIdentifier{
						&s3.ObjectIdentifier{
							Key:       aws.String("index"),
							VersionId: aws.String("null"),
						},
					}
				})

				It("should produce a slice of *s3.ObjectVersion", func() {
					Expect(actualResp).NotTo(BeNil())
					Expect(actualResp).To(Equal(expectedResp))
				})
			})

			Context("when there are only delete markers", func() {
				BeforeEach(func() {
					listObjVersionOutput.DeleteMarkers = []*s3.DeleteMarkerEntry{
						&s3.DeleteMarkerEntry{
							IsLatest:  aws.Bool(true),
							Key:       aws.String("index"),
							VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
						},
					}

					expectedResp = []*s3.ObjectIdentifier{
						&s3.ObjectIdentifier{
							Key:       aws.String("index"),
							VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
						},
					}
				})

				It("should produce a slice of *s3.ObjectVersion", func() {
					Expect(actualResp).NotTo(BeNil())
					Expect(actualResp).To(Equal(expectedResp))
				})
			})

			Context("when there are both object versions and delete markers", func() {
				BeforeEach(func() {
					listObjVersionOutput.Versions = []*s3.ObjectVersion{
						&s3.ObjectVersion{
							ETag:         aws.String("\"4dc5bf55892540efa31afa0463667a82\""),
							Key:          aws.String("obj1"),
							LastModified: aws.Time(time.Now()),
							Owner: &s3.Owner{
								DisplayName: aws.String("GetTerminus"),
								ID:          aws.String("7c799faaab6148a7b7104bc9a3835f65"),
							},
							Size:         aws.Int64(1024),
							StorageClass: aws.String("STANDARD"),
							VersionId:    aws.String("c3ecadb1b13d4278a67a9a07927222b2"),
						},
					}

					listObjVersionOutput.DeleteMarkers = []*s3.DeleteMarkerEntry{
						&s3.DeleteMarkerEntry{
							IsLatest:     aws.Bool(true),
							Key:          aws.String("dm1"),
							LastModified: aws.Time(time.Now()),
							Owner: &s3.Owner{
								DisplayName: aws.String("GetTerminus"),
								ID:          aws.String("7c799faaab6148a7b7104bc9a3835f65"),
							},
							VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
						},
					}

					expectedResp = []*s3.ObjectIdentifier{
						&s3.ObjectIdentifier{
							Key:       aws.String("dm1"),
							VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
						},
						&s3.ObjectIdentifier{
							Key:       aws.String("obj1"),
							VersionId: aws.String("c3ecadb1b13d4278a67a9a07927222b2"),
						},
					}
				})

				It("should produce a slice of *s3.ObjectVersion", func() {
					Expect(actualResp).NotTo(BeNil())
					Expect(actualResp).To(ConsistOf(expectedResp))
				})
			})
		})

		Context("when s3.ListObjectVersions fails", func() {
			BeforeEach(func() {
				listObjVersionOutputError = errors.New("fail")
				expectedResp = nil
			})

			It("should produce an error", func() {
				Expect(actualErr).NotTo(BeNil())
			})

			It("should not return any s3.ObjectIdentifiers", func() {
				Expect(actualResp).To(BeNil())
			})
		})
	})

	Describe("DeleteObjects", func() {
		var (
			actualErr           error
			actualResp          []*s3.DeletedObject
			expectedResp        []*s3.DeletedObject
			deleteObjectsError  error
			deleteObjectIDs     []*s3.ObjectIdentifier
			deleteObjectsOutput *s3.DeleteObjectsOutput
		)

		JustBeforeEach(func() {
			fakeS3.DeleteObjectsReturns(deleteObjectsOutput, deleteObjectsError)
			actualResp, actualErr = s3Client.DeleteObjects(bucket, deleteObjectIDs)
		})

		Context("when all is well", func() {
			BeforeEach(func() {
				deleteObjectsOutput = &s3.DeleteObjectsOutput{
					Deleted: []*s3.DeletedObject{
						&s3.DeletedObject{
							Key:       aws.String("key"),
							VersionId: aws.String("0ec634937b5f403f9c1a4cb5b39036d8"),
						},
						&s3.DeletedObject{
							DeleteMarker:          aws.Bool(true),
							DeleteMarkerVersionId: aws.String("a152b504be6a414488c12abe9a14390e"),
							Key:       aws.String("Key"),
							VersionId: aws.String("db8d5c24983d47caa1e298512f2e8e8b"),
						},
					},
				}

				expectedResp = []*s3.DeletedObject{
					&s3.DeletedObject{
						Key:       aws.String("key"),
						VersionId: aws.String("0ec634937b5f403f9c1a4cb5b39036d8"),
					},
					&s3.DeletedObject{
						DeleteMarker:          aws.Bool(true),
						DeleteMarkerVersionId: aws.String("a152b504be6a414488c12abe9a14390e"),
						Key:       aws.String("Key"),
						VersionId: aws.String("db8d5c24983d47caa1e298512f2e8e8b"),
					},
				}
			})

			It("should produce a slice of *s3.DeletedObject", func() {
				Expect(actualErr).To(BeNil())
				Expect(actualResp).To(ConsistOf(expectedResp))
			})
		})

		Context("when s3.DeleteObjects fails", func() {
			Context("when there is nothing to delete", func() {
				BeforeEach(func() {
					deleteObjectsError = errors.New("fail")
					expectedResp = nil
				})

				It("should return an error", func() {
					Expect(actualResp).To(BeNil())
					Expect(actualErr).NotTo(BeNil())
				})
			})
		})
	})
})
