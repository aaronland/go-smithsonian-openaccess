package oembed

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/tidwall/pretty"
	"log"
	"net/url"
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

	images, err := rec.ImageURLsWithLabel(openaccess.SCREEN_IMAGE)

	if err != nil {
		return nil, err
	}

	if len(images) == 0 {

		online_media, err := rec.OnlineMedia()

		if err != nil {
			return nil, err
		}

		if online_media.MediaCount >= 1 {

			for _, m := range online_media.Media {

				if m.Type != "Images" {
					continue
				}

				if m.Thumbnail == "" {
					continue
				}

				images = append(images, m.Thumbnail)
			}
		}
	}

	if len(images) == 0 {

		body, _ := json.Marshal(rec)
		log.Println(string(pretty.Pretty(body)))

		msg := fmt.Sprintf("OpenAccess record lacks any media objects of type '%s'", openaccess.SCREEN_IMAGE)
		return nil, errors.New(msg)
	}

	records := make([]*OEmbed, 0)

	title := rec.Title
	creditline := rec.CreditLine()

	title = fmt.Sprintf("%s (%s)", title, creditline)

	author_url := rec.Content.DescriptiveNonRepeating.RecordLink
	provider_name := rec.Content.DescriptiveNonRepeating.DataSource

	author_name := fmt.Sprintf("Collection of %s", provider_name)
	provider_url := author_url

	if rec.Content.FreeText.Name != nil {
		author_name = rec.Content.FreeText.Name[0].Content
	} else if rec.Content.FreeText.Manufacturer != nil {
		author_name = rec.Content.FreeText.Manufacturer[0].Content
	} else {
		// pass
	}

	u, err := url.Parse(provider_url)

	if err == nil {
		u.Path = ""
		u.RawQuery = ""
		provider_url = u.String()
	}

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

	return records, nil
}
