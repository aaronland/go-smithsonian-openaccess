package clone

import (
	"context"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/mholt/archiver/v3"
	"gocloud.dev/blob"
	"io"
	"log"
	"sync"
)

type CloneOptions struct {
	URI      string
	Workers  int
	Force    bool
	Compress bool
}

func CloneBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	throttle := make(chan bool, opts.Workers)

	for i := 0; i < opts.Workers; i++ {
		throttle <- true
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

		wg := new(sync.WaitGroup)

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
				return fmt.Errorf("Failed to iterate next, %w", err)
			}

			if obj.IsDir {

				err = walkFunc(ctx, bucket, obj.Key)

				if err != nil {
					return fmt.Errorf("Failed to walk '%s', %w", obj.Key, err)
				}

				continue
			}

			<-throttle

			wg.Add(1)

			go func(uri string) {

				defer func() {
					wg.Done()
					throttle <- true
				}()

				err = cloneObject(ctx, opts, source_bucket, target_bucket, uri)

				if err != nil {
					log.Printf("Failed to clone '%s', %w", uri, err)
				}

			}(obj.Key)
		}

		wg.Wait()
		return nil
	}

	err := walkFunc(ctx, source_bucket, opts.URI)

	if err != nil {
		return fmt.Errorf("Failed to clone bucket, %w", err)
	}

	return nil
}

func cloneObject(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket, uri string) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		//
	}

	compare_md5 := true

	if opts.Force {
		compare_md5 = false
	}

	v := ctx.Value(openaccess.IS_SMITHSONIAN_S3)

	if v != nil && v.(bool) == true {

		// OpenAccess files in Smithsonian S3 bucket are uncompressed so
		// if we are compressing on the receiving side there is no way to
		// compare source and target files.

		if compare_md5 && opts.Compress {
			compare_md5 = false
		}
	}

	if compare_md5 {

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
	}

	source_fh, err := source_bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer source_fh.Close()

	target_uri := uri

	if opts.Compress {
		target_uri = fmt.Sprintf("%s.bz2", uri)
	}

	target_fh, err := target_bucket.NewWriter(ctx, target_uri, nil)

	if err != nil {
		return err
	}

	if opts.Compress {
		arch := archiver.NewBz2()
		err = arch.Compress(source_fh, target_fh)
	} else {
		_, err = io.Copy(target_fh, source_fh)
	}

	if err != nil {
		return err
	}

	err = target_fh.Close()

	if err != nil {
		return err
	}

	return nil
}
