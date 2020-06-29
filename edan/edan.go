package edan

// https://edan.si.edu/openaccess/docs/more.html

type OpenAccessRecord struct {
	Id              string          `json:"id"`
	Title           string          `json:title"`
	UnitCode        string          `json:"unitCode"`
	LinkedId        string          `json:"linkedId"`
	Type            string          `json:"type"`
	URL             string          `json:"url"`
	Content         IIMObjectRecord `json:"content"`
	Hash            string          `json:"hash"`
	DocSignature    string          `json:"docSignature"`
	Timestamp       int64           `json:"timestamp"`
	LastTimeUpdated int64           `json:"lastTimeUpdated"`
	Status          int             `json:"status"`
	Version         string          `json:"version"`
	PublicSearch    bool            `json:"publicSearch"`
	Extensions      interface{}     `json:"extensions"`
}

// The primary content in the repository at this time is collection data which we type as edanmdm.
// You can read more about the schema and evolution of this model here.
// https://sirismm.si.edu/siris/EDAN_IMM_OBJECT_RECORDS_1.09.pdf

// Index Metadata ModelFor Objects

type IIMObjectRecord struct {
	DescriptiveNonRepeating IIMDescriptiveNonRepeating `json:"descriptiveNonRepeating"`
	FreeText                IIMFreeText                `json:"freetext"`
	IndexedStructured       IIMIndexedStructured       `json:"indexStructured"`
}

type IIMUsage struct {
	Access string `json:"access"`
}

type IIMMedia struct {
	Content   string             `json:"content"`
	GUID      string             `json:"guid"`
	IDSId     string             `json:"idsId"`
	Thumbnail string             `json:"thumbnail"`
	Usage     IIMUsage           `json:"usage"`
	Resources []IIMMediaResource `json:"resources"`
	Type      string             `json:"type"`
}

type IIMMediaResource struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type IIMOnlineMedia struct {
	Media      []IIMMedia `json:"media"`
	MediaCount int        `json:"mediaCount"`
}

type IIMContentLabel struct {
	Content string `json:"content"`
	Label   string `json:"label"`
}

type IIMDescriptiveNonRepeating struct {
	DataSource    string          `json:"data_source"`
	GUID          string          `json:"guid"`
	MetadataUsage IIMUsage        `json:"metadata_usage"`
	OnlineMedia   IIMOnlineMedia  `json:"online_media"`
	RecordId      string          `json:"record_ID"`
	RecordLink    string          `json:"record_link"`
	Title         IIMContentLabel `json:"title"`
	TitleSort     string          `json:"title_sort"`
	UnitCode      string          `json:"unit_code"`
}

// maybe just this instead?
// type IIMFreeText map[string][]IIMContentLabel

type IIMFreeText struct {
	CreditLine           []IIMContentLabel `json:"creditLine"`
	DataSource           []IIMContentLabel `json:"dataSource"`
	Date                 []IIMContentLabel `json:"date"`
	Identifier           []IIMContentLabel `json:"identifier"`
	Name                 []IIMContentLabel `json:"name"`
	Notes                []IIMContentLabel `json:"notes"`
	ObjectRights         []IIMContentLabel `json:"objectRights"`
	ObjectType           []IIMContentLabel `json:"objectType"`
	PhysicalDescriptions []IIMContentLabel `json:"physicalDescription"`
	Place                []IIMContentLabel `json:"place"`
	SetName              []IIMContentLabel `json:"setName"`
}

type IIMGeoLocationLevel struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}

type IIMGeoLocation struct {
	L2 IIMGeoLocationLevel `json:"L2"`
}

type IIMIndexedStructured struct {
	Date            []string         `json:"date"`
	GeoLocation     []IIMGeoLocation `json:"geoLocation"`
	Name            []string         `json:"name"`
	ObjectType      []string         `json:"object_type"`
	OnlineMediaType []string         `json:"online_media_type"`
	Place           []string         `json:"place"`
}
