# openaccess-tools

Tools for working with the Smithsonian Open Access release

## Important

This is work in progress. Proper documentation to follow.

## Tools

### emit

```
> go run -mod vendor cmd/emit/main.go -h
  -bucket-uri string
    	A valid GoCloud bucket URI.
  -format-json
    	Format JSON output for each record.
  -json
    	Emit a JSON list.
  -null
    	Emit to /dev/null
  -stats
    	Display timings and statistics.
  -stdout
    	Emit to STDOUT (default true)
  -validate-json
    	Ensure each record is valid JSON. (default true)
  -workers int
    	The maximum number of concurrent workers. This is used to prevent filehandle exhaustion. (default 10)
```

A command-line tool for parsing and emitting individual records from a directory containing compressed and line-delimited Smithsonian OpenAccess JSON files.

For example:

```
go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess -json metadata/objects/NASM/ | jq '.[]["title"]' | grep 'Space' | sort
"Medal, NASA Space Flight, Sally Ride"
"Medal, STS-7, Smithsonian National Air and Space Museum, Sally Ride"
"Mirror, Primary Backup, Hubble Space Telescope"
"Model, 1:5, Hubble Space Telescope"
"Model, Space Shuttle, Delta-Wing High Cross-Range Orbiter Concept"
"Model, Space Shuttle, Final Orbiter Concept"
"Model, Space Shuttle, North American Rockwell Final Design, 1:15"
"Model, Space Shuttle, Straight-Wing Low Cross-Range Orbiter Concept"
"Model, Wind Tunnel, Convair Space Shuttle, 0.006 scale"
"Orbiter, Space Shuttle, OV-103, Discovery"
"Space Food, Beef and Vegetables, Mercury, Friendship 7"
"Spacecraft, Mariner 10, Flight Spare"
"Spacecraft, New Horizons, Mock-up, model"
"Suit, SpaceShipOne, Mike Melvill"
```

Or:

```
go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess -json -stats metadata/objects/CHNDM/ | jq '.[]["title"]' | grep -i 'kitten' | sort

2020/06/26 09:45:15 Processed 43695 records in 4.175884858s
"Cat and kitten"
"Tabby's Kittens"
```

Or:

```
go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess -json -format-json -validate-json=false -stats metadata/objects/CHNDM | grep '"title"' | grep -i 'kitten' | sort
2020/06/26 10:02:59 Processed 43695 records in 5.045081835s
  "title": "Cat and kitten"
  "title": "Tabby\u0027s Kittens"
```

## See also

* https://github.com/Smithsonian/OpenAccess