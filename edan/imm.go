package edan

// The primary content in the repository at this time is collection data which we type as edanmdm.
// You can read more about the schema and evolution of this model here.
// https://sirismm.si.edu/siris/EDAN_IMM_OBJECT_RECORDS_1.09.pdf

// Index Metadata ModelFor Objects

type IIMObjectRecord struct {
	DescriptiveNonRepeating IIMDescriptiveNonRepeating `json:"descriptiveNonRepeating,omitempty"`
	FreeText                IIMFreeText                `json:"freetext,omitempty"`
	IndexedStructured       IIMIndexedStructured       `json:"indexStructured,omitempty"`
}

type IIMUsage struct {
	Access string `json:"access,omitempty"`
}

type IIMMedia struct {
	Content   string             `json:"content,omitempty"`
	GUID      string             `json:"guid,omitempty"`
	IDSId     string             `json:"idsId,omitempty"`
	Thumbnail string             `json:"thumbnail,omitempty"`
	Usage     IIMUsage           `json:"usage,omitempty"`
	Resources []IIMMediaResource `json:"resources,omitempty"`
	Type      string             `json:"type,omitempty"`
}

type IIMMediaResource struct {
	Label string `json:"label,omitempty"`
	URL   string `json:"url,omitempty"`
}

type IIMOnlineMedia struct {
	Media      []IIMMedia `json:"media,omitempty"`
	MediaCount int        `json:"mediaCount,omitempty"`
}

type IIMContentLabel struct {
	Content string `json:"content,omitempty"`
	Label   string `json:"label,omitempty"`
}

type IIMDescriptiveNonRepeating struct {
	DataSource    string          `json:"data_source,omitempty"`
	GUID          string          `json:"guid,omitempty"`
	MetadataUsage IIMUsage        `json:"metadata_usage,omitempty"`
	OnlineMedia   IIMOnlineMedia  `json:"online_media,omitempty"`
	RecordId      string          `json:"record_ID,omitempty"`
	RecordLink    string          `json:"record_link,omitempty"`
	Title         IIMContentLabel `json:"title,omitempty"`
	TitleSort     string          `json:"title_sort,omitempty"`
	UnitCode      string          `json:"unit_code,omitempty"`
}

// maybe just this instead?
// type IIMFreeText map[string][]IIMContentLabel

type IIMFreeText struct {
	CreditLine           []IIMContentLabel `json:"creditLine,omitempty"`
	DataSource           []IIMContentLabel `json:"dataSource,omitempty"`
	Date                 []IIMContentLabel `json:"date,omitempty"`
	Identifier           []IIMContentLabel `json:"identifier,omitempty"`
	Manufacturer         []IIMContentLabel `json:"manufacturer,omitempty"`
	Name                 []IIMContentLabel `json:"name,omitempty"`
	Notes                []IIMContentLabel `json:"notes,omitempty"`
	ObjectRights         []IIMContentLabel `json:"objectRights,omitempty"`
	ObjectType           []IIMContentLabel `json:"objectType,omitempty"`
	PhysicalDescriptions []IIMContentLabel `json:"physicalDescription,omitempty"`
	Place                []IIMContentLabel `json:"place,omitempty"`
	SetName              []IIMContentLabel `json:"setName,omitempty"`
}

type IIMGeoLocationLevel struct {
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
}

type IIMGeoLocation struct {
	L2 IIMGeoLocationLevel `json:"L2,omitempty"`
}

type IIMIndexedStructured struct {
	Date            []string         `json:"date,omitempty"`
	GeoLocation     []IIMGeoLocation `json:"geoLocation,omitempty"`
	Name            []string         `json:"name,omitempty"`
	ObjectType      []string         `json:"object_type,omitempty"`
	OnlineMediaType []string         `json:"online_media_type,omitempty"`
	Place           []string         `json:"place,omitempty"`
	UsageFlag       []string         `json:"usage_flag,omitempty"`
}
