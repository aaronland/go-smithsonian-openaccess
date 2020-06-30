# go-smithsonian-openaccess

Go package for working with the Smithsonian Open Access release

## Important

This is work in progress. Proper documentation to follow.

## Tools

### emit

A command-line tool for parsing and emitting individual records from a directory containing compressed and line-delimited Smithsonian OpenAccess JSON files.

```
$> go run -mod vendor cmd/emit/main.go -h
  -bucket-uri string
    	A valid GoCloud bucket file:// URI.
  -format-json
    	Format JSON output for each record.
  -json
    	Emit a JSON list.
  -null
    	Emit to /dev/null
  -oembed
    	Emit results as OEmbed records
  -query value
    	One or more {PATH}={REGEXP} parameters for filtering records.
  -query-mode string
    	Specify how query filtering should be evaluated. Valid modes are: ALL, ANY (default "ALL")
  -stats
    	Display timings and statistics.
  -stdout
    	Emit to STDOUT (default true)
  -validate-edan
    	Ensure each record is a valid EDAN document.
  -validate-json
    	Ensure each record is valid JSON.
  -workers int
    	The maximum number of concurrent workers. This is used to prevent filehandle exhaustion. (default 10)
```

For example, processing every record in the OpenAccess dataset ensuring it is valid JSON and emitting it to `/dev/null`:

```
> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
  -stdout=false \
  -validate-json \  		
  -null \
  -stats \
  -workers 20 \
  metadata/objects

2020/06/26 10:19:17 Processed 11620642 records in 12m1.141284159s
```

