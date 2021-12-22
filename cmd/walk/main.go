package main

import (
	"context"
	"flag"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"
	"io"
	"log"
	"path/filepath"
	"regexp"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "si://", "...")
	flag.Parse()

	ctx := context.Background()

	ctx, bucket, err := openaccess.OpenMetadataBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer bucket.Close()

	re_datafile, err := regexp.Compile(`[a-f0-9]{2}\.txt`)

	if err != nil {
		log.Fatalf("Failed to compile data file regular expression, %v", err)
	}

	var list func(context.Context, *blob.Bucket, string) error

	list = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		iter := bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {
			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			path := obj.Key

			if obj.IsDir {

				err := list(ctx, bucket, path)

				if err != nil {
					return err
				}

				continue
			}

			fname := filepath.Base(path)

			if !re_datafile.MatchString(fname) {
				continue
			}

			log.Println("WALK", path)
		}

		return nil
	}

	err = list(ctx, bucket, "")

	if err != nil {
		log.Fatalf("Failed to list '%s', %v", *bucket_uri, err)
	}

}
