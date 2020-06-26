package walk

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"fmt"
	"gocloud.dev/blob"
	"io"
	"strings"
)

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

func Walk(ctx context.Context, bucket *blob.Bucket, uri string, record_ch chan *WalkRecord, error_ch chan *WalkError) error {

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
					continue
				}

			}

			// parse file of line-demilited records

			// trailing slashes confuse Go Cloud...

			path := strings.TrimRight(obj.Key, "/")

			fh, err := bucket.NewReader(ctx, path, nil)

			if err != nil {

				e := &WalkError{
					Path:       path,
					LineNumber: 0,
					Err:        err,
				}

				error_ch <- e
				continue
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

				rec := &WalkRecord{
					Path:       path,
					LineNumber: lineno,
					Body:       body,
				}

				record_ch <- rec
			}
		}

		return nil
	}

	return walkFunc(ctx, bucket, uri)
}
