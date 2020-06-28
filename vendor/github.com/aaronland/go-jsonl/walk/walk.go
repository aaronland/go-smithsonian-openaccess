package walk

import (
	"fmt"
	"regexp"
)

const QUERYSET_MODE_ANY string = "ANY"
const QUERYSET_MODE_ALL string = "ALL"

const CONTEXT_PATH string = "github.com/aaronland/go-jsonl#path"

type WalkOptions struct {
	URI           string
	Workers       int
	RecordChannel chan *WalkRecord
	ErrorChannel  chan *WalkError
	ValidateJSON  bool
	FormatJSON    bool
	QuerySet      *WalkQuerySet
	IsBzip        bool
}

type WalkQuerySet struct {
	Queries []*WalkQuery
	Mode    string
}

type WalkQuery struct {
	Path  string
	Match *regexp.Regexp
}

type WalkRecord struct {
	Path       string
	LineNumber int
	Body       []byte
}

type WalkError struct {
	Path       string
	LineNumber int
	Err        error
}

func (e *WalkError) Error() string {
	return e.String()
}

func (e *WalkError) String() string {
	return fmt.Sprintf("[%s] line %d, %v", e.Path, e.LineNumber, e.Err)
}
