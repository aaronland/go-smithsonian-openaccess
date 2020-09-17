package oembed

import (
	"context"
)

type Photo struct {
	Version          string `json:"version,xml:"version""`
	Type             string `json:"type",xml:"type"`
	Width            int    `json:"width",xml:"width"`
	Height           int    `json:"height",xml:"height"`
	Title            string `json:"title",xml:"title"`
	URL              string `json:"url",xml:"url"`
	AuthorName       string `json:"author_name",xml:"author_name"`
	AuthorURL        string `json:"author_url",xml:"author_url"`
	ProviderName     string `json:"provider_name",xml:"provider_name"`
	ProviderURL      string `json:"provider_url",xml:"provider_url"`
	ObjectURI        string `json:"object_uri",xml:"object_uri"`
	ThumbnailURL     string `json:"thumbnail_url,omitempty",xml:"thumbnail_url,omitempty"`
	ThumbnailWidth   int    `json:"thumbnail_width,omitempty",xml:"thumbnail_width,omitempty"`
	ThumbnailHeight  int    `json:"thumbnail_height,omitempty",xml:"thumbnail_height,omitempty"`
	DataURL          string `json:"data_url,omitempty",xml:"data_url,omitempty"`
	ThumbnailDataURL string `json:"thumbnail_data_url,omitempty",xml:"thumbnail_data_url,omitempty"`
}

// TBD - how to handle (eventually) things that aren't "photos"
// (20200713/thisisaaronland)

type OEmbedDatabaseCallback func(context.Context, *Photo) error

type OEmbedDatabase interface {
	AddOEmbed(context.Context, *Photo) error
	GetRandomOEmbed(context.Context) (*Photo, error)
	GetOEmbedWithURL(context.Context, string) (*Photo, error)
	GetOEmbedWithObjectURI(context.Context, string) ([]*Photo, error)
	GetOEmbedWithCallback(context.Context, OEmbedDatabaseCallback) error
	Close() error
}
