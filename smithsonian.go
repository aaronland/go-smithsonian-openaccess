package openaccess

import (
	"fmt"
)

var SMITHSONIAN_UNITS []string
var SMITHSONIAN_DATA_FILES []string

func init() {

	// https://github.com/Smithsonian/OpenAccess/tree/master/metadata/objects
	// It would be better if we didn't have to hard code this but today we do...
	// (20201118/straup)

	// see also: https://github.com/aaronland/go-smithsonian-openaccess/issues/7
	// (20201119/straup)

	SMITHSONIAN_UNITS = []string{
		"ACAH",
		"ACM",
		"CFCHFOLKLIFE",
		"CHNDM",
		"FBR",
		"FSA",
		"FSG",
		"HAC",
		"HMSG",
		"HSFA",
		"NAA",
		"NASM",
		"NMAAHC",
		"NMAH",
		"NMAI",
		"NMAfA", // note the mixed-case
		"NMNHANTHRO",
		"NMNHBIRDS",
		"NMNHBOTANY",
		"NMNHEDUCATION",
		"NMNHENTO",
		"NMNHFISHES",
		"NMNHHERPS",
		"NMNHINV",
		"NMNHMAMMALS",
		"NMNHMINSCI",
		"NMNHPALEO",
		"NPG",
		"NPM",
		"SAAM",
		"SI",
		"SIA",
		"SIL",
	}

	// https://github.com/Smithsonian/OpenAccess/issues/7#issuecomment-696833714

	digits := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	letters := []string{"a", "b", "c", "d", "e", "f"}

	for _, first := range digits {

		for _, second := range digits {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			SMITHSONIAN_DATA_FILES = append(SMITHSONIAN_DATA_FILES, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			SMITHSONIAN_DATA_FILES = append(SMITHSONIAN_DATA_FILES, fname)
		}
	}

	for _, first := range letters {

		for _, second := range digits {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			SMITHSONIAN_DATA_FILES = append(SMITHSONIAN_DATA_FILES, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			SMITHSONIAN_DATA_FILES = append(SMITHSONIAN_DATA_FILES, fname)
		}
	}

}
