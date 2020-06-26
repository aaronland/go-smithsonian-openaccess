package walk

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tidwall/pretty"
	"gocloud.dev/blob"
	"io"
	_ "log"
	"strings"
	"sync"
)

type WalkOptions struct {
	URI           string
	Workers       int
	RecordChannel chan *WalkRecord
	ErrorChannel  chan *WalkError
	Validate      bool
	Format        bool
}

type WalkRecord struct {
	Path       string
	LineNumber int
	Body       []byte
}

type WalkError struct {
	Path       string
	LineNumber int
	Err        error
}

func (e *WalkError) Error() string {
	return e.String()
}

func (e *WalkError) String() string {

	return fmt.Sprintf("[%s] line %d, %v", e.Path, e.LineNumber, e.Err)
}

func Walk(ctx context.Context, bucket *blob.Bucket, opts *WalkOptions) error {

	record_ch := opts.RecordChannel
	error_ch := opts.ErrorChannel

	workers := opts.Workers

	throttle := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		throttle <- true
	}

	wg := new(sync.WaitGroup)

	var walkFunc func(context.Context, *blob.Bucket, string) error

	walkFunc = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

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

				e := &WalkError{
					Path:       prefix,
					LineNumber: 0,
					Err:        err,
				}

				error_ch <- e
				return nil
			}

			if obj.IsDir {

				err = walkFunc(ctx, bucket, obj.Key)

				if err != nil {

					e := &WalkError{
						Path:       obj.Key,
						LineNumber: 0,
						Err:        err,
					}

					error_ch <- e
				}

				continue
			}

			// parse file of line-demilited records

			// trailing slashes confuse Go Cloud...

			path := strings.TrimRight(obj.Key, "/")

			go func(path string) {

				// log.Println("WAIT", path)
				<-throttle

				wg.Add(1)

				defer func() {
					// log.Println("CLOSE", path)
					wg.Done()
					throttle <- true
				}()

				// log.Println("OPEN", path)
				fh, err := bucket.NewReader(ctx, path, nil)

				if err != nil {

					e := &WalkError{
						Path:       path,
						LineNumber: 0,
						Err:        err,
					}

					error_ch <- e
					return
				}

				defer fh.Close()

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

						e := &WalkError{
							Path:       path,
							LineNumber: lineno,
							Err:        err,
						}

						error_ch <- e
						continue
					}

					if opts.Validate {
						var stub interface{}
						err = json.Unmarshal(body, &stub)

						if err != nil {

							e := &WalkError{
								Path:       path,
								LineNumber: lineno,
								Err:        err,
							}

							error_ch <- e
							continue
						}

						body, err = json.Marshal(stub)

						if err != nil {

							e := &WalkError{
								Path:       path,
								LineNumber: lineno,
								Err:        err,
							}

							error_ch <- e
							continue
						}
					}

					if opts.Format {
						body = pretty.Pretty(body)
					}

					rec := &WalkRecord{
						Path:       path,
						LineNumber: lineno,
						Body:       body,
					}

					record_ch <- rec
				}
			}(path)

		}

		return nil
	}

	walkFunc(ctx, bucket, opts.URI)
	wg.Wait()

	return nil
}
