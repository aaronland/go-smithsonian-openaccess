package oembed

import (
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

func OEmbedFromOpenAccessRecord(rec *edan.OpenAccessRecord) (*OEmbed, error) {

	o := &OEmbed{
		Version:      "1.0",
		Type:         "photo",
		Height:       -1,
		Width:        -1,
		Title:        rec.Title,
		AuthorName:   rec.Content.IndexedStructured.Name[0],
		AuthorURL:    rec.Content.DescriptiveNonRepeating.RecordLink,
		ProviderName: rec.Content.DescriptiveNonRepeating.DataSource,
		ProviderURL:  rec.Content.DescriptiveNonRepeating.RecordLink,
	}

	return o, nil
}
