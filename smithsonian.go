package openaccess

import (
	"path/filepath"
	"regexp"
)

var re_datafile *regexp.Regexp

func init() {
	re_datafile = regexp.MustCompile(`[a-f0-9]{2}\.txt`)
}

func IsMetaDataFile(path string) bool {
	fname := filepath.Base(path)
	return re_datafile.MatchString(fname)
}
