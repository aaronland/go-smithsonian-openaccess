package oembed

import (
	"errors"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
)

type OEmbed struct {
	Version      string `json:"version,xml:"version""`
	Type         string `json:"type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
}

func OEmbedRecordsFromOpenAccessRecord(rec *openaccess.OpenAccessRecord) ([]*OEmbed, error) {

	records := make([]*OEmbed, 0)

	images, err := rec.ImageURLsWithLabel(openaccess.SCREEN_IMAGE)

	if err != nil {
		return nil, err
	}

	title := rec.Title
	creditline := rec.CreditLine()

	title = fmt.Sprintf("%s. %s", title, creditline)

	author_name := rec.Content.IndexedStructured.Name[0]
	author_url := rec.Content.DescriptiveNonRepeating.RecordLink

	provider_name := rec.Content.DescriptiveNonRepeating.DataSource
	provider_url := author_url

	for _, url := range images {

		o := &OEmbed{
			Version:      "1.0",
			Type:         "photo",
			Height:       -1,
			Width:        -1,
			URL:          url,
			Title:        title,
			AuthorName:   author_name,
			AuthorURL:    author_url,
			ProviderName: provider_name,
			ProviderURL:  provider_url,
		}

		records = append(records, o)
	}

	if len(records) == 0 {
		return nil, errors.New("Unable to find any suitable OEmbed records")
	}

	return records, nil
}
