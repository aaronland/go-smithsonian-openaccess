package clone

import (
	"context"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
	"io"
	"log"
	"path/filepath"
	"strings"
)

type CloneOptions struct {
	URI     string
	Workers int
	Force   bool
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

		log.Println("WALK", prefix)
		
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

			log.Println("NEXT", obj, err)
			
			if err == io.EOF {
				log.Println("EOF")
				break
			}

			if err != nil {
				log.Println("ERR")
				return err
			}

			log.Println("OBJ", obj)
			<-throttle

			go func(obj *blob.ListObject) {

				defer func() {
					throttle <- true
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

		log.Println("DONE")
		return nil
	}

	go func() {

		err := walkFunc(ctx, source_bucket, opts.URI)
	
		log.Println("WALK", err)	
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

	log.Println("CLONE", uri)
	
	select {
	case <-ctx.Done():
		return nil
	default:
		//
	}

	if !opts.Force {

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
