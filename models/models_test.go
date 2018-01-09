package models

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// A Mock io.reader to trigger errors on reading
type reader struct {
}

var exampleRequest = `{
  "filename": "abcd1234",
  "title": "This is an example table",
  "subtitle": "with a subtitle",
  "type": "new-table",
  "uri": "/path/to/the/table/json",
  "style_class": "foo bar",
  "row_formats": [
    {"row": 0, "style_class": "emphasis", "style": "text-color: blue", "heading": true}
  ],
  "column_formats": [
    {"col": 0, "align": "right", "style": "width: 20em", "heading": true},
    {"col": 1, "align": "right", "style_class": "emphasis", "style": "width: 20em", "heading": true}
  ],
  "cell_formats": [
    {"row": 0, "col": 0, "align": "middle", "colspan": 2},
    {"row": 1, "col": 0, "align": "left", "vertical_align": "top", "rowspan": 2},
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
    "And can be referenced from any data element or title using square brackets: [1]"
    ]

}`

func (f reader) Read(bytes []byte) (int, error) {
	return 0, fmt.Errorf("Reader failed")
}

func TestCreateRenderRequestWithValidJSON(t *testing.T) {
	Convey("When a render request has a minimally valid json body, a valid struct is returned", t, func() {
		reader := strings.NewReader(`{"title":"table_title", "type":"table_type"}`)
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "table_title")
		So(request.TableType, ShouldEqual, "table_type")
	})

	Convey("When a render request has a valid json body, a valid struct is returned", t, func() {
		reader := strings.NewReader(exampleRequest)
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "This is an example table")
		So(request.TableType, ShouldEqual, "new-table")
		So(len(request.RowFormats), ShouldEqual, 1)
		So(len(request.ColumnFormats), ShouldEqual, 2)
		So(len(request.CellFormats), ShouldEqual, 3)
		So(len(request.Data), ShouldEqual, 14)
		So(len(request.Footnotes), ShouldEqual, 2)
	})

}

func TestCreateRenderRequestWithNoBody(t *testing.T) {
	Convey("When a render request has no body, an error is returned", t, func() {
		_, err := CreateRenderRequest(reader{})
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorReadingBody)
	})

	Convey("When a render request has an empty body, an error is returned", t, func() {
		filter, err := CreateRenderRequest(strings.NewReader("{}"))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorNoData)
		So(filter, ShouldNotBeNil)
	})
}

func TestCreateRenderRequestBlueprintWithInvalidJSON(t *testing.T) {
	Convey("When a render request contains json with an invalid syntax, and error is returned", t, func() {
		_, err := CreateRenderRequest(strings.NewReader(`{"foo`))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorParsingBody)
	})
}
