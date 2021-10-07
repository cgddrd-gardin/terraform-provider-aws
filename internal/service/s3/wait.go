package s3

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	bucketCreatedTimeout = 2 * time.Minute
	propagationTimeout   = 1 * time.Minute
)

// waitRetryWhenBucketNotFound retries the specified function if the returned error indicates that a bucket is not found.
// If the retries time out the specified function is called one last time.
func waitRetryWhenBucketNotFound(f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, f, s3.ErrCodeNoSuchBucket)
}
