package oembed

import (
	"errors"
	"github.com/aaronland/go-smithsonian-openaccess/edan"
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

func OEmbedRecordsFromOpenAccessRecord(rec *edan.OpenAccessRecord) ([]*OEmbed, error) {

	media := rec.Content.DescriptiveNonRepeating.OnlineMedia.Media

	records := make([]*OEmbed, 0)

	for _, m := range media {

		var url string

		// https://github.com/Smithsonian/OpenAccess/issues/2
		var width int
		var height int

		for _, r := range m.Resources {

			if r.Label == "Screen Image" {
				url = r.URL
				break
			}
		}

		if url == "" {
			continue
		}

		// TO DO : GET CREDITLINE

		o := &OEmbed{
			Version:      "1.0",
			Type:         "photo",
			Height:       height,
			Width:        width,
			URL:          url,
			Title:        rec.Title,
			AuthorName:   rec.Content.IndexedStructured.Name[0],
			AuthorURL:    rec.Content.DescriptiveNonRepeating.RecordLink,
			ProviderName: rec.Content.DescriptiveNonRepeating.DataSource,
			ProviderURL:  rec.Content.DescriptiveNonRepeating.RecordLink,
		}

		records = append(records, o)
	}

	if len(records) == 0 {
		return nil, errors.New("Unable to find any suitable OEmbed records")
	}

	return records, nil
}
