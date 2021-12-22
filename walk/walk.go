package walk

import (
	"context"
	"github.com/aaronland/go-json-query"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
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

func WalkSmithsonianRecord(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, uri string) error {

	if !openaccess.IsMetaDataFile(uri){
		return nil
	}
	
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

// deprecated - keeping it around for a bit just in case
// (20201119/straup)

/*

func WalkSmithsonianBucketWithIndexForUnit(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, unit string) error {

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

		err = WalkSmithsonianRecord(ctx, opts, bucket, uri)

		if err != nil {
			return err
		}
	}

	return nil
}

*/
