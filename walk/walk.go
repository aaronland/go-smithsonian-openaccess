package walk

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aaronland/go-json-query"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"io"
	_ "log"
	"path/filepath"
	"strings"
)

type WalkOptions struct {
	URI          string
	Workers      int
	ValidateJSON bool
	FormatJSON   bool
	QuerySet     *query.QuerySet
	Callback     WalkRecordCallbackFunc
	IsBzip       bool
}

type WalkRecordCallbackFunc func(context.Context, *jw.WalkRecord, error) error

func WalkBucket(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	// Because the GitHub repo is too large to check out we want
	// to be able to query the corresponding S3 bucket with the same
	// files but since those buckets disallow directory listings we
	// need do things outside the normal bucket.List abstraction
	// (20201119/straup)
	
	var s3_bucket *s3.S3

	if bucket.As(&s3_bucket) {
		return WalkS3Bucket(ctx, opts, bucket)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cb := opts.Callback

	jw_record_ch := make(chan *jw.WalkRecord)
	jw_error_ch := make(chan *jw.WalkError)

	jw_opts := &jw.WalkOptions{
		URI:           opts.URI,
		Workers:       opts.Workers,
		RecordChannel: jw_record_ch,
		ErrorChannel:  jw_error_ch,
		FormatJSON:    opts.FormatJSON,
		ValidateJSON:  opts.ValidateJSON,
		QuerySet:      opts.QuerySet,
		IsBzip:        opts.IsBzip,
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-jw_error_ch:
				cb(ctx, nil, err)
			case rec := <-jw_record_ch:
				cb(ctx, rec, nil)
			default:
				// pass
			}
		}
	}()

	return jw.WalkBucket(ctx, jw_opts, bucket)
}

func WalkS3Bucket(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	// create a new bucket where the root URI is simply
	// openaccess.AWS_S3_BUCKET so we don't have to do a bunch of
	// URI/path checking below

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(openaccess.AWS_S3_REGION),
		Credentials: credentials.AnonymousCredentials,
	})

	if err != nil {
		return err
	}

	new_bucket, err := s3blob.OpenBucket(ctx, sess, openaccess.AWS_S3_BUCKET, nil)

	if err != nil {
		return err
	}

	bucket = new_bucket

	uri := opts.URI
	base := filepath.Base(uri)

	switch base {
	case "metadata":
		return WalkS3BucketForAll(ctx, opts, bucket)
	default:
		return WalkS3BucketForUnit(ctx, opts, bucket, base)
	}

}

func WalkS3BucketForAll(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	for _, unit := range openaccess.SMITHSONIAN_UNITS {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		err := WalkS3BucketForUnit(ctx, opts, bucket, unit)

		if err != nil {
			return err
		}
	}

	return nil
}

func WalkS3BucketForUnit(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, unit string) error {

	// https://github.com/Smithsonian/OpenAccess/issues/7#issuecomment-696833714

	digits := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	letters := []string{"a", "b", "c", "d", "e", "f"}

	files := make([]string, 0)

	for _, first := range digits {

		for _, second := range digits {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}
	}

	for _, first := range letters {

		for _, second := range digits {

			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}
	}

	unit = strings.ToLower(unit)

	for _, fname := range files {

		uri := fmt.Sprintf("metadata/edan/%s/%s", unit, fname)

		err := WalkS3Record(ctx, opts, bucket, uri)

		if err != nil {
			return err
		}
	}

	return nil
}

// deprecated - keeping it around for a bit just in case
// (20201119/straup)

func WalkS3BucketWithIndexForUnit(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, unit string) error {

	unit = strings.ToLower(unit)
	index := fmt.Sprintf("metadata/edan/%s/index.txt", unit)

	fh, err := bucket.NewReader(ctx, index, nil)

	if err != nil {
		return err
	}

	defer fh.Close()

	reader := bufio.NewReader(fh)

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		uri, err := reader.ReadString('\n')

		if err != nil {

			if err == io.EOF {
				break
			} else {
				continue
			}
		}

		uri = strings.TrimSpace(uri)
		uri = strings.Replace(uri, openaccess.AWS_S3_URI, "", 1)

		fmt.Println(uri)
		continue

		err = WalkS3Record(ctx, opts, bucket, uri)

		if err != nil {
			return err
		}
	}

	return nil
}

func WalkS3Record(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, uri string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fh, err := bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer fh.Close()

	cb := opts.Callback

	jw_record_ch := make(chan *jw.WalkRecord)
	jw_error_ch := make(chan *jw.WalkError)

	jw_opts := &jw.WalkOptions{
		URI:           opts.URI,
		Workers:       opts.Workers,
		RecordChannel: jw_record_ch,
		ErrorChannel:  jw_error_ch,
		FormatJSON:    opts.FormatJSON,
		ValidateJSON:  opts.ValidateJSON,
		QuerySet:      opts.QuerySet,
		IsBzip:        opts.IsBzip,
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-jw_error_ch:
				cb(ctx, nil, err)
			case rec := <-jw_record_ch:
				cb(ctx, rec, nil)
			default:
				// pass
			}
		}
	}()

	jw.WalkReader(ctx, jw_opts, fh)
	return nil
}
