package walk

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aaronland/go-json-query"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
	"io"
	"path/filepath"
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

	/*

		because this: https://github.com/Smithsonian/OpenAccess/issues/7

		if bucket scheme is s3:// then:

		if uri is objects or objects/metdata then:

		loop over list of SI units and invoke code for reading index.txt (below)

		else:

		fetch uri/index.txt and loop over each file handing off to default code

	*/

	// FIX ME - make me a real test please

	is_s3 := true

	if is_s3 {
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

	// TO DO: create a new bucket where the root URI is simply
	// openaccess.AWS_S3_URI so we don't have to do a bunch of
	// URI/path checking below

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

	index := fmt.Sprintf("metadata/%s/index.txt", unit)

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

	// FIX ME: uri is still a fully qualified URL

	fh, err := bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer fh.Close()

	// this is untested...

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
