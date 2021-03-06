package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

func main() {

	uniq := flag.Bool("unique", true, "Only unique emit placename strings once.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	r := csv.NewReader(os.Stdin)

	seen := new(sync.Map)

	for {
		row, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		pl := row[3]

		if *uniq {

			_, ok := seen.Load(pl)

			if ok {
				continue
			}

			seen.Store(pl, true)
		}

		fmt.Println(pl)
	}

}
