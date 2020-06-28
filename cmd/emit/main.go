package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess/edan"
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

	var queries jw.WalkQueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_modes := strings.Join([]string{jw.QUERYSET_MODE_ALL, jw.QUERYSET_MODE_ANY}, ", ")
	desc_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_modes)

	query_mode := flag.String("query-mode", jw.QUERYSET_MODE_ALL, desc_modes)

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

	count := uint32(0)

	if *stats {

		t1 := time.Now()

		defer func() {

			final_count := atomic.LoadUint32(&count)
			log.Printf("Processed %d records in %v\n", final_count, time.Since(t1))
		}()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cb := func(ctx context.Context, rec *jw.WalkRecord, err error) error {

		if err != nil {
			log.Println(err)
			return err
		}

		var object *edan.OpenAccessRecord

		err = json.Unmarshal(rec.Body, &object)

		if err != nil {
			log.Println(err)
			return err
		}

		new_count := atomic.AddUint32(&count, 1)

		if *as_json && new_count > 1 {
			wr.Write([]byte(","))
		}

		wr.Write(rec.Body)
		return nil
	}

	uris := flag.Args()

	if *as_json {
		wr.Write([]byte("["))
	}

	for _, uri := range uris {

		opts := &walk.WalkOptions{
			URI:          uri,
			Workers:      *workers,
			FormatJSON:   *format_json,
			ValidateJSON: *validate_json,
			Callback:     cb,
		}

		if len(queries) > 0 {

			qs := &jw.WalkQuerySet{
				Queries: queries,
				Mode:    *query_mode,
			}

			opts.QuerySet = qs
		}

		err := walk.WalkBucket(ctx, opts, bucket)

		if err != nil {
			log.Fatalf("Failed to crawl %s, %v", uri, err)
		}
	}

	if *as_json {
		wr.Write([]byte("]"))
	}

}
