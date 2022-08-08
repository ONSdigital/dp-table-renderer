package renderer_test

import (
	"bytes"
	"testing"

	"fmt"

	"strings"

	. "github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/ONSdigital/dp-table-renderer/testdata"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const footnoteLinkClass = "footnote__link"

func TestRenderHTML(t *testing.T) {

	Convey("Successfully render an html table", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		renderRequest, err := models.CreateRenderRequest(mockContext, reader)
		if err != nil {
			t.Fatal(err)
		}

		container, responseHTML := invokeRenderHTML(renderRequest)

		So(GetAttribute(container, "class"), ShouldEqual, "figure")
		So(GetAttribute(container, "id"), ShouldEqual, "table-"+renderRequest.Filename)

		// the table
		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)
		So(GetAttribute(table, "class"), ShouldEqual, "table")
		// with caption
		So(FindNode(table, atom.Caption), ShouldNotBeNil)
		// and correct number of rows
		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(renderRequest.Data))

		// the footer - source
		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		// footnotes
		notes := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__notes"})
		So(notes, ShouldNotBeNil)
		So(notes.FirstChild.Data, ShouldResemble, "Notes")
		footnotes := FindNodes(footer, atom.Li)
		So(len(footnotes), ShouldEqual, len(renderRequest.Footnotes))

		// new line characters are converted to <br/> tags
		So(responseHTML, ShouldContainSubstring, "CPIH 12-<br/>month rate")

		// Anchor tags in cells are not stripped
		So(responseHTML, ShouldContainSubstring, "<a href=\"cell-link\">link</a>")

		// Anchor tags in caption and footer are stripped
		So(responseHTML, ShouldNotContainSubstring, "<a href=\"foot-link\">")
		So(responseHTML, ShouldNotContainSubstring, "<a href=\"title-link\">")
		So(responseHTML, ShouldNotContainSubstring, "<a href=\"subtitle-link\">")
	})
}

func TestRenderHTML_Table(t *testing.T) {

	Convey("A table should have title and subtitle in the caption", t, func() {
		request := models.RenderRequest{Filename: "filename", Title: "Heading", Subtitle: "Subtitle"}
		container, _ := invokeRenderHTML(&request)

		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)
		So(GetAttribute(table, "id"), ShouldBeEmpty)

		caption := FindNode(table, atom.Caption)
		So(caption, ShouldNotBeNil)
		So(GetAttribute(caption, "class"), ShouldEqual, "table__caption")
		So(caption.FirstChild.Data, ShouldEqual, "Heading")
		span := FindNode(caption, atom.Span)
		So(span, ShouldNotBeNil)
		So(span.FirstChild.Data, ShouldEqual, "Subtitle")
		So(GetAttribute(span, "class"), ShouldEqual, "table__subtitle")
	})

	Convey("A table without title or subtitle should not have a caption", t, func() {
		request := models.RenderRequest{Filename: "myId"}
		container, _ := invokeRenderHTML(&request)

		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)
		So(GetAttribute(table, "id"), ShouldBeEmpty)
		So(FindNode(table, atom.Caption), ShouldBeNil)
	})

	Convey("A table with unbalanced cell counts is still rendered", t, func() {
		cells := [][]string{
			{"Cell 1", "Cell 2", "Cell 3"},
			{"Cell 1", "Cell 2", "Cell 3", "Cell 4"},
			{"Cell 1", "Cell 2", "Cell 3", "Cell 4", "Cell 5"}}
		formats := []models.ColumnFormat{{Column: 0, Width: "10em"}}
		request := models.RenderRequest{Filename: "myId", Data: cells, ColumnFormats: formats}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(request.Data))
		for i, row := range rows {
			So(len(FindNodes(row, atom.Td)), ShouldEqual, len(cells[i]))
		}
	})
}

