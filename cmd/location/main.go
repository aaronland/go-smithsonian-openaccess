package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"github.com/aaronland/go-smithsonian-openaccess"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

func main() {

	to_stdout := flag.Bool("stdout", true, "Emit to STDOUT")
	to_devnull := flag.Bool("null", false, "Emit to /dev/null")

	flag.Parse()

	ctx := context.Background()

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

	reader := bufio.NewReader(os.Stdin)

	wg := new(sync.WaitGroup)
	mu := new(sync.RWMutex)

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		body, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Failed to read bytes, %v", err)
		}

		body = bytes.TrimSpace(body)

		wg.Add(1)

		go func(body []byte) {

			defer func() {
				wg.Done()
			}()

			var openaccess_record *openaccess.OpenAccessRecord

			err = json.Unmarshal(body, &openaccess_record)

			if err != nil {
				log.Println(err)
				return
			}

			// "content.indexStructured.geoLocation",
			// "content.indexStructured.place",
			// "content.freetext.place",

			iim_record := openaccess_record.Content
			place := iim_record.FreeText.Place

			if place != nil {

				mu.Lock()

				for _, pl := range place {

					row := []string{
						openaccess_record.Id,
						"content.freetext.place",
						pl.Label,
						pl.Content,
					}

					err := csv_wr.Write(row)

					if err != nil {
						log.Println(err)
					}
				}

				mu.Unlock()
			}

		}(body)
	}

	wg.Wait()

	csv_wr.Flush()

	err := csv_wr.Error()

	if err != nil {
		log.Fatal(err)
	}
}
