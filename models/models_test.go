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
		reader := strings.NewReader(`{"title":"table_title", "type":"table_type"}`)
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "table_title")
		So(request.TableType, ShouldEqual, "table_type")
	})

	Convey("When a render request has a valid json body, a valid struct is returned", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		request, err := CreateRenderRequest(reader)

		So(err, ShouldBeNil)
		So(request.ValidateRenderRequest(), ShouldBeNil)
		So(request.Title, ShouldEqual, "This is an example table")
		So(request.TableType, ShouldEqual, "new-table")
		So(len(request.RowFormats), ShouldEqual, 1)
		So(len(request.ColumnFormats), ShouldEqual, 2)
		So(len(request.CellFormats), ShouldEqual, 3)
		So(len(request.Data), ShouldEqual, 14)
		So(len(request.Footnotes), ShouldEqual, 3)
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
