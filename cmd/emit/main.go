package main

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"	
	"io"
	"log"
	"strings"
)

type Record struct {
	Path       string
	LineNumber int
	Body       []byte
}

type Error struct {
	Path       string
	LineNumber int
	Error error
}

func (e *Error) String() string {

	return fmt.Sprintf("[%s] line %d, %v", e.Path, e.LineNumber, e.Error)
}

func main() {

	bucket_uri := flag.String("bucket-uri", "", "A valid GoCloud bucket URI.")
	flag.Parse()

	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer bucket.Close()

	record_ch := make(chan *Record)
	error_ch := make(chan *Error)
	done_ch := make(chan bool)

	var crawlFunc func(context.Context, *blob.Bucket, string) error

	crawlFunc = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}			
		
		iter := bucket.List(&blob.ListOptions{
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

				e := &Error{
					Path: prefix,
					LineNumber: 0,
					Error: err,					
				}

				error_ch <- e
				return nil
			}

			if obj.IsDir {

				err = crawlFunc(ctx, bucket, obj.Key)

				if err != nil {

					e := &Error{
						Path: obj.Key,
						LineNumber: 0,
						Error: err,
					}
					
					error_ch <- e
					return nil
				}

			}

			// parse file of line-demilited records

			// trailing slashes confuse Go Cloud...

			path := strings.TrimRight(obj.Key, "/")
			log.Println(path)

			fh, err := bucket.NewReader(ctx, path, nil)

			if err != nil {
				
				
				e := &Error{
					Path: path,
					LineNumber: 0,
					Error: err,					
				}
				
				error_ch <- e
				return nil
			}

			reader := bufio.NewReader(fh)

			if strings.HasSuffix(path, ".bz2") {
				br := bufio.NewReader(fh)
				cr := bzip2.NewReader(br)
				reader = bufio.NewReader(cr)
			}

			lineno := 0

			for {

				select {
				case <-ctx.Done():
					break
				default:
					// pass
				}

				lineno += 1

				body, err := reader.ReadBytes('\n')

				if err == io.EOF {
					break
				}

				if err != nil {

					e := &Error{
						Path: path,
						LineNumber: lineno,
						Error: err,						
					}
					
					error_ch <- e
					return nil
				}

				var stub interface{}
				err = json.Unmarshal(body, &stub)

				if err != nil {

					e := &Error{
						Path: path,
						LineNumber: lineno,
						Error: err,
					}
					
					error_ch <- e
					return nil
				}

				body, err = json.Marshal(stub)

				if err != nil {

					e := &Error{
						Path: path,
						LineNumber: lineno,
						Error: err,						
					}
					
					error_ch <- e
					return nil
				}

				rec := &Record{
					Path:       path,
					LineNumber: lineno,
					Body:       body,
				}

				record_ch <- rec
				return nil
			}
		}

		return nil
	}

	go func() {

		for {
			select {
			case <-done_ch:
				return
			case err := <- error_ch:
				log.Println(err)
			case rec := <-record_ch:
				log.Println(string(rec.Body))
			default:
				// pass

			}
		}
	}()

	uris := flag.Args()

	for _, uri := range uris {

		err := crawlFunc(ctx, bucket, uri)

		if err != nil {
			log.Fatalf("Failed to crawl %s, %v", uri, err)
		}
	}

	done_ch <- true
}
