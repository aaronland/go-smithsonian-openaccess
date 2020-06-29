package openaccess

import (
	"github.com/aaronland/go-smithsonian-openaccess/edan"
	"strings"
)

// https://edan.si.edu/openaccess/docs/more.html

type OpenAccessRecord struct {
	Id              string               `json:"id"`
	Title           string               `json:title"`
	UnitCode        string               `json:"unitCode"`
	LinkedId        string               `json:"linkedId"`
	Type            string               `json:"type"`
	URL             string               `json:"url"`
	Content         edan.IIMObjectRecord `json:"content"`
	Hash            string               `json:"hash"`
	DocSignature    string               `json:"docSignature"`
	Timestamp       int64                `json:"timestamp"`
	LastTimeUpdated int64                `json:"lastTimeUpdated"`
	Status          int                  `json:"status"`
	Version         string               `json:"version"`
	PublicSearch    bool                 `json:"publicSearch"`
	Extensions      interface{}          `json:"extensions"`
}

func (rec *OpenAccessRecord) OnlineMedia() (edan.IIMOnlineMedia, error) {
	return rec.Content.DescriptiveNonRepeating.OnlineMedia, nil
}

func (rec *OpenAccessRecord) ImageURLsWithLabel(label string) ([]string, error) {

	online_media, err := rec.OnlineMedia()

	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)

	for _, m := range online_media.Media {

		for _, r := range m.Resources {

			if r.Label == label {
				urls = append(urls, r.URL)
			}
		}
	}

	return urls, nil
}

func (rec *OpenAccessRecord) CreditLine() string {

	entries := rec.Content.FreeText.CreditLine
	phrases := make([]string, 0)

	for _, e := range entries {
		phrases = append(phrases, e.Content)
	}

	return strings.Join(phrases, " ")
}
