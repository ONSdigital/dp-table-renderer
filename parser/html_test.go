package parser_test

import (
	"testing"

	"encoding/json"
	"fmt"
	"strings"

	"bytes"

	"github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/parser"
	"github.com/ONSdigital/dp-table-renderer/testdata"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestParseHTML(t *testing.T) {
	Convey("ParseHTML should successfully parse the example request", t, func() {

		request, err := models.CreateParseRequest(bytes.NewReader(testdata.LoadExampleHandsonTable(t)))
		So(err, ShouldBeNil)

		resultBytes, err := parser.ParseHTML(request)

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
		So(result.JSON.Units, ShouldEqual, request.Units)
		So(result.JSON.TableType, ShouldEqual, "table")
		So(result.JSON.TableVersion, ShouldEqual, "2")
		So(result.JSON.Footnotes, ShouldResemble, request.Footnotes)
		So(result.PreviewHTML, ShouldNotBeNil)

		nodes, err := html.ParseFragment(strings.NewReader(result.PreviewHTML), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		// PreviewHTML should contain a figure that contains a table
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)
		node := nodes[0]
		So(node.DataAtom, ShouldEqual, atom.Figure)
		So(htmlutil.FindNode(node, atom.Table), ShouldNotBeNil)
	})

	Convey("ParseHTML should create a valid RenderRequest", t, func() {

		request := models.ParseRequest{
			Filename:  "myFilename",
			Title:     "myTitle",
			Subtitle:  "mySubtitle",
			Source:    "mySource",
			Footnotes: []string{"Note0", "Note1"},
			TableHTML: "<table></table>"}

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
		So(result.JSON.TableType, ShouldEqual, "table")
		So(result.JSON.TableVersion, ShouldEqual, "2")
		So(result.JSON.Footnotes, ShouldResemble, request.Footnotes)
		So(result.PreviewHTML, ShouldNotBeNil)

		nodes, err := html.ParseFragment(strings.NewReader(result.PreviewHTML), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		// PreviewHTML should contain a figure that contains a table
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)
		node := nodes[0]
		So(node.DataAtom, ShouldEqual, atom.Figure)
		So(htmlutil.FindNode(node, atom.Table), ShouldNotBeNil)
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

func TestParseHTML_Data(t *testing.T) {

	Convey("ParseHTML should parse the cells correctly, ignoring thead", t, func() {
		response := invokeParseHTML("<table>"+
			"<thead><tr><th></th><th>A</th><th>B</th><th>C</th></tr></thead>"+
			"<tbody>"+
			"<tr><th>1</th><td>r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><th>2</th><td>r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><th>3</th><td>r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", true, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 3)
		for r, row := range data {
			So(len(row), ShouldEqual, 3)
			for c, cell := range row {
				So(cell, ShouldEqual, fmt.Sprintf("r%dc%d", r, c))
			}
		}
	})

	Convey("ParseHTML should parse the cells correctly, including first row and column", t, func() {
		response := invokeParseHTML("<table>"+
			"<thead><tr><th>r0c0</th><th>r0c1</th><th>r0c2</th></tr></thead>"+
			"<tbody>"+
			"<tr><th>r1c0</th><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><th>r2c0</th><td>r2c1</td><td>r2c2</td></tr>"+
			"<tr><th>r3c0</th><td>r3c1</td><td>r3c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 0)

		data := response.JSON.Data
		So(len(data), ShouldEqual, 4)
		for r, row := range data {
			So(len(row), ShouldEqual, 3)
			for c, cell := range row {
				So(cell, ShouldEqual, fmt.Sprintf("r%dc%d", r, c))
			}
		}
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

	Convey("ParseHTML should parse th cells", t, func() {
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

}

func TestParseHTML_ColumnFormats(t *testing.T) {

	Convey("ParseHTML should create column formats with alignment, heading flags, width", t, func() {
		response := invokeParseHTML("<table>"+
			"<colgroup><col style=\"foo: bar; width: 5em\" /><col/><col/>"+
			"<tbody>"+
			"<tr><td class=\"right\">r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"top right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 2)

		formats := response.JSON.ColumnFormats
		So(len(formats), ShouldEqual, 2)
		for i, format := range formats {
			So(format.Column, ShouldEqual, i)
			So(format.Heading, ShouldBeTrue)
		}
		So(formats[0].Align, ShouldEqual, models.AlignRight)
		So(formats[0].Width, ShouldEqual, "5em")

	})

	Convey("Column width in pixels should be converted to em", t, func() {
		request := createParseRequest("<table>"+
			"<colgroup><col style=\"foo: bar; width: 60px\" /><col style=\"width: 65px;\"/><col/>"+
			"<tbody>"+
			"<tr><td class=\"right\">r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"top right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 2)

		request.SizeUnits = "em"
		request.SingleEmHeight = 15
		response := invokeParseHTMLWithRequest(request)

		formats := response.JSON.ColumnFormats
		So(len(formats), ShouldEqual, 2)
		for i, format := range formats {
			So(format.Column, ShouldEqual, i)
			So(format.Heading, ShouldBeTrue)
		}
		So(formats[0].Align, ShouldEqual, models.AlignRight)
		So(formats[0].Width, ShouldEqual, "4em")
		So(formats[1].Width, ShouldEqual, "4.33em")

	})

	Convey("Column width in pixels should be converted to percent", t, func() {
		request := createParseRequest("<table>"+
			"<colgroup><col style=\"foo: bar; width: 50px\" /><col/><col/>"+
			"<tbody>"+
			"<tr><td class=\"right\">r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"top right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 2)

		request.SizeUnits = "%"
		request.CurrentTableWidth = 200
		response := invokeParseHTMLWithRequest(request)

		formats := response.JSON.ColumnFormats
		So(len(formats), ShouldEqual, 2)
		for i, format := range formats {
			So(format.Column, ShouldEqual, i)
			So(format.Heading, ShouldBeTrue)
		}
		So(formats[0].Align, ShouldEqual, models.AlignRight)
		So(formats[0].Width, ShouldEqual, "25%")

	})

	Convey("Default column width should be ignored", t, func() {
		request := createParseRequest("<table>"+
			"<colgroup><col style=\"foo: bar; width: 60em\" /><col/><col/>"+
			"<tbody>"+
			"<tr><td class=\"right\">r0c0</td><td>r0c1</td><td>r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"top right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 2)

		request.ColumnWidthToIgnore = "60em"
		response := invokeParseHTMLWithRequest(request)

		formats := response.JSON.ColumnFormats
		So(len(formats), ShouldEqual, 2)
		for i, format := range formats {
			So(format.Column, ShouldEqual, i)
			So(format.Heading, ShouldBeTrue)
		}
		So(formats[0].Align, ShouldEqual, models.AlignRight)
		So(formats[0].Width, ShouldBeEmpty)

	})

	Convey("First col of colgroup should be ignored if IgnoreFirstColumn is true", t, func() {
		request := createParseRequest("<table>"+
			"<colgroup><col style=\"width: 50px\" /><col style=\"width: 100px\" /><col/>"+
			"<tbody>"+
			"<tr><td>r0c0</td><td class=\"right\">r0c1</td><td>r0c2</td></tr>"+
			"<tr><td>r1c0</td><td class=\"right\">r1c1</td><td>r1c2</td></tr>"+
			"<tr><td>r2c0</td><td class=\"top right\">r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 0, 0)

		request.SizeUnits = "%"
		request.CurrentTableWidth = 200
		request.IgnoreFirstColumn = true
		response := invokeParseHTMLWithRequest(request)

		formats := response.JSON.ColumnFormats
		So(len(formats), ShouldEqual, 1)
		So(formats[0].Align, ShouldEqual, models.AlignRight)
		So(formats[0].Width, ShouldEqual, "50%")

	})

}

func TestParseHTML_RowFormats(t *testing.T) {

	Convey("ParseHTML should create row formats with alignment, heading flags", t, func() {
		response := invokeParseHTML("<table>"+
			"<tbody>"+
			"<tr><td class=\"top\">r0c0</td><td class=\"top\">r0c1</td><td class=\"top\">r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td>r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"top right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 2, 0)

		formats := response.JSON.RowFormats
		So(len(formats), ShouldEqual, 2)
		for i, format := range formats {
			So(format.Row, ShouldEqual, i)
			So(format.Heading, ShouldBeTrue)
		}
		So(formats[0].VerticalAlign, ShouldEqual, models.AlignTop)

	})

}

func TestParseHTML_CellFormats(t *testing.T) {

	Convey("ParseHTML should create cell formats with alignment, rowspan and colspan", t, func() {
		request := createParseRequest("<table>"+
			"<tbody>"+
			"<tr><td class=\"top right\">r0c0</td><td class=\"top\">r0c1</td><td class=\"top\">r0c2</td></tr>"+
			"<tr><td class=\"right\">r1c0</td><td colspan=\"2\" rowspan=\"2\">r1c1</td><td>r1c2</td></tr>"+
			"<tr><td class=\"right\">r2c0</td><td>r2c1</td><td>r2c2</td></tr>"+
			"<tr><td class=\"top right\">r3c0</td><td colspan=\"2\">r3c1</td><td>r3c2</td></tr>"+
			"</tbody>"+
			"</table>", false, 2, 0)

		response := invokeParseHTMLWithRequest(request)

		formats := response.JSON.CellFormats
		So(len(formats), ShouldEqual, 3)

		So(getCellFormat(formats, 0, 0), ShouldBeNil)

		format := getCellFormat(formats, 1, 1)
		So(format, ShouldNotBeNil)
		So(format.Colspan, ShouldEqual, 2)
		So(format.Rowspan, ShouldEqual, 2)
		So(format.Align, ShouldBeEmpty)
		So(format.VerticalAlign, ShouldBeEmpty)

		format = getCellFormat(formats, 3, 0)
		So(format, ShouldNotBeNil)
		So(format.Colspan, ShouldEqual, 0)
		So(format.VerticalAlign, ShouldEqual, models.AlignTop)

		format = getCellFormat(formats, 3, 1)
		So(format, ShouldNotBeNil)
		So(format.Colspan, ShouldEqual, 2)
		So(format.Rowspan, ShouldEqual, 0)
	})

	Convey("ParseHTML should not create formats when no formatting is present in the source table", t, func() {
		request := models.ParseRequest{
			Filename: "abcd1234",
			TableHTML: `<table class="htCore"><tbody>
						<tr><td class=""></td><td class="">a</td><td class="">b</td><td class="">c</td><td class="">d</td><td class="">e</td></tr>
						<tr><td class="">2016</td><td class="">10</td><td class="">11</td><td class="">12</td><td class="">13</td><td class="">14</td></tr>
					</tbody></table>`,
		}

		response := invokeParseHTMLWithRequest(&request)

		So(len(response.JSON.CellFormats), ShouldBeZeroValue)
		So(len(response.JSON.RowFormats), ShouldBeZeroValue)
		So(len(response.JSON.ColumnFormats), ShouldBeZeroValue)
	})

}

func getCellFormat(formats []models.CellFormat, row int, col int) *models.CellFormat {
	for _, format := range formats {
		if format.Row == row && format.Column == col {
			return &format
		}
	}
	return nil
}

func invokeParseHTML(requestTable string, hasHeaders bool, headerRows int, headerCols int) *parser.ResponseModel {
	request := createParseRequest(requestTable, hasHeaders, headerRows, headerCols)

	return invokeParseHTMLWithRequest(request)
}

func createParseRequest(requestTable string, hasHeaders bool, headerRows int, headerCols int) *models.ParseRequest {
	request := models.ParseRequest{
		Filename:          "myFilename",
		Title:             "myTitle",
		Subtitle:          "mySubtitle",
		Source:            "mySource",
		Footnotes:         []string{"Note0", "Note1"},
		TableHTML:         requestTable,
		IgnoreFirstRow:    hasHeaders,
		IgnoreFirstColumn: hasHeaders,
		HeaderRows:        headerRows,
		HeaderCols:        headerCols,
		AlignmentClasses: models.ParseAlignments{
			Top:    "top",
			Middle: "middle",
			Bottom: "bottom",
			Left:   "left",
			Center: "center",
			Right:  "right",
		}}
	return &request
}

func invokeParseHTMLWithRequest(request *models.ParseRequest) *parser.ResponseModel {
	resultBytes, err := parser.ParseHTML(request)

	So(err, ShouldBeNil)
	So(resultBytes, ShouldNotBeNil)

	result := parser.ResponseModel{}
	err = json.Unmarshal(resultBytes, &result)
	So(err, ShouldBeNil)
	So(result, ShouldNotBeNil)

	return &result

}
