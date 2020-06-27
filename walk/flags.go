package walk

import (
	"errors"
	"regexp"
	"strings"
)

const SEP string = "="

type WalkQueryFlags []*WalkQuery

func (m *WalkQueryFlags) String() string {
	return ""
}

func (m *WalkQueryFlags) Set(value string) error {

	parts := strings.Split(value, SEP)

	if len(parts) != 2 {
		return errors.New("Invalid query flag")
	}

	path := parts[0]
	str_match := parts[1]

	re, err := regexp.Compile(str_match)

	if err != nil {
		return err
	}

	q := &WalkQuery{
		Path:  path,
		Match: re,
	}

	*m = append(*m, q)
	return nil
}
