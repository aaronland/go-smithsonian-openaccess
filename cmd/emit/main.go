package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess/walk"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "", "A valid GoCloud bucket file:// URI.")
	workers := flag.Int("workers", 10, "The maximum number of concurrent workers. This is used to prevent filehandle exhaustion.")

	to_stdout := flag.Bool("stdout", true, "Emit to STDOUT")
	to_devnull := flag.Bool("null", false, "Emit to /dev/null")

	as_json := flag.Bool("json", false, "Emit a JSON list.")
	validate_json := flag.Bool("validate-json", false, "Ensure each record is valid JSON.")
	format_json := flag.Bool("format-json", false, "Format JSON output for each record.")

	stats := flag.Bool("stats", false, "Display timings and statistics.")

	var queries walk.WalkQueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_modes := strings.Join([]string{ walk.QUERYSET_MODE_ALL, walk.QUERYSET_MODE_ANY }, ", ")
	desc_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_modes)

	query_mode := flag.String("query-mode", walk.QUERYSET_MODE_ALL, desc_modes)

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

	if *stats {

		t1 := time.Now()

		defer func() {

			final_count := atomic.LoadUint32(&count)
			log.Printf("Processed %d records in %v\n", final_count, time.Since(t1))
		}()
	}

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

		opts := &walk.WalkOptions{
			URI:           uri,
			Workers:       *workers,
			RecordChannel: record_ch,
			ErrorChannel:  error_ch,
			Format:        *format_json,
			Validate:      *validate_json,
		}

		if len(queries) > 0 {

			qs := &walk.WalkQuerySet{
				Queries: queries,
				Mode:    *query_mode,
			}

			opts.QuerySet = qs
		}

		err := walk.Walk(ctx, bucket, opts)

		if err != nil {
			log.Fatalf("Failed to crawl %s, %v", uri, err)
		}
	}

	done_ch <- true

	if *as_json {
		wr.Write([]byte("]"))
	}

}
