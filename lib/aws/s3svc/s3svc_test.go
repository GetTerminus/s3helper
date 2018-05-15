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

		actualErr error

		listObjVersionOutput      *s3.ListObjectVersionsOutput
		listObjVersionOutputError error

		deleteObjectsOutput      *s3.DeleteObjectsOutput
		deleteObjectsOutputError error
	)

	BeforeEach(func() {
		bucket = "fake_bucket"
		fakeS3 = &s3svcfakes.FakeAPI{}
		s3Client = s3svc.NewClient(fakeS3, false)
	})

	Describe("GetObjectIdentifiers", func() {
		var (
			actualResp   []*s3.ObjectIdentifier
			expectedResp []*s3.ObjectIdentifier
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
			actualResp   []*s3.DeletedObject
			expectedResp []*s3.DeletedObject

			deleteObjectIDs []*s3.ObjectIdentifier
		)

		JustBeforeEach(func() {
			fakeS3.DeleteObjectsReturns(deleteObjectsOutput, deleteObjectsOutputError)
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
					deleteObjectsOutputError = errors.New("fail")
				})

				It("should return an error", func() {
					Expect(actualResp).To(BeNil())
					Expect(actualErr).NotTo(BeNil())
				})
			})
		})
	})

	Describe("DeleteBucketContents", func() {
		var (
			actualResp   []*s3.DeletedObject
			expectedResp []*s3.DeletedObject
		)

		JustBeforeEach(func() {
			fakeS3.ListObjectVersionsReturns(listObjVersionOutput, listObjVersionOutputError)
			fakeS3.DeleteObjectsReturns(deleteObjectsOutput, deleteObjectsOutputError)
			actualResp, actualErr = s3Client.DeleteBucketContents(bucket)
		})

		Context("when everything works", func() {
			BeforeEach(func() {
				listObjVersionOutput = &s3.ListObjectVersionsOutput{}
				listObjVersionOutputError = nil

				deleteObjectsOutput = &s3.DeleteObjectsOutput{}
				deleteObjectsOutputError = nil
			})

			Context("when the bucket is already empty", func() {
				BeforeEach(func() {
					expectedResp = []*s3.DeletedObject{}
				})

				It("should not produce an error", func() {
					Expect(actualErr).To(BeNil())
				})

				It("should return an emtpy list of s3.DeletedObjects", func() {
					Expect(actualResp).To(Equal(expectedResp))
				})

				It("should not call the DeleteObjects method", func() {
					Expect(fakeS3.DeleteObjectsCallCount()).To(Equal(0))
				})
			})

			Context("when the bucket is not emtpy", func() {
				Context("and the results from s3.ListObjectVersions are paginated", func() {
					BeforeEach(func() {
						fakeS3.ListObjectVersionsReturnsOnCall(0,
							&s3.ListObjectVersionsOutput{
								Versions: []*s3.ObjectVersion{
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
								},
							},
							listObjVersionOutputError)

						fakeS3.ListObjectVersionsReturnsOnCall(1,
							&s3.ListObjectVersionsOutput{
								DeleteMarkers: []*s3.DeleteMarkerEntry{
									&s3.DeleteMarkerEntry{
										IsLatest:     aws.Bool(true),
										Key:          aws.String("obj1"),
										LastModified: aws.Time(time.Now()),
										Owner: &s3.Owner{
											DisplayName: aws.String("GetTerminus"),
											ID:          aws.String("7c799faaab6148a7b7104bc9a3835f65"),
										},
										VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
									},
								},
							},
							listObjVersionOutputError)

						fakeS3.ListObjectVersionsReturnsOnCall(2, listObjVersionOutput, listObjVersionOutputError)

						fakeS3.DeleteObjectsReturnsOnCall(0,
							&s3.DeleteObjectsOutput{
								Deleted: []*s3.DeletedObject{
									&s3.DeletedObject{
										Key:       aws.String("obj1"),
										VersionId: aws.String("c3ecadb1b13d4278a67a9a07927222b2"),
									},
								},
							},
							deleteObjectsOutputError)

						fakeS3.DeleteObjectsReturnsOnCall(1,
							&s3.DeleteObjectsOutput{
								Deleted: []*s3.DeletedObject{
									&s3.DeletedObject{
										DeleteMarker:          aws.Bool(true),
										DeleteMarkerVersionId: aws.String("1a624c9c02b74b09845f0623521ea571"),
										Key:       aws.String("obj1"),
										VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
									},
								},
							},
							deleteObjectsOutputError)

						expectedResp = []*s3.DeletedObject{
							&s3.DeletedObject{
								Key:       aws.String("obj1"),
								VersionId: aws.String("c3ecadb1b13d4278a67a9a07927222b2"),
							},
							&s3.DeletedObject{
								DeleteMarker:          aws.Bool(true),
								DeleteMarkerVersionId: aws.String("1a624c9c02b74b09845f0623521ea571"),
								Key:       aws.String("obj1"),
								VersionId: aws.String("5a8a51ef218747c0aafddc3009c22cf7"),
							},
						}
					})

					It("should call ListObjectVersions 3 times", func() {
						Expect(fakeS3.ListObjectVersionsCallCount()).To(Equal(3))
					})

					It("should call DeleteObjects 2 times", func() {
						Expect(fakeS3.DeleteObjectsCallCount()).To(Equal(2))
					})

					It("should delete the contents of the bucket", func() {
						Expect(actualResp).To(ConsistOf(expectedResp))
					})
				})
			})
		})

		Context("when an error occurs", func() {
			Context("when s3.ListObjectVersions returns an error", func() {
				BeforeEach(func() {
					listObjVersionOutput = nil
					listObjVersionOutputError = errors.New("fail")

					deleteObjectsOutput = nil
					deleteObjectsOutputError = nil
				})

				It("returns an error", func() {
					Expect(actualResp).To(BeNil())
					Expect(actualErr).NotTo(BeNil())
				})

				It("does not call s3.DeleteObjects", func() {
					Expect(fakeS3.DeleteObjectsCallCount()).To(Equal(0))
				})
			})

			Context("when s3.DeleteObjecs returns an error", func() {
				BeforeEach(func() {
					listObjVersionOutput = &s3.ListObjectVersionsOutput{
						Versions: []*s3.ObjectVersion{
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
						},
					}
					listObjVersionOutputError = nil

					deleteObjectsOutput = nil
					deleteObjectsOutputError = errors.New("fail")
				})

				It("returns an error", func() {
					Expect(actualResp).To(BeNil())
					Expect(actualErr).NotTo(BeNil())
				})

				It("calls s3.ListObjectVersionsCallCount only once", func() {
					Expect(fakeS3.ListObjectVersionsCallCount()).To(Equal(1))
				})
			})
		})
	})
})
