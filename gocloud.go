package openaccess

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	_ "log"
	"net/url"
)

const IS_SMITHSONIAN_S3 string = "github.com/aaronland/go-smithsonian-openaccess#is_smithsonian_s3"

// Special-case bucket opener to deal with setting the necessary flags to know how
// to fetch data fromt the `smithsonian-open-access` S3 bucket. Note how we are returning
// a new context.Context with signals for the rest of the code to use (20201119/straup)

func OpenBucket(ctx context.Context, uri string) (context.Context, *blob.Bucket, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, nil, err
	}

	is_smithsonian_s3 := false

	switch u.Scheme {
	case "s3":

		if u.Host == AWS_S3_BUCKET {
			is_smithsonian_s3 = true
		}

	case "si":

		uri = fmt.Sprintf("s3://%s?region=%s", AWS_S3_BUCKET, AWS_S3_REGION)
		is_smithsonian_s3 = true

	default:
		// pass
	}

	var bucket *blob.Bucket

	if is_smithsonian_s3 {

		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(AWS_S3_REGION),
			Credentials: credentials.AnonymousCredentials,
		})

		if err != nil {
			return nil, nil, err
		}

		// SKIPMETADATA GOES HERE

		b, err := s3blob.OpenBucket(ctx, sess, AWS_S3_BUCKET, nil)

		if err != nil {
			return nil, nil, err
		}

		bucket = b

	} else {

		b, err := blob.OpenBucket(ctx, uri)

		if err != nil {
			return nil, nil, err
		}

		bucket = b
	}

	ctx = context.WithValue(ctx, IS_SMITHSONIAN_S3, is_smithsonian_s3)
	return ctx, bucket, nil
}

func OpenMetadataBucket(ctx context.Context, uri string) (context.Context, *blob.Bucket, error) {

	ctx, bucket, err := OpenBucket(ctx, uri)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open metadata bucket, %w", err)
	}

	bucket = blob.PrefixedBucket(bucket, "metadata/edan")
	return ctx, bucket, nil
}
