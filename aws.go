package openaccess

import (
	"fmt"
)

var AWS_S3_BUCKET string
var AWS_S3_URI string
var AWS_S3_REGION string

func init() {
	AWS_S3_BUCKET = "smithsonian-open-access"
	AWS_S3_URI = fmt.Sprintf("https://%s.s3-us-west-2.amazonaws.com/", AWS_S3_BUCKET)
	AWS_S3_REGION = "us-west-2"
}