func TestRenderHTML_Source(t *testing.T) {

	Convey("A renderRequest without a source should not have a source paragraph", t, func() {
		request := models.RenderRequest{Filename: "myId"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		So(FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__source"}), ShouldBeNil)
	})

	Convey("A renderRequest with a source should have a source paragraph", t, func() {
		request := models.RenderRequest{Filename: "myId", Source: "mySource"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		source := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__source"})
		So(source, ShouldNotBeNil)
		So(source.FirstChild.Data, ShouldResemble, "Source: "+request.Source)
	})
}

func TestRenderHTML_Units(t *testing.T) {

	Convey("A renderRequest without units should not have a units paragraph", t, func() {
		request := models.RenderRequest{Filename: "myId"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		So(FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__units"}), ShouldBeNil)
	})

	Convey("A renderRequest with a source should have a source paragraph", t, func() {
		request := models.RenderRequest{Filename: "myId", Units: "myUnits"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		units := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__units"})
		So(units, ShouldNotBeNil)
		So(units.FirstChild.Data, ShouldResemble, "Units: "+request.Units)
	})
}

func TestRenderHTML_Footer(t *testing.T) {
	Convey("A renderRequest without footnotes should not have notes paragraph", t, func() {
		request := models.RenderRequest{Filename: "myId"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		So(GetAttribute(footer, "class"), ShouldEqual, "figure__footer")
		So(FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__notes"}), ShouldBeNil)
		So(len(FindNodes(footer, atom.Li)), ShouldBeZeroValue)
	})

	Convey("Footnotes should render as li elements with id", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)

		p := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "figure__notes"})
		So(p, ShouldNotBeNil)
		So(p.FirstChild.Data, ShouldResemble, "Notes")

		list := FindNode(footer, atom.Ol)
		So(list, ShouldNotBeNil)
		So(GetAttribute(list, "class"), ShouldEqual, "figure__footnotes")
		notes := FindNodes(list, atom.Li)
		So(len(notes), ShouldEqual, len(request.Footnotes))
		for i, note := range request.Footnotes {
			So(GetAttribute(notes[i], "id"), ShouldEqual, fmt.Sprintf("table-%s-note-%d", request.Filename, i+1))
			So(GetAttribute(notes[i], "class"), ShouldEqual, "figure__footnote-item")
			So(strings.Trim(notes[i].FirstChild.Data, " "), ShouldResemble, note)
		}
	})

	Convey("Footnotes should be properly parsed", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2\nOn Two Lines"}}
		_, result := invokeRenderHTML(&request)

		So(result, ShouldContainSubstring, "Note2<br/>On Two Lines")
	})
}

func TestRenderHTML_FootnoteLinks(t *testing.T) {

	Convey("A renderRequest with references to footnotes should convert those to links", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}, Data: [][]string{{"Cell 1[1]", "Cell[2] 2[1]"}, {"Cell 3[3]", "Cell[0][]"}}}
		container, raw := invokeRenderHTML(&request)

		links := FindNodesWithAttributes(container, atom.A, map[string]string{"class": footnoteLinkClass})
		So(len(links), ShouldEqual, 3)
		for _, link := range links {
			span := FindNode(link, atom.Span)
			So(GetAttribute(span, "class"), ShouldEqual, "visuallyhidden")
			So(GetText(span), ShouldEqual, "Footnote ")
		}
		So(GetAttribute(links[0], "href"), ShouldEqual, "#table-myId-note-1")
		So(GetAttribute(links[1], "href"), ShouldEqual, "#table-myId-note-2")
		So(GetAttribute(links[2], "href"), ShouldEqual, "#table-myId-note-1")

		p := FindNodeWithAttributes(container, atom.P, map[string]string{"class": "figure__notes"})
		So(p, ShouldNotBeNil)

		So(raw, ShouldNotContainSubstring, "Cell 1[1]")
		So(raw, ShouldNotContainSubstring, "Cell[2] 2[1]")
		So(raw, ShouldContainSubstring, "Cell 3[3]")
		So(raw, ShouldContainSubstring, "Cell[0][]")
	})

	Convey("A renderRequest with lots of footnotes (>10) is handled correctly", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, Data: [][]string{{"Cell [11]"}}}
		container, _ := invokeRenderHTML(&request)

		links := FindNodesWithAttributes(container, atom.A, map[string]string{"class": footnoteLinkClass})
		So(len(links), ShouldEqual, 1)
		So(GetAttribute(links[0], "href"), ShouldEqual, "#table-myId-note-11")
	})

	Convey("Multiple references to the same footnote in the same value should all be converted to links", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}, Title: "This contains [1] links[1]"}
		container, _ := invokeRenderHTML(&request)

		links := FindNodesWithAttributes(container, atom.A, map[string]string{"class": footnoteLinkClass})
		So(len(links), ShouldEqual, 2)
		for _, link := range links {
			So(GetAttribute(link, "href"), ShouldEqual, "#table-myId-note-1")
		}
	})
}

