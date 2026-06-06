package walk

import (
	"context"

	"github.com/aaronland/go-json-query"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
)

type WalkOptions struct {
	Workers      int
	ValidateJSON bool
	FormatJSON   bool
	QuerySet     *query.QuerySet
	Callback     WalkRecordCallbackFunc
	IsBzip       bool
	Filter       jw.WalkFilterFunc
}

type WalkRecordCallbackFunc func(context.Context, *jw.WalkRecord, error) error

func WalkBucket(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	iter_opts := &jw.IterateOptions{
		ValidateJSON: opts.ValidateJSON,
		FormatJSON:   opts.FormatJSON,
		IsBzip:       opts.IsBzip,
		QuerySet:     opts.QuerySet,
		Filter:       opts.Filter,
	}

	for rec, err := range jw.IterateBucket(ctx, iter_opts, bucket) {

		cb_err := opts.Callback(ctx, rec, err)

		if cb_err != nil {
			return cb_err
		}
	}

	return nil
}

func WalkSmithsonianRecord(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket, uri string) error {

	if !openaccess.IsMetaDataFile(uri) {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer r.Close()

	iter_opts := &jw.IterateOptions{
		FormatJSON:   opts.FormatJSON,
		ValidateJSON: opts.ValidateJSON,
		QuerySet:     opts.QuerySet,
		IsBzip:       opts.IsBzip,
	}

	for rec, err := range jw.IterateReader(ctx, iter_opts, r) {

		cb_err := opts.Callback(ctx, rec, err)

		if cb_err != nil {
			return cb_err
		}
	}

	return nil
}
