# go-smithsonian-openaccess

Go package for working with the Smithsonian Open Access release

## Documentation

Documentation is incomplete.

## Data sources

This package and the tools it exports support two types of data sources for the Smithsonian Open Access: A local file system and an AWS S3 bucket. Under the hood the code is using the GoCloud [blob](https://godoc.org/gocloud.dev/blob) abstraction layer so other [storage services](https://gocloud.dev/howto/blob/) could be supported but currently they are not.

Access to the data on a local file system is presumed to be a clone of the [OpenAccess](https://github.com/Smithsonian/OpenAccess) S3 bucket (described on the [Smithsonian Open Access AWS Registry page](https://registry.opendata.aws/smithsonian-open-access/). A local copy of this data can be created using the `clone` tool described below.

The data itself also lives in a Smithsonian-operated AWS S3 bucket so this code has been updated to retrieve data from there if asked to. For a number of reasons specific to the Smithsonian retrieving data from their S3 bucket does not fit neatly in to the `GoCloud` abstraction layer but efforts have been made to hide those details from users of this code.

Most of the examples below assume a local Git checkout. For example:

```
$> ./bin/emit -bucket-uri file:///usr/local/OpenAccess metadata/edan/nmah
```

In order to retrieve data from the Smithsonian-operated S3 bucket you would change the `-bucket-uri` flag to:

```
$> ./bin/emit -bucket-uri 's3://smithsonian-open-access?region=us-west-2' metadata/edan/nmah
```

Or the following, which is included as a convenience method:

```
$> ./bin/emit -bucket-uri 'si://' metadata/edan/nmah
```

A by-product of this work is that the code is also able to retrieve data from any other S3 bucket. For example:

```
$> ./bin/emit -bucket-uri 's3://YOUR-OPENACCESS-BUCKET?region=YOUR-BUCKET-REGION' metadata/edan/nmah
```

As of this writing the code to retrieve data from S3 buckets (other than the Smithsonian's) assumes that those buckets allow public access and have public directory listings enabled.

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/clone cmd/clone/main.go
go build -mod vendor -o bin/emit cmd/emit/main.go
go build -mod vendor -o bin/findingaid cmd/findingaid/main.go
go build -mod vendor -o bin/location cmd/location/main.go
go build -mod vendor -o bin/placename cmd/placename/main.go
```

### clone

A command-line tool to clone OpenAccess data to a target destination.

This tool was written principally to clone OpenAccess data from the Smithsonian's `smithsonian-open-access` S3 bucket to a local filesystem but it can be used to clone data to and from any supported `GoCloud.blob` source.

```
$> ./bin/clone -h
Usage:
  ./bin/clone [options] [path1 path2 ... pathN]

Options:
  -compress
    	Compress files in the target bucket using bzip2 encoding. Files will be appended with a '.bz2' suffix.
  -force
    	Clone files even if they are present in target bucket and MD5 hashes between source and target buckets match.
  -source-bucket-uri string
    	A valid GoCloud bucket URI. Valid schemes are: file://, s3:// and si:// which is signals that data should be retrieved from the Smithsonian's 'smithsonian-open-access' S3 bucket. (default "si://")
  -target-bucket-uri string
    	A valid GoCloud bucket URI. Valid schemes are: file://, s3://.
  -workers int
    	The maximum number of concurrent workers. This is used to prevent filehandle exhaustion. (default 10)
```

For example:

```
$> ./bin/clone \
	-source-bucket-uri si:// \
	-target-bucket-uri 'file:///usr/local/data/si/?metadata=skip' \
	-workers 100 \
	metadata/edan
	
...time passes

$> du -h -d 1 /usr/local/data/si
 37G	/usr/local/data/si/metadata
 37G	/usr/local/data/si
```

See the way we're only cloning the `metadata/edan` tree? If you don't specify a subfolder you'll end up cloning _all the images_ in the OpenAccess data and that might be prohibitive in terms of bandwidth and storage.

And then later on:

```
$> ./bin/emit \
	-json \
	-format-json \
	-bucket-uri file:///usr/local/data/si/metadata/edan \
	chndm

[{
  "id": "edanmdm-chndm_1931-45-37",
  "version": "",
  "unitCode": "CHNDM",
  "linkedId": "0",
  "type": "edanmdm",
  "content": {
    "descriptiveNonRepeating": {
      "record_ID": "chndm_1931-45-37",
      "online_media": {
        "mediaCount": 1,
  ...and so on
}]
```

#### Notes

* If no extra URI or URIs (for example `metadata/edan/chndm`) are specified then the code will attempt to clone everything in the "source" bucket recursively.

* Under the hood this code is using the [GoCloud blob abstraction layer](https://gocloud.dev/howto/blob/). The default behaviour for the abstraction is to assume restrictive permissions when creating new files. Unfortunately, as of this writing, there is no common way for assigning permissions using the `GoCloud blob` abstraction so this is something you'll need to account for separately from this tool.

### emit

A command-line tool for parsing and emitting individual records from a directory containing compressed and line-delimited Smithsonian OpenAccess JSON files.

```
$> ./bin/emit -h
Usage:
  ./bin/emit [options] [path1 path2 ... pathN]

Options:
  -bucket-uri string
    	A valid GoCloud bucket URI. Valid schemes are: file://, s3:// and si:// which is signals that data should be retrieved from the Smithsonian's 'smithsonian-open-access' S3 bucket.
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
$> ./bin/emit -bucket-uri file:///usr/local/data/si \
  -stdout=false \
  -validate-json \  		
  -null \
  -stats \
  -workers 20 \
  metadata/edan/cfchfolklife

2021/12/22 14:52:57 Processed 59410 records in 5.713120297s
```

Or processing everything in the [Air and Space](https://airandspace.si.edu/collections) collection as JSON, passing the result to the `jq` tool and searching for things with "space" in the title:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si \
   -json \
   -validate-json \  		   
   metadata/edan/nasm/ \
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
$> ./bin/emit -bucket-uri file:///usr/local/data/si \
   -json \
   -validate-json \  		      
   -stats \
   metadata/edan/chndm/ \
   | jq '.[]["title"]' \
   | grep -i 'kitten' \
   | sort

2020/06/26 09:45:15 Processed 43695 records in 4.175884858s
"Cat and kitten"
"Tabby's Kittens"
```

Or something similar by not emitting a JSON list but formatting each record (as JSON) and filtering for the words "title" and "kitten":

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si \
   -format-json \
   -validate-json=false \
   -stats \
   metadata/edan/chndm \
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
$> ./bin/emit -bucket-uri file:///usr/local/data/si\
   -json \
   -query 'title=cats?\s+' \
   metadata/edan/chndm \
   | jq '.[]["title"]'
   
"View of Moat Mountain from Wildcat Brook, Jackson, New Hampshire, Looking Southwest"
"Near Falls of Wildcat Brook, Jackson, New Hampshire"
```

You can pass multiple `-query` parameters:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si\
   -json \
   -query 'title=cats?\s+' \
   -query 'title=(?i)^view' \
   metadata/objects/CHNDM \
   | jq '.[]["title"]'
   
"View of Moat Mountain from Wildcat Brook, Jackson, New Hampshire, Looking Southwest"
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by passing the `-query-mode ANY` flag:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si\
   -json \
   -query 'title=cats?\s+' \
   -query 'title=(?i)^view' \
   -query-mode ANY \
   metadata/edan/chndm \
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
$> ./bin/emit -bucket-uri file:///usr/local/data/si\
   -json \
   -query 'title=(?i)kitten' \
   -stats \
   -workers 50 \
   metadata/edan \
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

#### OEmbed

It is also possible to emit OpenAccess records as [OEmbed](https://oembed.com/) documents of type "photo". An OEmbed record will be created for each media object of type "Screen Image" or "Images" associated with an OpenAccess record. OpenAccess records that do not have an suitable media objects will be excluded.

For example:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si \
   -json \
   -oembed \
   metadata/edan/nasm \
   | jq
   
[
{
  "version": "1.0",
  "type": "photo",
  "width": -1,
  "height": -1,
  "title": "Clerget 9 A Diesel, Radial 9 Engine (Gift of the Musee de L' Air)",
  "url": "https://ids.si.edu/ids/download?id=NASM-A19721334000-NASM2016-04025_screen",
  "author_name": "Clerget, Blin and Cie",
  "author_url": "https://airandspace.si.edu/collection/id/nasm_A19721334000",
  "provider_name": "National Air and Space Museum",
  "provider_url": "https://airandspace.si.edu",
  "object_uri": "si://nasm/o/A19721334000"
},
{
  "version": "1.0",
  "type": "photo",
  "width": -1,
  "height": -1,
  "title": "Swagger Stick, Royal Flying Corps (Gift of Eloise and John Charlton)",
  "url": "https://ids.si.edu/ids/download?id=NASM-A19830196000_PS01_screen",
  "author_name": "Lt. Wes D. Archer",
  "author_url": "https://airandspace.si.edu/collection/id/nasm_A19830196000",
  "provider_name": "National Air and Space Museum",
  "provider_url": "https://airandspace.si.edu",
  "object_uri": "si://nasm/o/A19830196000"
}
  ... and so on
```

* The OEmbed record `title` property will be constructed in the form of "{OBJECT TITLE} ({OBJECT CREDIT LINE})".

* The OEmbed record `author_name` property will be constructed using the OpenAccess record's `content.freetext.name` or `content.freetext.manufacturer` properties, in that order. If neither are present the `author_name` property will be constructed in the form of "Collection of {SMITHSONIAN UNIT NAME}".

* The OEmbed record `author_url` property will that object's URL on the web.

* The OEmbed record `width` and `height` properties are both set to "-1" to indicate that image dimensions are [not available at this time](https://github.com/Smithsonian/OpenAccess/issues/2).

* The OEmbed record will contain a non-standard `object_uri` string that is compatiable with [RFC6570 URI Templates](https://tools.ietf.org/html/rfc6570). It is constructed in the form of `si://{SMITHSONIAN_UNIT}/o/{NORMALIZAED_EDAN_OBJECT_ID}`. The `object_uri` property should still be considered experimental. It may change or be removed in future releases.

* `{NORMALIZAED_EDAN_OBJECT_ID}` strings are derived from the OpenAccess `id` property. The normalization rules are: Remove the leading `edanmdm-{SMITHSONIAN_UNIT}_` prefix and replace all instances of the `.` character the with a `_` character. For example the string `edanmdm-nmaahc_2017.30.9` will be normalized as `2017_30_9`.

### findingaid

A command-line tool for emitting a CSV document mapping individual record identifiers to their corresponding OpenAccess JSON file and line number, produced from a directory containing compressed and line-delimited Smithsonian OpenAccess JSON files.

```
$> ./bin/findingaid -h
Usage:
  ./bin/findingaid [options] [path1 path2 ... pathN]

Options:
  -bucket-uri string
    	A valid GoCloud bucket URI. Valid schemes are: file://, s3:// and si:// which is signals that data should be retrieved from the Smithsonian's 'smithsonian-open-access' S3 bucket.
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

For example:

```
$> ./bin/findingaid -bucket-uri file:///usr/local/data/si \
	metadata/edan/saam

id,path,line_number
saam_1909.7.47,metadata/edan/saam/eb.txt,36
saam_1979.135.66,metadata/edan/saam/e7.txt,50
saam_1976.113.20,metadata/edan/saam/e6.txt,50
saam_1970.355.503,metadata/edan/saam/ed.txt,31
saam_1985.66.153_298,metadata/edan/saam/ec.txt,36
saam_1971.292.9,metadata/edan/saam/e8.txt,44
saam_1973.122.51,metadata/edan/saam/ea.txt,41
saam_2000.110,metadata/edan/saam/e9.txt,44
saam_1929.6.159,metadata/edan/saam/e5.txt,64
saam_1946.10.2,metadata/edan/saam/eb.txt,37
saam_1983.17.2,metadata/edan/saam/e7.txt,51
saam_1970.125,metadata/edan/saam/e6.txt,51
saam_1970.355.388,metadata/edan/saam/ed.txt,32
saam_1985.66.153_325,metadata/edan/saam/ec.txt,37
saam_1970.335.16,metadata/edan/saam/e8.txt,45
saam_1979.135.56,metadata/edan/saam/ea.txt,42
saam_1971.446.174,metadata/edan/saam/e9.txt,45
saam_1973.130.114,metadata/edan/saam/e5.txt,65
saam_1971.244,metadata/edan/saam/eb.txt,38
saam_1929.6.105,metadata/edan/saam/e7.txt,52
saam_2000.83.50,metadata/edan/saam/e6.txt,52
saam_2017.24.4,metadata/edan/saam/ed.txt,33
saam_1962.8.15,metadata/edan/saam/ec.txt,38
saam_1979.135.72,metadata/edan/saam/e8.txt,46
saam_1967.63.39,metadata/edan/saam/ea.txt,43
saam_1974.13.2,metadata/edan/saam/e9.txt,46
... and so on
```

By default only the `OpenAccess content.descriptiveNonRepeating.record_ID` identifier is included in the finding aid. You can include other identifiers with their corresponding command-line flag or enable include all identifiers by passing the `-include-all` flag. For example:

```
$> ./bin/findingaid -bucket-uri file:///usr/local/data/si \
   -include-all \
   metadata/edan/nmaahc

id,path,line_number
ebl-1554224450096-1554224455450-0,metadata/edan/nmaahc/01.txt,1
ebl-1503511359102-1503511359173-3,metadata/edan/nmaahc/ff.txt,1
ebl-1586784637019-1586784637080-3,metadata/edan/nmaahc/00.txt,1
ebl-1554224450096-1554224455453-0,metadata/edan/nmaahc/c0.txt,1
ebl-1554224450096-1554224455616-0,metadata/edan/nmaahc/01.txt,2
ebl-1586797256118-1586797256421-2,metadata/edan/nmaahc/80.txt,1
ebl-1586784637019-1586784637078-4,metadata/edan/nmaahc/81.txt,1
ebl-1586797256118-1586797256344-2,metadata/edan/nmaahc/82.txt,1
ebl-1554224450096-1554224455608-1,metadata/edan/nmaahc/c1.txt,1
ebl-1554224450096-1554224455535-0,metadata/edan/nmaahc/02.txt,1
ebl-1519826450657-1519826450770-2,metadata/edan/nmaahc/ff.txt,2
ebl-1594040408251-1594040408443-1,metadata/edan/nmaahc/00.txt,2
ebl-1554224450096-1554224455622-3,metadata/edan/nmaahc/c0.txt,2
ebl-1554224450096-1554224455534-1,metadata/edan/nmaahc/01.txt,3
ebl-1525728005819-1525728005875-1,metadata/edan/nmaahc/83.txt,1
ebl-1554224450096-1554224455544-3,metadata/edan/nmaahc/80.txt,2
ebl-1525783231307-1525783231525-0,metadata/edan/nmaahc/81.txt,2
ebl-1588593635567-1588593635698-2,metadata/edan/nmaahc/c1.txt,2
ebl-1586797256118-1586797256300-1,metadata/edan/nmaahc/02.txt,2
ebl-1586797256118-1586797256406-2,metadata/edan/nmaahc/ff.txt,3
ebl-1519826450657-1519826450821-5,metadata/edan/nmaahc/00.txt,3
... and so on
```

The `findingaid` tool also supports inline queries (described above). For example there are 4044 records with the word "panda" in their title:

```
$> ./bin/findingaid -bucket-uri file:///usr/local/data/si \
   -query 'title=(?i)pandas?' \
   -workers 50 \
   metadata/edan/ \
   > pandas.csv

time passes...

$> wc -l pandas.csv
    23264 pandas.csv	

$> less pandas.csv
id,path,line_number
ebl-1503510573996-1503510574151-6,metadata/edan/aaa/49.txt,610
ebl-1503512355391-1503512355404-5,metadata/edan/aaa/5b.txt,348
ebl-1503512825373-1503512825482-2,metadata/edan/aaa/67.txt,444
ebl-1503513876560-1503513876610-8,metadata/edan/aaa/c3.txt,1261
ebl-1562776092361-1562776096447-4,metadata/edan/aag/9d.txt,27
ebl-1537785066473-1537785075553-1,metadata/edan/acah/20.txt,326
ebl-1505824233925-1505824234112-2,metadata/edan/acah/56.txt,447
ebl-1543431025153-1543431025397-1,metadata/edan/acah/55.txt,2161
ebl-1614774862593-1614774866357-2,metadata/edan/acah/62.txt,365
ebl-1550683206232-1550683206274-0,metadata/edan/acah/9e.txt,946
ebl-1503512575563-1503512575585-4,metadata/edan/acah/a6.txt,453
ebl-1568040184345-1568040186378-2,metadata/edan/acah/aa.txt,651
ebl-1505824233925-1505824234080-2,metadata/edan/acah/c2.txt,2078
ebl-1505824233925-1505824234116-1,metadata/edan/acah/d0.txt,289
ebl-1510071055254-1510071055423-6,metadata/edan/acah/ce.txt,1083
ebl-1510071055254-1510071055423-5,metadata/edan/acah/d2.txt,458
ebl-1562715031827-1562715031860-10,metadata/edan/acah/df.txt,1148
ebl-1562715031827-1562715031860-9,metadata/edan/acah/fc.txt,46
siris_arc_347337,metadata/edan/cfchfolklife/15.txt,49
edanmdm-siris_arc_347337,metadata/edan/cfchfolklife/15.txt,49
ebl-1503510195028-1503510195534-3,metadata/edan/cfchfolklife/1c.txt,191
ebl-1539103222248-1539103222432-0,metadata/edan/cfchfolklife/4a.txt,66
ebl-1612558900451-1612558901326-2,metadata/edan/cfchfolklife/44.txt,139
... and so on
```

### location

A command-line tool for parsing line-delimited Smithsonian OpenAccess JSON files and emiting place data as a stream of CSV records.

```
$> ./bin/location -h
Usage:
  ./bin/location [options]

Options:
  -null
    	Emit to /dev/null
  -stdout
    	Emit to STDOUT (default true)
```

For example:

```
$> ./bin/emit \
	-bucket-uri file:///usr/local/data/si metadata/edan/nmah/ \

   | ./bin/location

edanmdm-nmah_1097068,content.freetext.place,place made,United States: Connecticut
edanmdm-nmah_737712,content.freetext.place,place made,"United States: Illinois, Chicago"
edanmdm-nmah_322930,content.freetext.place,place made,"United States: New York, New York"
edanmdm-nmah_322930,content.freetext.place,place family from,"United States: New Hampshire, Laconia"
edanmdm-nmah_606951,content.freetext.place,place made,"United States: New York, New York City"
edanmdm-nmah_607006,content.freetext.place,place made,United States
edanmdm-nmah_737768,content.freetext.place,place made,"United States: Indiana, Indianapolis"
edanmdm-nmah_607041,content.freetext.place,place made,United States
edanmdm-nmah_1085323,content.freetext.place,associated place,United States
edanmdm-nmah_554669,content.freetext.place,Associated Place,"United States: Maryland, Bethesda"
edanmdm-nmah_1873824,content.freetext.place,place made,United States
edanmdm-nmah_1055983,content.freetext.place,associated place,"United States: Virginia, Hayes"
edanmdm-nmah_1331161,content.freetext.place,place made,"United States: New York, Brooklyn"
edanmdm-nmah_682266,content.freetext.place,place made,United States
edanmdm-nmah_1055994,content.freetext.place,associated place,"United States: Virginia, Hayes"
edanmdm-nmah_1055994,content.freetext.place,associated place,"United States: Virginia, Hayes"
edanmdm-nmah_1328760,content.freetext.place,Place Made,"United States: California, Cupertino"
edanmdm-nmah_445379,content.freetext.place,place made,Brazil
...and so on
```

The column in the CSV output are:

| Index | Value | Example |
| --- | --- | --- |
| 0 | OpenAccess record ID | edanmdm-nmah_715051 |
| 1 | Path to the property used to lookup place data | content.freetext.place |
| 2 | Label associated with place data | place made |
| 3 | Place name | "United States: New York, New York City" |

### placename

A command-line tool for extracting only placename data from a CSV stream produced by the `location` tool.

```
$> ./bin/placename -h
Usage:
  ./bin/placename [options]

Options:
  -unique
    	Only unique emit placename strings once. (default true)
```

For example:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si/metadata/edan/nmah \

   | ./bin/location \
   
   | ./bin/placename \
   
   | wc -l
   
12164
```

Or:

```
$> ./bin/emit -bucket-uri file:///usr/local/data/si/metadata/edan/ \

   | ./bin/location \
   
   | ./bin/placename

United States
Washington (D.C.)
Florida
Miami (Fla.)
France
USA
Italy
Europe
Italy or Spain
France, Europe
probably Venice, Italy
Milan, Italy
France or Italy
Florence, Italy

...time passes

From top of pass to Hoja Verde., Tamaulipas, Mexico, North America
Perto Dom Pedro II, Paraná, Brazil, South America
Zealand: peat-bog at Søgärd., Denmark, Europe
Tarumã Alta, 14 km NW of Manaus., Manaus, Brazil, South America - Neotropics
Woods near Taxodium swamp, 2 miles south of Eagletown, McCurtain Co., Oklahoma, United States, North America
Yarmouth County. Deep water of St. John (Wilson's) Lake., Nova Scotia, Canada, North America
½ mi. S. Olivet., Osage, Kansas, United States, North America
Range of low hills ca. 20 km west of Redenção, near Córrego São João and Troncamento Santa Teresa, Conceição do Araguaia, Brazil, South America - Neotropics
Tatama. Santa Cecilia. Cordillera Occidental. Vertiente Occidental, Caldas, Colombia, South America - Neotropics
San Rafael Ranch - on banky rivers. Cameron Co, Texas, United States, North America
Sultanabad, Khorassan., Khorasan [obsolete], Iran, Asia-Temperate
Pointe Du Lac, comte du St-Maurice: sur les sables du lac St-Pierre., Quebec, Canada, North America

...2.5M records later

Limburg
Japão
Lado Enclave (Congo Free State)
Mauritanie
Igboho (Nigeria)
Accra Plains
Hollywood (Fla.)
Broward County (Fla.)
GrÃ-Bretanha
América latina
Cousin
Peace River Watershed (B.C. and Alta.)
Peace River Watershed
```

## See also

* https://github.com/Smithsonian/OpenAccess
* https://gocloud.dev/howto/blob/
* https://github.com/aaronland/go-jsonl/
* https://github.com/aaronland/go-json-query/
