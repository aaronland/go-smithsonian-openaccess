package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/aaronland/go-smithsonian-openaccess/clone"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"
	"log"
	"os"
)

func main() {

	source_bucket_uri := flag.String("source-bucket-uri", "si://", "A valid GoCloud bucket URI. Valid schemes are: file://, s3:// and si:// which is signals that data should be retrieved from the Smithsonian's 'smithsonian-open-access' S3 bucket.")

	target_bucket_uri := flag.String("target-bucket-uri", "", "A valid GoCloud bucket URI. Valid schemes are: file://, s3://.")

	workers := flag.Int("workers", 10, "The maximum number of concurrent workers. This is used to prevent filehandle exhaustion.")
	force := flag.Bool("force", false, "Clone files even if they are present in target bucket and MD5 hashes between source and target buckets match.")
	compress := flag.Bool("compress", false, "Compress files in the target bucket using bzip2 encoding. Files will be appended with a '.bz2' suffix.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options] path(N) path(n)\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	ctx := context.Background()

	ctx, source_bucket, err := openaccess.OpenBucket(ctx, *source_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer source_bucket.Close()

	target_bucket, err := blob.OpenBucket(ctx, *target_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer target_bucket.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	uris := flag.Args()

	for _, uri := range uris {

		opts := &clone.CloneOptions{
			URI:      uri,
			Workers:  *workers,
			Force:    *force,
			Compress: *compress,
		}

		err := clone.CloneBucket(ctx, opts, source_bucket, target_bucket)

		if err != nil {
			log.Fatalf("Failed to clone %s, %v", uri, err)
		}
	}

}
