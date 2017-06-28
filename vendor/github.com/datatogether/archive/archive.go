/*
Archive holds all common model definitions for archivers 2.0.
*/
package archive

import (
	"os"
	"time"
)

var (
	// how long before a url is considered stale. default is 72 hours.
	StaleDuration = time.Hour * 72
	// all these need to be set for file saving to work
	AwsRegion          string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsS3BucketName    string
	AwsS3BucketPath    string
)

func init() {
	AwsRegion = os.Getenv("AWS_REGION")
	AwsAccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AwsS3BucketName = os.Getenv("AWS_S3_BUCKET_NAME")
	AwsS3BucketPath = os.Getenv("AWS_S3_BUCKET_PATH")
}
