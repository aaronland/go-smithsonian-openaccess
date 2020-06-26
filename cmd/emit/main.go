package main

import (
	"context"
	"flag"
	"github.com/aaronland/go-smithsonian-openaccess/walk"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync/atomic"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "", "A valid GoCloud bucket URI.")

	to_stdout := flag.Bool("stdout", true, "Emit to STDOUT")
	to_devnull := flag.Bool("null", false, "Emit to /dev/null")

	as_json := flag.Bool("json", false, "Emit a JSON list")

	flag.Parse()

	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer bucket.Close()

	writers := make([]io.Writer, 0)

	if *to_stdout {
		writers = append(writers, os.Stdout)
	}

	if *to_devnull {
		writers = append(writers, ioutil.Discard)
	}

	if len(writers) == 0 {
		log.Fatal("Nothing to write to.")
	}

	wr := io.MultiWriter(writers...)

	if *as_json {
		wr.Write([]byte("["))
	}

	count := uint32(0)

	record_ch := make(chan *walk.WalkRecord)
	error_ch := make(chan *walk.WalkError)
	done_ch := make(chan bool)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {

		for {
			select {
			case <-done_ch:
				return
			case err := <-error_ch:
				log.Println(err)
			case rec := <-record_ch:

				new_count := atomic.AddUint32(&count, 1)

				if *as_json && new_count > 1 {
					wr.Write([]byte(","))
				}

				wr.Write(rec.Body)

			default:
				// pass

			}
		}
	}()

	uris := flag.Args()

	for _, uri := range uris {

		err := walk.Walk(ctx, bucket, uri, record_ch, error_ch)

		if err != nil {
			log.Fatalf("Failed to crawl %s, %v", uri, err)
		}
	}

	done_ch <- true

	if *as_json {
		wr.Write([]byte("]"))
	}

}
