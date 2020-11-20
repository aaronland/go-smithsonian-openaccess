package clone

import (
	"context"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
	"io"
	"log"
	"strings"
	"sync"
)

type CloneOptions struct {
	URI     string
	Workers int
}

func CloneBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	v := ctx.Value(openaccess.IS_SMITHSONIAN_S3)

	if v != nil && v.(bool) == true {
		return CloneSmithsonianBucket(ctx, opts, source_bucket, target_bucket)
	}

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

			// do this concurrently in a go routine

			return cloneObject(ctx, source_bucket, target_bucket, obj.Key)
		}

		return nil
	}

	walkFunc(ctx, source_bucket, opts.URI)

	return nil
}

func CloneSmithsonianBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	for _, unit := range openaccess.SMITHSONIAN_UNITS {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		err := CloneSmithsonianBucketForUnit(ctx, opts, source_bucket, target_bucket, unit)

		if err != nil {
			return err
		}
	}

	return nil
}

func CloneSmithsonianBucketForUnit(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket, unit string) error {

	throttle := make(chan bool, opts.Workers)

	for i := 0; i < opts.Workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	unit = strings.ToLower(unit)

	for _, fname := range openaccess.SMITHSONIAN_DATA_FILES {

		<-throttle
		wg.Add(1)

		uri := fmt.Sprintf("metadata/edan/%s/%s", unit, fname)

		go func(uri string) {

			defer func() {
				wg.Done()
				throttle <- true
			}()

			err := cloneObject(ctx, source_bucket, target_bucket, uri)

			log.Println(err)
		}(uri)
	}

	wg.Wait()
	return nil
}

func cloneObject(ctx context.Context, source_bucket *blob.Bucket, target_bucket *blob.Bucket, uri string) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		//
	}

	target_attrs, err := target_bucket.Attributes(ctx, uri)

	if err == nil {

		source_attrs, err := source_bucket.Attributes(ctx, uri)

		if err != nil {
			return err
		}

		if string(target_attrs.MD5) == string(source_attrs.MD5) {
			return nil
		}
	}

	source_fh, err := source_bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer source_fh.Close()

	target_fh, err := target_bucket.NewWriter(ctx, uri, nil)

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

	return nil
}