Or processing everything in the [Air and Space](https://airandspace.si.edu/collections) collection as JSON, passing the result to the `jq` tool and searching for things with "space" in the title:

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -validate-json \  		   
   metadata/objects/NASM/ \
   | jq '.[]["title"]' \
   | grep -i 'space' \
   | sort

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

Or doing the same, but for [things about kittens](https://collection.cooperhewitt.org/objects/18382391/) in the [Cooper Hewitt](https://collection.cooperhewitt.org) collection:

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -validate-json \  		      
   -stats \
   metadata/objects/CHNDM/ \
   | jq '.[]["title"]' \
   | grep -i 'kitten' \
   | sort

2020/06/26 09:45:15 Processed 43695 records in 4.175884858s
"Cat and kitten"
"Tabby's Kittens"
```

Or something similar by not emitting a JSON list but formatting each record (as JSON) and filtering for the words "title" and "kitten":

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -format-json \
   -validate-json=false \
   -stats \
   metadata/objects/CHNDM \
   | grep '"title"' \
   | grep -i 'kitten' \
   | sort
   
2020/06/26 10:02:59 Processed 43695 records in 5.045081835s
  "title": "Cat and kitten"
  "title": "Tabby\u0027s Kittens"
```

#### Inline queries

You can also specify inline queries by passing a `-query` parameter which is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -query 'title=cats?\s+' \
   metadata/objects/CHNDM \
   | jq '.[]["title"]'
   
"View of Moat Mountain from Wildcat Brook, Jackson, New Hampshire, Looking Southwest"
"Near Falls of Wildcat Brook, Jackson, New Hampshire"
```

You can pass multiple `-query` parameters:

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -query 'title=cats?\s+' \
   -query 'title=(?i)^view' \
   metadata/objects/CHNDM \
   | jq '.[]["title"]'
   
"View of Moat Mountain from Wildcat Brook, Jackson, New Hampshire, Looking Southwest"
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by passing the `-query-mode ANY` flag:

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -query 'title=cats?\s+' \
   -query 'title=(?i)^view' \
   -query-mode ANY \
   metadata/objects/CHNDM \
   | jq '.[]["title"]'
   
"View of Santi Giovanni e Paolo a Celio, Rome"
"View of a Morning Room Interior"
"View of the Louvre from the River"
"View of the Acropolis, Athens"
"View of Santi Giovanni e Paolo a Celio, Rome"
"Views Representing the Most Considerable Transactions in the Siege of a Place, from Twelve of the Most REmarkable Sieges and Battles in Europe"
"View of Shiba Coast (Shibaura no fukei) From the Series One Hundred Famous views of Edo"
"View of Florence, Plate from \"Scelta di XXIV Vedute delle principali contrade, piazze, chiese, e palazzi della Città di Firenze\""
"View of Venice, Italy"
"View Across a River"
"View of the Canadian Falls and Goat Island"

...and so on
```

Did you know that there are 61 (out of 11 million) objects in the Smithsonian collection with the word "kitten" in their title?

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -query 'title=(?i)kitten' \
   -stats \
   -workers 50 \
   metadata/objects \
   | jq '.[]["title"]'
   
2020/06/26 18:22:04 Processed 62 records in 5m9.567738657s
"Cat and kitten"
"Tabby's Kittens"
"Ye Kitten (Number 17 May 1944)"
"The Kitten (No. 15 March 1944)"
"Three kittens on a stool"
"I'll Never Go Back On My Word; Leave My Kitten Alone"
"Untitled (Two Kittens)"
"Kitten Number Nine"
"Kitten No. Six"
"Bashful Baby Blues; Kitten On the Keys"
"Let's Make Believe We're Sweethearts; Three Naughty Kittens"
"Kittens playing"
"Take Me Back Again; Listen Kitten"
"Kittens"
"Diga Diga Do; Kitten WIth the Big Green Eyes, The"
"I Ain't Nothin' But a Tomcat's Kitten; I'm On My Way"
"Figurine, Kitten Small plastic"
"All the Time; Leave My Kitten Alone"
"Kitten On the Keys; That Place Down the Road Apiece"
"One Dime Blues; Three Little Kittens Rag"
"Kitten No. Eleven"
"I Ain't Nothin' But a Tomcat's Kitten; I'm On My Way"
"Weaker Kitten No. 2/64/41"
"Reward of Merit with Boy and Girl Playing with Cat and Kittens"
"Two Dollar Rag; Kitten on the Keys"
"The Kitten (No. 13 January 1944)"
"Kittens Playing with Camera"
"Reward of Merit with Two White Kittens in Basket"
"Doug and Toad - Kitten on Stump, 1942"
"I'll Never Go Back On My Word; Leave My Kitten Alone"
"The Kitten (No. 40 Sept. 1953)"
"Diga Diga Do; Kitten With the Big Green Eyes, The"
"The Kitten's Breakfast"
"Bunch Of Keys, A; Kitten On the Keys"
"The Kitten (No. 14 February 1944)"
"Figurine, Siamese Kitten"
"Tom Kitten"
"Young boy and his kitten"
"The Young Kittens"
"Little Kittens Learning Abc"
"Jump Jump of Holiday House. Three Little Kittens."
"The Kitten (Number 25 November 1946)"
"Kitten in Shoe"
"The Color Kittens"
"All the Time; Leave My Kitten Alone"
"Kitten on a Stool"
"Super Kitten; We'd Better Stop"
"Kitten mitten roller derby button"
"Little Girl holding Kitten"
"The Kitten (No. 39 May 27, 1953)"
"The Kitten (Number 18 June 1944)"
"My Love Is a Kitten; Strange Little Melody, The"
"Live and Let Live; Tom Cat's Kitten"
"Live and Let Live; Tom Cat's Kitten"
"One Dime Blues; Three Little Kittens Rag"
"The Kitten (No. 47 Dec. 1954)"
"Kittens Playing with Camera"
"\"Okimono\" Figure Of A Cat And Three Kittens"
"Mummy Of \"Kitten\""
"Plicate Kitten's Paw"
"Atlantic Kitten's Paw"
"Drosophila arawakana kittensis"
```

#### Oembed

It is also possible to emit OpenAccess records as [OEmbed](https://oembed.com/) documents of type "photo". An OEmbed record will be created for each media object of type "Screen Image" or "Images" associated with an OpenAccess record. OpenAccess records that do not have an suitable media objects will be excluded.

The OEmbed record `title` property will be constructed in the form of "{OBJECT TITLE} ({OBJECT CREDIT LINE}".

The OEmbed record `author_name` property will be constructed using the OpenAccess record's `content.freetext.name` or `content.freetext.manufacturer` properties, in that order. If neither are present the `author_name` property will be constructed in the form of "Collection of {SMITHSONIAN UNIT NAME}".

The OEmbed record `width` and `height` property are both set to "-1" to indicate that image dimensions are [not available at this time](https://github.com/Smithsonian/OpenAccess/issues/2).

```
$> go run -mod vendor cmd/emit/main.go -bucket-uri file:///usr/local/OpenAccess \
   -json \
   -oembed \
   metadata/objects/SAAM \
   | jq
   
[
  {
    "version": "1.0",
    "type": "photo",
    "width": -1,
    "height": -1,
    "title": "San Piero a Grado, near Pisa (Smithsonian American Art Museum, Bequest of Carolann Smurthwaite in memory of her mother, Caroline Atherton Connell Smurthwaite)",
    "url": "https://ids.si.edu/ids/download?id=SAAM-1983.83.134_1_screen",
    "author_name": "George Elbert Burr, born Monroe Falls, OH 1859-died Phoenix, AZ 1939",
    "author_url": "http://americanart.si.edu/collections/search/artwork/?id=3299",
    "provider_name": "Smithsonian American Art Museum",
    "provider_url": "http://americanart.si.edu"
  },
  {
    "version": "1.0",
    "type": "photo",
    "width": -1,
    "height": -1,
    "title": "Politician (Smithsonian American Art Museum, Gift of Jack Lord)",
    "url": "https://ids.si.edu/ids/download?id=SAAM-1971.439.22_1_screen",
    "author_name": "José Guadalupe Posada, Mexican, Aguascalientes, Mexico 1852-died Mexico City, Mexico 1913",
    "author_url": "http://americanart.si.edu/collections/search/artwork/?id=19951",
    "provider_name": "Smithsonian American Art Museum",
    "provider_url": "http://americanart.si.edu"
  },
  ... and so on
```

### findingaid

A command-line tool for emitting a CSV document mapping individual record identifiers to their corresponding OpenAccess JSON file and line number, produced from a directory containing compressed and line-delimited Smithsonian OpenAccess JSON files.

```
> go run -mod vendor cmd/findingaid/main.go -h
  -bucket-uri string
    	A valid GoCloud bucket file:// URI.
  -csv-header
    	Include a CSV header row in the output (default true)
  -include-all
    	Include all OpenAccess identifiers
  -include-guid content.descriptiveNonRepeating.guid
    	Include the OpenAccess content.descriptiveNonRepeating.guid identifier
  -include-openaccess-id id
    	Include the OpenAccess id identifier
  -include-record-id content.descriptiveNonRepeating.record_ID
    	Include the OpenAccess content.descriptiveNonRepeating.record_ID identifier (default true)
  -include-record-link content.descriptiveNonRepeating.record_link
    	Include the OpenAccess content.descriptiveNonRepeating.record_link identifier
  -null
    	Emit to /dev/null
  -query value
    	One or more {PATH}={REGEXP} parameters for filtering records.
  -query-mode string
    	Specify how query filtering should be evaluated. Valid modes are: ALL, ANY (default "ALL")
  -stats
    	Display timings and statistics.
  -stdout
    	Emit to STDOUT (default true)
  -workers int
    	The maximum number of concurrent workers. This is used to prevent filehandle exhaustion. (default 10)
```

```
$> go run -mod vendor cmd/findingaid/main.go -bucket-uri file:///usr/local/OpenAccess \
	metadata/objects/SAAM 

id,path,line_number
saam_1971.439.94,metadata/objects/SAAM/00.txt.bz2,1
saam_1971.439.92,metadata/objects/SAAM/08.txt.bz2,1
saam_1915.5.1,metadata/objects/SAAM/00.txt.bz2,2
saam_1971.439.78,metadata/objects/SAAM/03.txt.bz2,1
saam_XX32,metadata/objects/SAAM/12.txt.bz2,1
saam_1983.90.173,metadata/objects/SAAM/00.txt.bz2,3
saam_1970.335.1,metadata/objects/SAAM/03.txt.bz2,2
saam_1971.439.97,metadata/objects/SAAM/0d.txt.bz2,1
saam_1968.155.158,metadata/objects/SAAM/12.txt.bz2,2
saam_1967.14.149,metadata/objects/SAAM/08.txt.bz2,2
saam_1979.98.188,metadata/objects/SAAM/02.txt.bz2,1
saam_1985.66.295_540,metadata/objects/SAAM/00.txt.bz2,4
saam_1970.334,metadata/objects/SAAM/03.txt.bz2,3
saam_1968.19.12,metadata/objects/SAAM/0d.txt.bz2,2
... and so on
```

```
$> go run -mod vendor cmd/findingaid/main.go -bucket-uri file:///usr/local/OpenAccess \
   -include-all \
   metadata/objects/NMAAHC
   
id,path,line_number
http://n2t.net/ark:/65665/fd53f870fc2-73af-4c50-b1c5-a3fd2829ad1f,metadata/objects/NMAAHC/ff.txt.bz2,1
nmaahc_2014.72.2,metadata/objects/NMAAHC/ff.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2014.72.2,metadata/objects/NMAAHC/ff.txt.bz2,1
edanmdm-nmaahc_2014.72.2,metadata/objects/NMAAHC/ff.txt.bz2,1
http://n2t.net/ark:/65665/fd5343a21ed-73d9-4014-a34c-b175b84168c8,metadata/objects/NMAAHC/21.txt.bz2,1
nmaahc_2014.75.130,metadata/objects/NMAAHC/21.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2014.75.130,metadata/objects/NMAAHC/21.txt.bz2,1
edanmdm-nmaahc_2014.75.130,metadata/objects/NMAAHC/21.txt.bz2,1
http://n2t.net/ark:/65665/fd59212a6e2-b745-4eb9-84ad-4368ffea8223,metadata/objects/NMAAHC/17.txt.bz2,1
nmaahc_2016.140.1.3,metadata/objects/NMAAHC/17.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2016.140.1.3,metadata/objects/NMAAHC/17.txt.bz2,1
edanmdm-nmaahc_2016.140.1.3,metadata/objects/NMAAHC/17.txt.bz2,1
http://n2t.net/ark:/65665/fd599a84051-37d5-49d4-98d3-9052e5cbcea9,metadata/objects/NMAAHC/22.txt.bz2,1
nmaahc_2012.30.3,metadata/objects/NMAAHC/22.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2012.30.3,metadata/objects/NMAAHC/22.txt.bz2,1
edanmdm-nmaahc_2012.30.3,metadata/objects/NMAAHC/22.txt.bz2,1
http://n2t.net/ark:/65665/fd53a114ad8-2cc2-4ce2-bbd0-6dd09cc715df,metadata/objects/NMAAHC/0c.txt.bz2,1
nmaahc_2013.133.1.4,metadata/objects/NMAAHC/0c.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2013.133.1.4,metadata/objects/NMAAHC/0c.txt.bz2,1
edanmdm-nmaahc_2013.133.1.4,metadata/objects/NMAAHC/0c.txt.bz2,1
http://n2t.net/ark:/65665/fd5d302d893-ae7c-4b4d-93bb-59f87237d23a,metadata/objects/NMAAHC/1c.txt.bz2,1
nmaahc_2014.222.2,metadata/objects/NMAAHC/1c.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2014.222.2,metadata/objects/NMAAHC/1c.txt.bz2,1
edanmdm-nmaahc_2014.222.2,metadata/objects/NMAAHC/1c.txt.bz2,1
http://n2t.net/ark:/65665/fd5ab09d12b-42bc-40f2-9557-b924d182723e,metadata/objects/NMAAHC/ff.txt.bz2,2
nmaahc_2016.166.17,metadata/objects/NMAAHC/ff.txt.bz2,2
https://nmaahc.si.edu/object/nmaahc_2016.166.17,metadata/objects/NMAAHC/ff.txt.bz2,2
edanmdm-nmaahc_2016.166.17,metadata/objects/NMAAHC/ff.txt.bz2,2
http://n2t.net/ark:/65665/fd588400c0f-66c3-4259-999d-57f112e05479,metadata/objects/NMAAHC/2b.txt.bz2,1
nmaahc_2014.263.5,metadata/objects/NMAAHC/2b.txt.bz2,1
https://nmaahc.si.edu/object/nmaahc_2014.263.5,metadata/objects/NMAAHC/2b.txt.bz2,1
... and so on
```

## See also

* https://github.com/Smithsonian/OpenAccess
* https://gocloud.dev/howto/blob/
* https://github.com/aaronland/go-jsonl/