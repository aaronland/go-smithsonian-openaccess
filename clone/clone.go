package clone

import (
	"context"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/mholt/archiver/v3"
	"gocloud.dev/blob"
	"io"
	_ "log"
	"path/filepath"
	"strings"
	"sync"
)

type CloneOptions struct {
	URI      string
	Workers  int
	Force    bool
	Compress bool
}

func CloneBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	v := ctx.Value(openaccess.IS_SMITHSONIAN_S3)

	if v != nil && v.(bool) == true {
		return CloneSmithsonianBucket(ctx, opts, source_bucket, target_bucket)
	}

	throttle := make(chan bool, opts.Workers)

	for i := 0; i < opts.Workers; i++ {
		throttle <- true
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err_ch := make(chan error)
	done_ch := make(chan bool)

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
				return err
			}

			<-throttle

			wg.Add(1)

			go func(obj *blob.ListObject) {

				defer func() {
					throttle <- true
					wg.Done()
				}()

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				if obj.IsDir {

					err = walkFunc(ctx, bucket, obj.Key)

					if err != nil {
						err_ch <- err
					}

					return
				}

				err := cloneObject(ctx, opts, source_bucket, target_bucket, obj.Key)

				if err != nil {
					err_ch <- err
					return
				}

			}(obj)
		}

		wg.Wait()
		return nil
	}

	go func() {

		err := walkFunc(ctx, source_bucket, opts.URI)

		if err != nil {
			err_ch <- err
		}

		done_ch <- true
	}()

	for {
		select {
		case <-done_ch:
			return nil
		case err := <-err_ch:
			return err
		}
	}

	return nil
}

func CloneSmithsonianBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	uri := opts.URI
	base := filepath.Base(uri)

	switch base {
	case "metadata":
		return CloneSmithsonianBucketForAll(ctx, opts, source_bucket, target_bucket)
	case "objects":
		return CloneSmithsonianBucketForAll(ctx, opts, source_bucket, target_bucket)
	default:
		return CloneSmithsonianBucketForUnit(ctx, opts, source_bucket, target_bucket, base)
	}
}

func CloneSmithsonianBucketForAll(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

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

	unit = strings.ToLower(unit)

	remaining := 0

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err_ch := make(chan error)
	done_ch := make(chan bool)

	for _, fname := range openaccess.SMITHSONIAN_DATA_FILES {

		remaining += 1
		<-throttle

		uri := fmt.Sprintf("metadata/edan/%s/%s", unit, fname)

		go func(uri string) {

			defer func() {
				throttle <- true
				done_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			err := cloneObject(ctx, opts, source_bucket, target_bucket, uri)

			if err != nil {
				err_ch <- err
			}

		}(uri)
	}

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		default:
			// pass
		}
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
