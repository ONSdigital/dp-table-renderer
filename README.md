dp-table-renderer
================

Given json defining a table, capable of rendering a table in multiple formats

### Getting started


| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | :23100                                    | The host and port to bind to
| HOST                       | http://localhost:23100                    | The host name used to build URLs
| SHUTDOWN_TIMEOUT           | 5s                                        | The graceful shutdown timeout (`time.Duration` format)

### Endpoints

| url          | Method | Description                                                                      |
| ---          | ------ | -----------                                                                      |
| /render/html | POST   | Renders the (json) data provided in the post body as a self-contained html table |

### Request content

POST requests must contain a json body that defines the table to render, for example:

```json
{
  "filename": "abcd1234",
  "title": "This is an example table",
  "subtitle": "with a subtitle",
  "source": "Office for National Statistics",
  "type": "generated-table",
  "uri": "/path/to/the/table/json",
  "style_class": "foo bar",
  "row_formats": [
    {"row": 0, "heading": true},
    {"row": 1, "height": "5em", "vertical_align": "top"}
  ],
  "column_formats": [
    {"col": 0, "align": "right", "width": "5em", "heading": true},
    {"col": 1, "align": "right", "width": "4em", "heading": true}
  ],
  "cell_formats": [
    {"row": 0, "col": 0, "align": "center", "colspan": 2},
    {"row": 1, "col": 1, "align": "left", "vertical_align": "middle", "rowspan": 2},
    {"row": 3, "col": 0, "align": "left", "vertical_align": "top", "rowspan": 11}
  ],
  "data": [
    ["Date",null,"CPIH Index[1]\n(UK, 2015 = 100)","CPIH 12-\nmonth rate ","CPI Index[1]\n(UK, 2015=100)","CPI 12- \nmonth rate","OOH Index[1]\n(UK, 2015=100)","OOH 12-\nmonth rate "],
    ["2016","Nov","101.8","1.5","101.4","1.2","103.4","2.6"],
    [null,"Dec","102.2","1.8","101.9","1.6","103.6","2.6"],
    ["2017","Jan","101.8","1.9","101.4","1.8","103.8","2.5"],
    [null,"Feb","102.4","2.3","102.1","2.3","103.9","2.5"],
    [null,"Mar","102.7","2.3","102.5","2.3","104.0","2.4"],
    [null,"Apr","103.2","2.6","102.9","2.7","104.1","2.2"],
    [null,"May","103.5","2.7","103.3","2.9","104.2","2.1"],
    [null,"Jun","103.5","2.6","103.3","2.6","104.2","2.0"],
    [null,"Jul","103.5","2.6","103.2","2.6","104.4","2.0"],
    [null,"Aug","104.0","2.7","103.8","2.9","104.6","1.9"],
    [null,"Sep","104.3","2.8","104.1","3.0","104.8","1.9"],
    [null,"Oct","104.4","2.8","104.2","3.0","104.8","1.6"],
    [null,"Nov","104.7","2.8","104.6","3.1","104.9","1.5"]
  ],
  "footnotes": [
    "Footnotes are indexed from 1",
    "And can be referenced from any data element or title using square brackets: [ 1 ]",
    "Note that when cells include rowspan or colspan you should still include all the cells in data - the merged cells should be null or empty string",
    "The align/vertical-align properties of row column and cell formats are output in html as class attributes"
  ]
}
```
Merged cells can be specified using `colspan` and `rowspan` properties of `cell_format` elements.
Please note that the `data` array should include *all* cells (i.e. each row should contain the same number of cells), even if some of them have been merged. This is the same approach/format used by some javascript spreadsheet components such as [Handsontable](https://handsontable.com/).
The values provided for `align` and `vertical_align` in `column_formats`, `row_formats` and `cell_formats` are assumed to be the name of a css class that correctly aligns the element.

### Healthchecking

Currently reported on endpoint `/healthcheck`. There are no other services consumed, so it will always return OK.

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2018, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
