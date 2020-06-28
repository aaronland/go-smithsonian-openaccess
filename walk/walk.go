package walk

import (
	"context"
	jw "github.com/aaronland/go-jsonl/walk"
	"gocloud.dev/blob"
	_ "log"
)

type WalkOptions struct {
	URI          string
	Workers      int
	ValidateJSON bool
	FormatJSON   bool
	QuerySet     *jw.WalkQuerySet
	Callback     WalkRecordCallbackFunc
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
		Format:        opts.FormatJSON,
		Validate:      opts.ValidateJSON,
		QuerySet:      opts.QuerySet,
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
