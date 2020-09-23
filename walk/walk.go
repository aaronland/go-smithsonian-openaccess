package walk

import (
	"context"
	"github.com/aaronland/go-json-query"
	jw "github.com/aaronland/go-jsonl/walk"
	"gocloud.dev/blob"
	_ "io"
)

type WalkOptions struct {
	URI          string
	Workers      int
	ValidateJSON bool
	FormatJSON   bool
	QuerySet     *query.QuerySet
	Callback     WalkRecordCallbackFunc
	IsBzip bool
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
		IsBzip: opts.IsBzip,
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
