package models

import (
	"fmt"
	"strings"
	"testing"

	"bytes"

	"github.com/ONSdigital/dp-table-renderer/testdata"
	. "github.com/smartystreets/goconvey/convey"
)

// A Mock io.reader to trigger errors on reading
type reader struct {
}

func (f reader) Read(bytes []byte) (int, error) {
	return 0, fmt.Errorf("Reader failed")
}

func TestCreateRenderRequestWithValidJSON(t *testing.T) {
	Convey("When a render request has a minimally valid json body, a valid struct is returned", t, func() {
		reader := strings.NewReader(`{"title":"table_title", "filename":"filename"}`)
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "table_title")
		So(request.Filename, ShouldEqual, "filename")
	})

	Convey("When a render request has a valid json body, a valid struct is returned", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "This is an example table")
		So(request.TableType, ShouldEqual, "table")
		So(len(request.RowFormats), ShouldEqual, 2)
		So(len(request.ColumnFormats), ShouldEqual, 2)
		So(len(request.CellFormats), ShouldEqual, 3)
		So(len(request.Data), ShouldEqual, 14)
		So(len(request.Footnotes), ShouldEqual, 4)
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

func TestCreateRenderRequestWithInvalidJSON(t *testing.T) {
	Convey("When a render request contains json with an invalid syntax, and error is returned", t, func() {
		_, err := CreateRenderRequest(strings.NewReader(`{"foo`))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorParsingBody)
	})
	Convey("When a render request contains json with missing required fields, validation fails", t, func() {
		request, err := CreateRenderRequest(strings.NewReader(`{"title":"foo"}`))

		So(err, ShouldBeNil)
		err = request.ValidateRenderRequest()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "filename")
	})
}

func TestCreateParseRequestWithValidJSON(t *testing.T) {
	Convey("When a parse request has a minimally valid json body, a valid struct is returned", t, func() {
		reader := strings.NewReader(`{"table_html":"<table></table>", "filename":"filename"}`)
		request, err := CreateParseRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateParseRequest(), ShouldBeNil)
		So(request.TableHTML, ShouldEqual, "<table></table>")
		So(request.Filename, ShouldEqual, "filename")
	})

	Convey("When a parse request has a valid json body, a valid struct is returned", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleHandsonTable(t))
		request, err := CreateParseRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateParseRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "This is an example table")
		So(request.HeaderCols, ShouldEqual, 1)
		So(request.HeaderRows, ShouldEqual, 1)
		So(len(request.TableHTML), ShouldBeGreaterThan, 1)
		So(len(request.Footnotes), ShouldEqual, 2)
	})

}

func TestCreateParseRequestWithNoBody(t *testing.T) {
	Convey("When a parse request has no body, an error is returned", t, func() {
		_, err := CreateParseRequest(reader{})
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrorReadingBody)
	})

	Convey("When a parse request has an empty body, an error is returned", t, func() {
		filter, err := CreateParseRequest(strings.NewReader("{}"))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorNoData)
		So(filter, ShouldNotBeNil)
	})
}

func TestCreateParseRequestWithInvalidJSON(t *testing.T) {
	Convey("When a parse request contains json with an invalid syntax, and error is returned", t, func() {
		_, err := CreateParseRequest(strings.NewReader(`{"foo`))
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrorParsingBody)
	})
	Convey("When a parse request contains json with missing required fields, validation fails", t, func() {
		request, err := CreateParseRequest(strings.NewReader(`{"title":"foo"}`))

		So(err, ShouldBeNil)
		err = request.ValidateParseRequest()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "filename")
	})
}
