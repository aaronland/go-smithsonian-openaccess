package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	jw "github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/aaronland/go-smithsonian-openaccess/walk"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "", "A valid GoCloud bucket file:// URI.")
	workers := flag.Int("workers", 10, "The maximum number of concurrent workers. This is used to prevent filehandle exhaustion.")

	to_stdout := flag.Bool("stdout", true, "Emit to STDOUT")
	to_devnull := flag.Bool("null", false, "Emit to /dev/null")

	stats := flag.Bool("stats", false, "Display timings and statistics.")

	var queries jw.WalkQueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_modes := strings.Join([]string{jw.QUERYSET_MODE_ALL, jw.QUERYSET_MODE_ANY}, ", ")
	desc_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_modes)

	query_mode := flag.String("query-mode", jw.QUERYSET_MODE_ALL, desc_modes)

	include_guid := flag.Bool("include-guid", false, "Include the OpenAccess `content.descriptiveNonRepeating.guid` identifier")
	include_record_id := flag.Bool("include-record-id", true, "Include the OpenAccess `content.descriptiveNonRepeating.record_ID` identifier")
	include_record_link := flag.Bool("include-record-link", false, "Include the OpenAccess `content.descriptiveNonRepeating.record_link` identifier")
	include_openaccess_id := flag.Bool("include-openaccess-id", false, "Include the OpenAccess `id` identifier")
	include_all := flag.Bool("include-all", false, "Include all OpenAccess identifiers")

	csv_header := flag.Bool("csv-header", true, "Include a CSV header row in the output")

	flag.Parse()

	if *include_all {
		*include_guid = true
		*include_record_id = true
		*include_record_link = true
		*include_openaccess_id = true
	}

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
	csv_wr := csv.NewWriter(wr)

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

	mu := new(sync.RWMutex)

	write := func(ctx context.Context, rec *jw.WalkRecord, ids ...string) error {

		mu.Lock()
		defer mu.Unlock()

		for _, id := range ids {

			select {
			case <-ctx.Done():
				return nil
			default:
				// pass
			}

			row := []string{
				id,
				rec.Path,
				strconv.Itoa(rec.LineNumber),
			}

			new_count := atomic.AddUint32(&count, 1)

			if new_count == 1 && *csv_header {

				header_row := []string{
					"id",
					"path",
					"line_number",
				}

				err = csv_wr.Write(header_row)

				if err != nil {
					return err
				}

			}

			err = csv_wr.Write(row)

			if err != nil {
				return err
			}
		}

		return nil
	}

	cb := func(ctx context.Context, rec *jw.WalkRecord, err error) error {

		if err != nil {
			log.Println(err)
			return err
		}

		var object *openaccess.OpenAccessRecord

		err = json.Unmarshal(rec.Body, &object)

		if err != nil {
			log.Println(err)
			return err
		}

		ids := make([]string, 0)

		if *include_guid {

			guid := object.Content.DescriptiveNonRepeating.GUID

			if guid != "" {
				ids = append(ids, guid)
			} else {
				log.Printf("Object record %s missing `content.descriptiveNonRepeating.guid` identifier, %s (%d)", object.Id, rec.Path, rec.LineNumber)
			}
		}

		if *include_record_id {

			record_id := object.Content.DescriptiveNonRepeating.RecordId

			if record_id != "" {
				ids = append(ids, record_id)
			} else {
				log.Printf("Object record %s missing `content.descriptiveNonRepeating.record_ID` identifier, %s (%d)", object.Id, rec.Path, rec.LineNumber)
			}
		}

		if *include_record_link {

			record_link := object.Content.DescriptiveNonRepeating.RecordLink

			if record_link != "" {
				ids = append(ids, record_link)
			} else {
				log.Printf("Object record %s missing `content.descriptiveNonRepeating.record_link` identifier, %s (%d)", object.Id, rec.Path, rec.LineNumber)
			}

		}

		// ids = append(ids, object.Content.FreeText.Identifier[0].Content)	// Inventory Number

		if *include_openaccess_id {

			openaccess_id := object.Id

			if openaccess_id != "" {
				ids = append(ids, openaccess_id)
			} else {
				log.Printf("Object record missing `id` identifier, %s (%d)", rec.Path, rec.LineNumber)
			}
		}

		return write(ctx, rec, ids...)
	}

	uris := flag.Args()

	for _, uri := range uris {

		opts := &walk.WalkOptions{
			URI:          uri,
			Workers:      *workers,
			FormatJSON:   false,
			ValidateJSON: false,
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

	csv_wr.Flush()

	err = csv_wr.Error()

	if err != nil {
		log.Fatal(err)
	}
}
