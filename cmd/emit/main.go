package main

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"flag"
	"github.com/tidwall/pretty"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	flag.Parse()

	ctx := context.Background()

	uris := flag.Args()

	writers := []io.Writer{
		os.Stdout,
		// ioutil.Discard,
	}

	wr := io.MultiWriter(writers...)

	for _, uri := range uris {

		fh, err := os.Open(uri)

		if err != nil {
			log.Fatal(err)
		}

		reader := bufio.NewReader(fh)

		if strings.HasSuffix(uri, ".bz2") {
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

			ln, err := reader.ReadBytes('\n')

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}

			var stub interface{}
			err = json.Unmarshal(ln, &stub)

			if err != nil {
				log.Printf("Failed to parse JSON in %s at line %d, %v", uri, lineno, err)
				continue
			}

			enc, err := json.Marshal(stub)

			if err != nil {
				log.Fatal(err)
			}

			enc = pretty.Pretty(enc)

			_, err = wr.Write(enc)

			if err != nil {
				log.Fatal(err)
			}

		}
	}
}