func TestRenderHTML_ColumnFormats(t *testing.T) {

	Convey("A renderRequest with column formats should output colgroup", t, func() {
		formats := []models.ColumnFormat{{Column: 0, Width: "10em"}, {Column: 2, Align: models.AlignRight}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: formats,
			Data: [][]string{
				{"Cell 1", "Cell 2", "Cell 3", "Cell 4"},
				{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		colgroup := FindNode(table, atom.Colgroup)
		So(colgroup, ShouldNotBeNil)
		cols := FindNodes(colgroup, atom.Col)
		So(len(cols), ShouldEqual, len(request.Data[0]))
		So(GetAttribute(cols[0], "style"), ShouldEqual, "width: 10em")
		So(GetAttribute(cols[1], "style"), ShouldBeEmpty)
		So(GetAttribute(cols[2], "class"), ShouldBeEmpty)
		So(GetAttribute(cols[3], "class"), ShouldBeEmpty)

		rows := FindNodes(table, atom.Tr)
		for _, row := range rows {
			cells := FindNodes(row, atom.Td)
			So(len(cells), ShouldEqual, len(request.Data[0]))
			So(GetAttribute(cells[2], "class"), ShouldContainSubstring, "right")
		}
	})

	Convey("If there are no column formats then there should be no colgroup element", t, func() {
		request := models.RenderRequest{Filename: "myId", Data: [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		So(FindNode(table, atom.Colgroup), ShouldBeNil)
	})

	Convey("Columns flagged as headers should create scoped th elements in each row", t, func() {
		formats := []models.ColumnFormat{{Column: 0, Heading: true}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: formats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		for _, row := range rows {
			header := FindNode(row, atom.Th)
			So(header, ShouldNotBeNil)
			So(GetAttribute(header, "scope"), ShouldEqual, "row")
			So(header.FirstChild.Data, ShouldResemble, "Cell 1")
			So(len(FindNodes(row, atom.Td)), ShouldEqual, 3)
		}
	})

	Convey("Columns flagged as headers should have the correct scope when in a header row", t, func() {
		colFormats := []models.ColumnFormat{{Column: 0, Heading: true}}
		rowFormats := []models.RowFormat{{Row: 0, Heading: true}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: colFormats, RowFormats: rowFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		headers := FindNodes(rows[0], atom.Th)
		So(len(headers), ShouldEqual, len(cells[0]))
		for _, col := range headers {
			So(GetAttribute(col, "scope"), ShouldEqual, "col")
		}
		rowHeaders := FindNodes(rows[1], atom.Th)
		So(len(rowHeaders), ShouldEqual, 1)
		So(GetAttribute(rowHeaders[0], "scope"), ShouldEqual, "row")
	})

	Convey("Columns with colspan flagged as headers should have the correct scope when in a header row", t, func() {
		colFormats := []models.ColumnFormat{{Column: 0, Heading: true}}
		rowFormats := []models.RowFormat{{Row: 0, Heading: true}}
		cellFormats := []models.CellFormat{{Row: 0, Column: 0, Colspan: 2}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: colFormats, RowFormats: rowFormats, CellFormats: cellFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		headers := FindNodes(rows[0], atom.Th)
		So(len(headers), ShouldEqual, len(cells[0])-1)
		So(GetAttribute(headers[0], "scope"), ShouldEqual, "colgroup")
		So(GetAttribute(headers[1], "scope"), ShouldEqual, "col")

		rowHeaders := FindNodes(rows[1], atom.Th)
		So(len(rowHeaders), ShouldEqual, 1)
		So(GetAttribute(rowHeaders[0], "scope"), ShouldEqual, "row")
	})

	Convey("Column formats beyond the count of columns are ignored", t, func() {
		formats := []models.ColumnFormat{{Column: -1, Width: "5em"}, {Column: 5, Width: "5em"}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: formats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		colgroup := FindNode(table, atom.Colgroup)
		So(colgroup, ShouldNotBeNil)
		cols := FindNodes(colgroup, atom.Col)
		So(len(cols), ShouldEqual, len(request.Data[0]))
		for _, col := range cols {
			So(GetAttribute(col, "style"), ShouldBeEmpty)
		}
	})

	Convey("Column that have no content shold not become headers but instead td", t, func() {
		colFormats := []models.ColumnFormat{{Column: 0, Heading: true}}
		rowFormats := []models.RowFormat{{Row: 0, Heading: true}}
		cells := [][]string{{"", "Head 1", "Head 2", "Head 3"}, {"Head 4", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: colFormats, RowFormats: rowFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		headers := FindNodes(rows[0], atom.Th)
		emptyHeaders := FindNodes(rows[0], atom.Td)
		So(len(headers), ShouldEqual, len(cells[0])-1)
		So(len(emptyHeaders), ShouldEqual, 1)
	})
}

func TestRenderHTML_Rows(t *testing.T) {

	Convey("Rows flagged as headers should have the correct class", t, func() {
		rowFormats := []models.RowFormat{{Row: 0, Heading: true}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", RowFormats: rowFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		So(GetAttribute(rows[0], "class"), ShouldContainSubstring, "table__header-row")
		So(GetAttribute(rows[1], "class"), ShouldNotContainSubstring, "table__header-row")
	})

	Convey("Rows flagged as headers with vertical alignment should have the correct classes", t, func() {
		rowFormats := []models.RowFormat{{Row: 0, Heading: true, VerticalAlign: "Top"}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", RowFormats: rowFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		class := GetAttribute(rows[0], "class")
		So(class, ShouldContainSubstring, "table__header-row")
		So(class, ShouldContainSubstring, "align-top")
	})

}

func TestRenderHTML_MergeCells(t *testing.T) {

	Convey("A renderRequest with merged cells should have the correct number of cells", t, func() {
		cellFormats := []models.CellFormat{
			{Row: 0, Column: 0, Colspan: 2, Rowspan: 2},
			{Row: 0, Column: 3, Colspan: 2},
			{Row: 3, Column: 3, Rowspan: 2}}
		cells := [][]string{
			{"0A", "0B", "0C", "0D", "0E"},
			{"1A", "1B", "1C", "1D", "1E"},
			{"2A", "2B", "2C", "2D", "2E"},
			{"3A", "3B", "3C", "3D", "3E"},
			{"4A", "4B", "4C", "4D", "4E"}}
		request := models.RenderRequest{Filename: "myId", CellFormats: cellFormats, Data: cells}
		container, raw := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		So(len(FindNodes(rows[0], atom.Td)), ShouldEqual, 3)
		So(len(FindNodes(rows[1], atom.Td)), ShouldEqual, 3)
		So(len(FindNodes(rows[2], atom.Td)), ShouldEqual, 5)
		So(len(FindNodes(rows[3], atom.Td)), ShouldEqual, 5)
		So(len(FindNodes(rows[4], atom.Td)), ShouldEqual, 4)

		So(raw, ShouldNotContainSubstring, "0B")
		So(raw, ShouldNotContainSubstring, "0E")
		So(raw, ShouldNotContainSubstring, "1A")
		So(raw, ShouldNotContainSubstring, "1B")
		So(raw, ShouldNotContainSubstring, "4D")
	})
}

func TestRenderHTML_ColumnAndRowAlignment(t *testing.T) {

	Convey("A renderRequest with various alignments should have correct classes", t, func() {
		rowFormats := []models.RowFormat{{Row: 0, VerticalAlign: models.AlignTop}}
		colFormats := []models.ColumnFormat{{Column: 0, Align: models.AlignRight}}
		cellFormats := []models.CellFormat{{Row: 0, Column: 0, VerticalAlign: models.AlignBottom},
			{Row: 1, Column: 0, Align: models.AlignJustify}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", ColumnFormats: colFormats, RowFormats: rowFormats, CellFormats: cellFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		colgroup := FindNode(table, atom.Colgroup)
		So(colgroup, ShouldNotBeNil)
		cols := FindNodes(colgroup, atom.Col)
		So(len(cols), ShouldEqual, len(request.Data[0]))

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		So(GetAttribute(rows[0], "class"), ShouldContainSubstring, "align-top")

		td := FindNodes(rows[0], atom.Td)
		So(len(td), ShouldEqual, len(request.Data[0]))
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "align-bottom")
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "align-right")
		So(GetAttribute(td[1], "class"), ShouldBeEmpty)

		td = FindNodes(rows[1], atom.Td)
		So(len(td), ShouldEqual, len(request.Data[1]))
		So(GetAttribute(td[0], "class"), ShouldNotContainSubstring, "align-bottom")
		So(GetAttribute(td[0], "class"), ShouldNotContainSubstring, "align-right")
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "align-justify")
		So(GetAttribute(td[1], "class"), ShouldBeEmpty)
	})
}

func TestRenderHTML_RowHeight(t *testing.T) {

	Convey("A renderRequest with row height should have correct style", t, func() {
		rowFormats := []models.RowFormat{{Row: 0, Height: "5em"}}
		cells := [][]string{{"Cell 1", "Cell 2", "Cell 3", "Cell 4"}, {"Cell 1", "Cell 2", "Cell 3", "Cell 4"}}
		request := models.RenderRequest{Filename: "myId", RowFormats: rowFormats, Data: cells}
		container, _ := invokeRenderHTML(&request)
		table := FindNode(container, atom.Table)

		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(cells))
		So(GetAttribute(rows[0], "style"), ShouldEqual, "height: 5em")
		So(GetAttribute(rows[1], "style"), ShouldBeEmpty)

	})
}

func TestRenderHTML_KeepHeadersTogether(t *testing.T) {

	Convey("Given a request with row & column headers and KeepHeadersTogether = true", t, func() {
		request := models.RenderRequest{Filename: "myId",
			RowFormats:          []models.RowFormat{{Row: 0, Heading: true}},
			ColumnFormats:       []models.ColumnFormat{{Column: 0, Heading: true}},
			KeepHeadersTogether: true,
			Data: [][]string{
				{"Cell 1", "Cell 2", "Cell 3", "Cell 4"},
				{"Cell 1", "Cell 2", "Cell 3", "Cell 4"},
			}}
		Convey("When rederHTML is invoked", func() {
			container, _ := invokeRenderHTML(&request)
			table := FindNode(container, atom.Table)
			rows := FindNodes(table, atom.Tr)

			Convey("Then the first row should have the correct css class", func() {
				So(GetAttribute(rows[0], "class"), ShouldContainSubstring, "table__nowrap")
			})

			Convey("The first column of remaining rows should have the correct css class", func() {
				So(GetAttribute(rows[1], "class"), ShouldNotContainSubstring, "table__nowrap")
				colHead := FindNode(rows[1], atom.Th)
				So(GetAttribute(colHead, "class"), ShouldContainSubstring, "table__nowrap")
				for _, td := range FindNodes(rows[1], atom.Td) {
					So(GetAttribute(td, "class"), ShouldNotContainSubstring, "table__nowrap")
				}
			})

		})

	})
}

func invokeRenderHTML(renderRequest *models.RenderRequest) (*html.Node, string) {
	response, err := renderer.RenderHTML(mockContext, renderRequest)
	So(err, ShouldBeNil)
	nodes, err := html.ParseFragment(bytes.NewReader(response), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	So(err, ShouldBeNil)
	So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)
	// the containing container
	node := nodes[0]
	So(node.DataAtom, ShouldEqual, atom.Figure)
	return node, string(response)
}
