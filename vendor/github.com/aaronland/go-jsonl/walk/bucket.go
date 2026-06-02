package walk

import (
	"context"
	"iter"

	gc_walk "github.com/aaronland/gocloud/blob/walk"
	"gocloud.dev/blob"
)

func IterateBucket(ctx context.Context, iter_opts *IterateOptions, bucket *blob.Bucket) iter.Seq2[*WalkRecord, error] {

	return func(yield func(*WalkRecord, error) bool) {

		walk_cb := func(ctx context.Context, obj *blob.ListObject) error {

			if iter_opts.Filter != nil {

				if !iter_opts.Filter(ctx, obj.Key) {
					return nil
				}
			}

			r, err := bucket.NewReader(ctx, obj.Key, nil)

			if err != nil {
				return err
			}

			defer r.Close()

			for rec, err := range IterateReader(ctx, iter_opts, r) {

				if !yield(rec, err) {
					break
				}
			}

			return nil
		}

		err := gc_walk.WalkBucket(ctx, bucket, walk_cb)

		if err != nil {
			yield(nil, err)
		}
	}
}
