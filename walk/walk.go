package walk

import (
	"context"
	jw "github.com/aaronland/go-jsonl/walk"
	"gocloud.dev/blob"
	"log"
	"sync"
)

type WalkOptions struct {
	URI           string
	Workers       int
	RecordChannel chan *jw.WalkRecord
	ErrorChannel  chan *jw.WalkError
	Validate      bool
	Format        bool
	QuerySet      *jw.WalkQuerySet
	Callback      WalkRecordCallbackFunc
}

type WalkRecordCallbackFunc func(context.Context, *jw.WalkRecord) error

func WalkBucket(ctx context.Context, opts *WalkOptions, bucket *blob.Bucket) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := new(sync.WaitGroup)
	cb := opts.Callback

	jw_record_ch := make(chan *jw.WalkRecord)
	jw_error_ch := make(chan *jw.WalkError)

	jw_opts := &jw.WalkOptions{
		URI:           opts.URI,
		Workers:       opts.Workers,
		RecordChannel: jw_record_ch,
		ErrorChannel:  jw_error_ch,
		Format:        opts.Format,
		Validate:      opts.Validate,
		QuerySet:      opts.QuerySet,
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-jw_error_ch:
				log.Println(err)
			case rec := <-jw_record_ch:

				wg.Add(1)

				go func() {

					defer wg.Done()

					err := cb(ctx, rec)

					if err != nil {
						jw_error_ch <- &jw.WalkError{
							Path:       rec.Path,
							LineNumber: rec.LineNumber,
							Err:        err,
						}
					}
				}()

			default:
				// pass
			}
		}
	}()

	err := jw.WalkBucket(ctx, jw_opts, bucket)

	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
