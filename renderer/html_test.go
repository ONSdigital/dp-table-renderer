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
	t.Parallel()
	Convey("Successfully render an html table", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		renderRequest, err := models.CreateRenderRequest(reader)
		if err != nil {
			t.Fatal(err)
		}

		container, responseHTML := invokeRenderHTML(renderRequest)

		So(GetAttribute(container, "class"), ShouldEqual, "figure__table")
		So(GetAttribute(container, "id"), ShouldEqual, "table_"+renderRequest.Filename)

		// the table
		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)
		// with caption
		So(FindNode(table, atom.Caption), ShouldNotBeNil)
		// and correct number of rows
		rows := FindNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, len(renderRequest.Data))

		// the footer - source
		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		source := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "table-source"})
		So(source, ShouldNotBeNil)
		So(source.FirstChild.Data, ShouldResemble, "Source: "+renderRequest.Source)
		// footnotes
		notes := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "table-notes"})
		So(notes, ShouldNotBeNil)
		So(notes.FirstChild.Data, ShouldResemble, "Notes")
		footnotes := FindNodes(footer, atom.Li)
		So(len(footnotes), ShouldEqual, len(renderRequest.Footnotes))

		// new line characters are converted to <br/> tags
		So(responseHTML, ShouldContainSubstring, "CPIH 12-<br/>month rate")
	})
}

func invokeRenderHTML(renderRequest *models.RenderRequest) (*html.Node, string) {
	response, err := renderer.RenderHTML(renderRequest)
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

func TestRenderHTML_Table(t *testing.T) {
	t.Parallel()
	Convey("A table should have title and subtitle in the caption", t, func() {
		request := models.RenderRequest{Filename: "filename", Title: "Heading", Subtitle: "Subtitle"}
		container, _ := invokeRenderHTML(&request)

		table := FindNode(container, atom.Table)
		So(table, ShouldNotBeNil)
		So(GetAttribute(table, "id"), ShouldBeEmpty)

		caption := FindNode(table, atom.Caption)
		So(caption, ShouldNotBeNil)
		So(caption.FirstChild.Data, ShouldEqual, "Heading")
		span := FindNode(caption, atom.Span)
		So(span, ShouldNotBeNil)
		So(span.FirstChild.Data, ShouldEqual, "Subtitle")
		So(GetAttribute(span, "class"), ShouldEqual, "caption__subtitle")
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

func TestRenderHTML_Footer(t *testing.T) {
	Convey("A renderRequest without a source or footnotes should not have source or notes paragraphs", t, func() {
		request := models.RenderRequest{Filename: "myId"}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)
		So(FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "table-source"}), ShouldBeNil)
		So(FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "table-notes"}), ShouldBeNil)
		So(len(FindNodes(footer, atom.Li)), ShouldBeZeroValue)
	})

	Convey("Footnotes should render as li elements with id", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}}
		container, _ := invokeRenderHTML(&request)

		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)

		p := FindNodeWithAttributes(footer, atom.P, map[string]string{"class": "table-notes"})
		So(p, ShouldNotBeNil)
		So(p.FirstChild.Data, ShouldResemble, "Notes")

		list := FindNode(footer, atom.Ol)
		So(list, ShouldNotBeNil)
		notes := FindNodes(list, atom.Li)
		So(len(notes), ShouldEqual, len(request.Footnotes))
		for i, note := range request.Footnotes {
			So(GetAttribute(notes[i], "id"), ShouldEqual, fmt.Sprintf("table_%s_note_%d", request.Filename, i+1))
			So(strings.Trim(notes[i].FirstChild.Data, " "), ShouldResemble, note)
		}
	})

	Convey("Footnotes should end with a link back to the top of the table", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}}
		container, _ := invokeRenderHTML(&request)
		footer := FindNode(container, atom.Footer)
		So(footer, ShouldNotBeNil)

		notes := FindNodes(footer, atom.Li)
		So(len(notes), ShouldEqual, len(request.Footnotes))
		for i := range request.Footnotes {
			back := notes[i].LastChild
			So(back.DataAtom, ShouldEqual, atom.A)
			So(GetAttribute(back, "class"), ShouldEqual, "footnote__back-link")
			So(GetAttribute(back, "href"), ShouldEqual, "#table_"+request.Filename)
			So(GetText(back), ShouldResemble, "Back to table")
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
		So(GetAttribute(links[0], "href"), ShouldEqual, "#table_myId_note_1")
		So(GetAttribute(links[1], "href"), ShouldEqual, "#table_myId_note_2")
		So(GetAttribute(links[2], "href"), ShouldEqual, "#table_myId_note_1")

		p := FindNodeWithAttributes(container, atom.P, map[string]string{"class": "table-notes"})
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
		So(GetAttribute(links[0], "href"), ShouldEqual, "#table_myId_note_11")
	})

	Convey("Multiple references to the same footnote in the same value should all be converted to links", t, func() {
		request := models.RenderRequest{Filename: "myId", Footnotes: []string{"Note1", "Note2"}, Title: "This contains [1] links[1]"}
		container, _ := invokeRenderHTML(&request)

		links := FindNodesWithAttributes(container, atom.A, map[string]string{"class": footnoteLinkClass})
		So(len(links), ShouldEqual, 2)
		for _, link := range links {
			So(GetAttribute(link, "href"), ShouldEqual, "#table_myId_note_1")
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
			{Row: 1, Column: 0, Align: models.AlignLeft}}
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
		So(GetAttribute(rows[0], "class"), ShouldContainSubstring, "top")

		td := FindNodes(rows[0], atom.Td)
		So(len(td), ShouldEqual, len(request.Data[0]))
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "bottom")
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "right")
		So(GetAttribute(td[1], "class"), ShouldBeEmpty)

		td = FindNodes(rows[1], atom.Td)
		So(len(td), ShouldEqual, len(request.Data[1]))
		So(GetAttribute(td[0], "class"), ShouldNotContainSubstring, "bottom")
		So(GetAttribute(td[0], "class"), ShouldNotContainSubstring, "right")
		So(GetAttribute(td[0], "class"), ShouldContainSubstring, "left")
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
