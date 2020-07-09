package oembed

import (
	"errors"
	"fmt"
	"github.com/aaronland/go-smithsonian-openaccess"
	"github.com/jtacoma/uritemplates"
	"net/url"
	"strings"
)

const OBJECT_URI_TEMPLATE string = "si://{collection}/o/{objectid}"

var object_uri_template *uritemplates.UriTemplate

func init() {

	t, err := uritemplates.Parse(OBJECT_URI_TEMPLATE)

	if err != nil {
		panic(err)
	}

	object_uri_template = t
}

type OEmbedRecord struct {
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
	ObjectURI    string `json:"object_uri"`
}

func OEmbedRecordsFromOpenAccessRecord(rec *openaccess.OpenAccessRecord) ([]*OEmbedRecord, error) {

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

		// this should go in a function... somewhere (20200630/thisisaaronland)
		// body, _ := json.Marshal(rec)
		// log.Println(string(pretty.Pretty(body)))

		msg := fmt.Sprintf("OpenAccess record lacks any media objects of type '%s' or 'Images'", openaccess.SCREEN_IMAGE)
		return nil, errors.New(msg)
	}

	records := make([]*OEmbedRecord, 0)

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

	// object_uri and author_url
	// ...please write me

	// https://nmaahc.si.edu/object/nmaahc_2010.39.8
	// si://nmaahc/o/2011_155_299ab

	// https://airandspace.si.edu/collection/id/nasm_A20060281000
	// si://nasm/o/A19820380000

	// http://collection.cooperhewitt.org/view/objects/asitem/id/81405
	// si://chndm/o/1972-42-130-a_b

	/*

	{
	  "version": "1.0",
	  "type": "photo",
	  "width": -1,
	  "height": -1,
	  "title": "License, Aviator's, Brevetto Superiore (Donated by the Family of George Harold Cronin)",
	  "url": "http://ids.si.edu/ids/deliveryService?id=NASM-A19940181000_PS02",
	  "author_name": "Collection of National Air and Space Museum",
	  "author_url": "https://airandspace.si.edu/collection/id/nasm_A19940181000",
	  "provider_name": "National Air and Space Museum",
	  "provider_url": "https://airandspace.si.edu",
	  "object_uri": "si://nasm/o/A19940181000"
	}

	*/

	unit := rec.UnitCode
	unit = strings.ToLower(unit)

	objectid_prefix := fmt.Sprintf("edanmdm-%s_", unit)

	objectid := rec.Id
	objectid = strings.Replace(objectid, objectid_prefix, "", 1)
	objectid = strings.Replace(objectid, ".", "_", -1)

	values := make(map[string]interface{})
	values["collection"] = unit
	values["objectid"] = objectid

	object_uri, err := object_uri_template.Expand(values)

	if err != nil {
		return nil, err
	}

	u, err := url.Parse(provider_url)

	if err == nil {
		u.Path = ""
		u.RawQuery = ""
		provider_url = u.String()
	}

	for _, url := range images {

		o := &OEmbedRecord{
			Version:      "1.0",
			Type:         "photo",
			Height:       -1, // https://github.com/Smithsonian/OpenAccess/issues/2
			Width:        -1, // see above
			URL:          url,
			Title:        title,
			AuthorName:   author_name,
			AuthorURL:    author_url,
			ProviderName: provider_name,
			ProviderURL:  provider_url,
			ObjectURI:    object_uri,
		}

		records = append(records, o)
	}

	return records, nil
}
