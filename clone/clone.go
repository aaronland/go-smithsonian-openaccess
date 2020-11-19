package clone

import (
	"context"
	"gocloud.dev/blob"
	"io"
)

type CloneOptions struct {
	URI     string
	Workers int
}

func CloneBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	var walkFunc func(context.Context, *blob.Bucket, string) error

	walkFunc = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		iter := source_bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {

			select {
			case <-ctx.Done():
				break
			default:
				// pass
			}

			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {

				return err
			}

			if obj.IsDir {

				err = walkFunc(ctx, bucket, obj.Key)

				if err != nil {
					return err
				}

				continue
			}

			source_fh, err := source_bucket.NewReader(ctx, obj.Key, nil)

			if err != nil {
				return err
			}

			defer source_fh.Close()

			target_fh, err := target_bucket.NewWriter(ctx, obj.Key, nil)

			if err != nil {
				return err
			}

			_, err = io.Copy(target_fh, source_fh)

			if err != nil {
				return err
			}

			err = target_fh.Close()

			if err != nil {
				return err
			}
		}

		return nil
	}

	walkFunc(ctx, source_bucket, opts.URI)

	return nil
}
