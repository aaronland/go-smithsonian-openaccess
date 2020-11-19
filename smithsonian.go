package openaccess

var SMITHSONIAN_UNITS []string

func init() {

	// https://github.com/Smithsonian/OpenAccess/tree/master/metadata/objects
	// It would be better if we didn't have to hard code this but today we do...
	// (20201118/straup)
	
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
}
