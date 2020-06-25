# openaccess-tools

Tools for working with the Smithsonian Open Access release

## Important

This is work in progress.

## Tools

### emit

This is poorly named tool for simultaneously parsing, pretty-printing and emitting (currently just to `STDOUT`) individual JSON records from a raw SI OpenAccess file.

```
> go run -mod vendor cmd/emit/main.go ~/code/OpenAccess/metadata/objects/CHNDM/* | less
{
  "content": {
    "descriptiveNonRepeating": {
      "data_source": "Cooper Hewitt, Smithsonian Design Museum",
      "guid": "http://n2t.net/ark:/65665/kq4c36db6e4-94dc-4820-8cd4-03bc91752157",
      "metadata_usage": {
        "access": "CC0"
      },
      "online_media": {
        "media": [
          {
            "content": "http://ids.si.edu/ids/deliveryService?id=CHSDM-E0207E65ACA82-000001",
            "guid": "http://n2t.net/ark:/65665/vc9fd5f62e8-0400-4cfe-ab2f-6edf592fa86e",
            "idsId": "CHSDM-E0207E65ACA82-000001",
            "thumbnail": "http://ids.si.edu/ids/deliveryService?id=CHSDM-E0207E65ACA82-000001",
            "type": "Images",
            "usage": {
              "access": "CC0"
            }
          }
        ],
        "mediaCount": 1
      },
      "record_ID": "chndm_1931-66-88",
      "record_link": "http://collection.cooperhewitt.org/view/objects/asitem/id/48147",
      "title": {
        "content": "Cathedral of Notre Dame in Paris",
        "label": "Title"
      },
      "title_sort": "CATHEDRAL OF NOTRE DAME IN PARIS",
      "unit_code": "CHNDM"
    },
    "freetext": {
      "creditLine": [
        {
          "content": "Gift of Sarah Cooper Hewitt",
          "label": "Credit Line"
        }
      ],
      "dataSource": [
        {
          "content": "Cooper Hewitt, Smithsonian Design Museum",
          "label": "Data Source"
        }
      ],

...and so on
```

## See also

* https://github.com/Smithsonian/OpenAccess