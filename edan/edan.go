package edan

// https://edan.si.edu/openaccess/docs/more.html

type OpenAccessRecord struct {
	Id              string      `json:"id"`
	Title           string      `json:title"`
	UnitCode        string      `json:"unitCode"`
	LinkedId        string      `json:"linkedId"`
	Type            string      `json:"type"`
	URL             string      `json:"url"`
	Content         interface{} `json:"content"`
	Hash            string      `json:"hash"`
	DocSignature    string      `json:"docSignature"`
	Timestamp       int64       `json:"timestamp"`
	LastTimeUpdated int64       `json:"lastTimeUpdated"`
	Status          int         `json:"status"`
	Version         string      `json:"version"`
	PublicSearch    bool        `json:"publicSearch"`
	Extensions      interface{} `json:"extensions"`
}

// The primary content in the repository at this time is collection data which we type as edanmdm.
// You can read more about the schema and evolution of this model here.
// https://sirismm.si.edu/siris/EDAN_IMM_OBJECT_RECORDS_1.09.pdf

// Index Metadata ModelFor Objects

type IIMObjectRecord struct {
}
