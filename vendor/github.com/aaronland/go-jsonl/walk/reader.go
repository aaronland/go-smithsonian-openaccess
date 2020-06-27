package walk

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"io"
)

func WalkReader(ctx context.Context, opts *WalkOptions, fh io.Reader) {

	record_ch := opts.RecordChannel
	error_ch := opts.ErrorChannel

	reader := bufio.NewReader(fh)

	if opts.IsBzip {
		br := bufio.NewReader(fh)
		cr := bzip2.NewReader(br)
		reader = bufio.NewReader(cr)
	}

	path := ""	
	lineno := 0

	v := ctx.Value(CONTEXT_PATH)

	if v != nil {
		path = v.(string)
	}
	
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

		if opts.QuerySet != nil {

			queries := opts.QuerySet.Queries
			mode := opts.QuerySet.Mode

			tests := len(queries)
			matches := 0

			for _, q := range queries {

				rsp := gjson.GetBytes(body, q.Path)

				if !rsp.Exists() {

					if mode == QUERYSET_MODE_ALL {
						break
					}
				}

				for _, r := range rsp.Array() {

					has_match := true

					if !q.Match.MatchString(r.String()) {

						has_match = false

						if mode == QUERYSET_MODE_ALL {
							break
						}
					}

					if !has_match {

						if mode == QUERYSET_MODE_ALL {
							break
						}

						continue
					}

					matches += 1
				}
			}

			if mode == QUERYSET_MODE_ALL {

				if matches < tests {
					continue
				}
			}

			if matches == 0 {
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

}
