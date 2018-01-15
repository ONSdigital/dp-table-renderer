package parser_test

import (
	"testing"

	"encoding/json"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/parser"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestParseHTML(t *testing.T) {
	Convey("ParseHTML should create a valid RenderRequest", t, func() {

		request := models.ParseRequest{
			Filename:   "myFilename",
			Title:      "myTitle",
			Subtitle:   "mySubtitle",
			Source:     "mySource",
			URI:        "myURI",
			StyleClass: "myStyleClass",
			Footnotes:  []string{"Note0", "Note1"},
			TableHTML:  "<table></table>"}

		resultBytes, err := parser.ParseHTML(&request)

		So(err, ShouldBeNil)
		So(resultBytes, ShouldNotBeNil)

		result := parser.ResponseModel{}
		err = json.Unmarshal(resultBytes, &result)
		So(err, ShouldBeNil)
		So(result, ShouldNotBeNil)
		So(result.JSON, ShouldNotBeNil)
		So(result.JSON.Filename, ShouldEqual, request.Filename)
		So(result.JSON.Title, ShouldEqual, request.Title)
		So(result.JSON.Subtitle, ShouldEqual, request.Subtitle)
		So(result.JSON.Source, ShouldEqual, request.Source)
		So(result.JSON.URI, ShouldEqual, request.URI)
		So(result.JSON.StyleClass, ShouldEqual, request.StyleClass)
		So(result.JSON.TableType, ShouldEqual, "generated-table")
		So(result.JSON.Footnotes, ShouldResemble, request.Footnotes)
		So(result.PreviewHTML, ShouldNotBeNil)

		nodes, err := html.ParseFragment(strings.NewReader(result.PreviewHTML), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		// PreviewHTML should contain a div that contains a table
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)
		node := nodes[0]
		So(node.DataAtom, ShouldEqual, atom.Div)
		So(htmlutil.FindNode(node, atom.Table), ShouldNotBeNil)
	})

	Convey("ParseHTML should parse the cells correctly, ignoring thead", t, func() {
		response := invokeParseHTML("<table>"+
			"<thead><tr><th>A</th><th>B</th><th>C</th></tr></thead>"+
			"<tbody>"+
			"<tr><td>r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td>r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td>r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 3)
		for r, row := range data {
			So(len(row), ShouldEqual, 3)
			for c, cell := range row {
				So(cell, ShouldEqual, fmt.Sprintf("r%dc%d", r, c))
			}
		}
	})

	Convey("ParseHTML should parse the cells correctly, including thead", t, func() {
		response := invokeParseHTML("<table>"+
			"<thead><tr><th>A</th><th>B</th><th>C</th></tr></thead>"+
			"<tbody>"+
			"<tr><td>r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td>r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td>r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", true, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 4)
	})

	Convey("ParseHTML should parse the cells correctly where there is no tbody element", t, func() {
		response := invokeParseHTML("<table>"+
			"<tr><td>r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td>r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td>r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</table>", false, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 3)
		for r, row := range data {
			So(len(row), ShouldEqual, 3)
			for c, cell := range row {
				So(cell, ShouldEqual, fmt.Sprintf("r%dc%d", r, c))
			}
		}
	})

	Convey("ParseHTML should parse td cells correctly", t, func() {
		response := invokeParseHTML("<table>"+
			"<tr><th>r0c0</th><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><th>r1c0</th><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><th>r2c0</th><td>r2c1</td><td>r2c2</td></tr>"+
			"</table>", false, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 3)
		for r, row := range data {
			So(len(row), ShouldEqual, 3)
			for c, cell := range row {
				So(cell, ShouldEqual, fmt.Sprintf("r%dc%d", r, c))
			}
		}
	})

	Convey("ParseHTML should return an error if the request contains invalid html", t, func() {
		request := models.ParseRequest{Filename: "myFilename", TableHTML: "<table"}

		result, err := parser.ParseHTML(&request)

		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("ParseHTML should return an error if the request does not contain a table", t, func() {
		request := models.ParseRequest{Filename: "myFilename", TableHTML: "<div></div>"}

		result, err := parser.ParseHTML(&request)

		So(result, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

}

func invokeParseHTML(requestTable string, includeThead bool, headerRows int, headerCols int) *parser.ResponseModel {
	request := models.ParseRequest{
		Filename:     "myFilename",
		Title:        "myTitle",
		Subtitle:     "mySubtitle",
		Source:       "mySource",
		URI:          "myURI",
		StyleClass:   "myStyleClass",
		Footnotes:    []string{"Note0", "Note1"},
		TableHTML:    requestTable,
		IncludeThead: includeThead,
		HeaderRows:   headerRows,
		HeaderCols:   headerCols}

	resultBytes, err := parser.ParseHTML(&request)

	So(err, ShouldBeNil)
	So(resultBytes, ShouldNotBeNil)

	result := parser.ResponseModel{}
	err = json.Unmarshal(resultBytes, &result)
	So(err, ShouldBeNil)
	So(result, ShouldNotBeNil)

	return &result
}
