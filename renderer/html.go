package renderer

import (
	"bytes"
	"fmt"

	"regexp"

	"strings"

	h "github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/go-ns/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	newLine        = regexp.MustCompile(`\n`)
	footnoteLink   = regexp.MustCompile(`\[[0-9+]]`)
	emptyCellModel = &cellModel{}
)

// Contains details of the table that need to be calculated once from the request and cached
type tableModel struct {
	request *models.RenderRequest
	columns []models.ColumnFormat
	rows    []models.RowFormat
	cells   map[int]map[int]*cellModel
}

// contains details of a cell that requires special handling
type cellModel struct {
	skip    bool
	colspan int
	rowspan int
	class   string
}

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(request *models.RenderRequest) ([]byte, error) {
	model := createModel(request)

	div := h.CreateNode("div", atom.Div,
		h.Attr("class", "table-renderer"),
		h.Attr("id", "table_"+request.Filename),
		"\n")

	table := addTable(request, div)

	addColumnGroup(model, table)
	addRows(model, table)

	addFooter(request, div)

	var buf bytes.Buffer
	html.Render(&buf, div)
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// addTable creates a table node with a caption and adds it to the given node
func addTable(request *models.RenderRequest, parent *html.Node) *html.Node {
	table := h.CreateNode("table", atom.Table, "\n")

	// add title and subtitle as a caption
	if len(request.Title) > 0 || len(request.Subtitle) > 0 {
		caption := h.CreateNode("caption", atom.Caption, parseValue(request, request.Title))
		if len(request.Subtitle) > 0 {
			subtitleID := fmt.Sprintf("table_%s_description", request.Filename)
			subtitle := h.CreateNode("span", atom.Span,
				h.Attr("id", subtitleID),
				h.Attr("class", "table-subtitle"),
				parseValue(request, request.Subtitle))

			caption.AppendChild(h.CreateNode("br", atom.Br))
			caption.AppendChild(subtitle)

			h.AddAttribute(table, "aria-describedby", subtitleID)
		}
		table.AppendChild(caption)
		table.AppendChild(h.Text("\n"))
	}
	parent.AppendChild(table)
	return table
}

// addColumnGroup adds a columnGroup, if required, to the given table. Cols in the colgroup specify column width.
func addColumnGroup(model *tableModel, table *html.Node) {
	if len(model.request.ColumnFormats) > 0 {
		colgroup := h.CreateNode("colgroup", atom.Colgroup)

		for _, col := range model.columns {
			node := h.CreateNode("col", atom.Col)
			if len(col.Width) > 0 {
				h.AddAttribute(node, "style", "width: "+col.Width)
			}
			colgroup.AppendChild(node)
		}

		table.AppendChild(colgroup)
		table.AppendChild(h.Text("\n"))
	}
}

// adds all rows to the table. Rows contain th or td cells as appropriate.
func addRows(model *tableModel, table *html.Node) {
	for rowIdx, row := range model.request.Data {
		tr := h.CreateNode("tr", atom.Tr)
		table.AppendChild(tr)
		if len(model.rows[rowIdx].VerticalAlign) > 0 {
			h.AddAttribute(tr, "class", model.rows[rowIdx].VerticalAlign)
		}
		if len(model.rows[rowIdx].Height) > 0 {
			h.AddAttribute(tr, "style", "height: "+model.rows[rowIdx].Height)
		}
		for colIdx, col := range row {
			addTableCell(model, tr, col, rowIdx, colIdx)
		}
		table.AppendChild(h.Text("\n"))
	}
}

// adds an individual table cell to the given tr node
func addTableCell(model *tableModel, tr *html.Node, colText string, rowIdx int, colIdx int) {
	cell := model.cells[rowIdx][colIdx]
	if cell == nil {
		cell = emptyCellModel
	}
	if cell.skip {
		return
	}
	value := parseValue(model.request, colText)
	var node *html.Node
	if model.rows[rowIdx].Heading {
		node = h.CreateNode("th", atom.Th, h.Attr("scope", "col"), value)
		if cell.colspan > 1 {
			h.ReplaceAttribute(node, "scope", "colgroup")
		}
	} else if model.columns[colIdx].Heading {
		node = h.CreateNode("th", atom.Th, h.Attr("scope", "row"), value)
		if cell.rowspan > 1 {
			h.ReplaceAttribute(node, "scope", "rowgroup")
		}
	} else {
		node = h.CreateNode("td", atom.Td, value)
	}
	if cell.colspan > 1 {
		h.AddAttribute(node, "colspan", fmt.Sprintf("%d", cell.colspan))
	}
	if cell.rowspan > 1 {
		h.AddAttribute(node, "rowspan", fmt.Sprintf("%d", cell.rowspan))
	}
	if len(model.columns[colIdx].Align) > 0 {
		h.AddAttribute(node, "class", model.columns[colIdx].Align)
	}
	if len(cell.class) > 0 {
		h.ReplaceAttribute(node, "class", strings.Trim(h.GetAttribute(node, "class") + " " + cell.class, " "))
	}
	tr.AppendChild(node)
}

// addFooter adds a footer to the given element, containing the source and footnotes
func addFooter(request *models.RenderRequest, parent *html.Node) {
	footer := h.CreateNode("footer", atom.Footer, "\n")
	if len(request.Source) > 0 {
		footer.AppendChild(h.CreateNode("p", atom.P,
			h.Attr("class", "table-source"),
			parseValue(request, "Source: "+request.Source)))
		footer.AppendChild(h.Text("\n"))
	}
	if len(request.Footnotes) > 0 {
		footer.AppendChild(h.CreateNode("p", atom.P,
			h.Attr("class", "table-notes"),
			h.Attr("id", "table_"+request.Filename+"_notes"),
			"Notes"))
		footer.AppendChild(h.Text("\n"))
		ol := h.CreateNode("ol", atom.Ol, "\n")

		for i, note := range request.Footnotes {
			ol.AppendChild(h.CreateNode("li", atom.Li,
				h.Attr("id", fmt.Sprintf("table_%s_note_%d", request.Filename, i+1)),
				parseValue(request, note)))
			ol.AppendChild(h.Text("\n"))
		}

		footer.AppendChild(ol)
		footer.AppendChild(h.Text("\n"))
	}
	parent.AppendChild(footer)
	parent.AppendChild(h.Text("\n"))
}

// Parses the string to replace \n with <br /> and wrap [1] with a link to the footnote
func parseValue(request *models.RenderRequest, value string) []*html.Node {
	hasBr := newLine.MatchString(value)
	hasFootnote := len(request.Footnotes) > 0 && footnoteLink.MatchString(value)
	if hasBr || hasFootnote {
		return replaceValues(request, value, hasBr, hasFootnote)
	}
	return []*html.Node{{Type: html.TextNode, Data: value}}
}

// replaceValues uses regexp to replace new lines and footnotes with <br/> and <a>.../<a> tags, then parses the result into an array of nodes
func replaceValues(request *models.RenderRequest, value string, hasBr bool, hasFootnote bool) []*html.Node {
	original := value
	if hasBr {
		value = newLine.ReplaceAllLiteralString(value, "<br />")
	}
	if hasFootnote {
		for i := range request.Footnotes {
			n := i + 1
			linkText := fmt.Sprintf("<a aria-describedby=\"table_%s_notes\" href=\"#table_%s_note_%d\" class=\"footnote-link\">[%d]</a>", request.Filename, request.Filename, n, n)
			value = strings.Replace(value, fmt.Sprintf("[%d]", n), linkText, -1)
		}
	}
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		log.ErrorC(request.Filename, err, log.Data{"replaceValues": "Unable to parse value!", "value": original})
		return []*html.Node{{Type: html.TextNode, Data: original}}
	}
	return nodes
}

// Creates a tableModel containing calculations that are referenced more than once while rendering the table
func createModel(request *models.RenderRequest) *tableModel {
	m := tableModel{request: request}
	m.columns = indexColumnFormats(request)
	m.rows = indexRowFormats(request)
	m.cells = createCellModels(request)
	return &m
}

// indexes the ColumnFormats so that columns[i] gives the correct format for column i
func indexColumnFormats(request *models.RenderRequest) []models.ColumnFormat {
	// find the maximum number of columns in the data - should be the same in every row, but don't trust that
	count := 0
	for i := range request.Data {
		n := len(request.Data[i])
		if n > count {
			count = n
		}
	}
	// create default ColumnFormats
	columns := make([]models.ColumnFormat, count)
	for i := range columns {
		columns[i] = models.ColumnFormat{Column: i}
	}
	// replace with actual ColumnFormats where they exist
	for _, format := range request.ColumnFormats {
		if format.Column >= count || format.Column < 0 {
			log.Debug("ColumnFormat specified for non-existent column", log.Data{"filename": request.Filename, "ColumnFormat": format, "column_count": count})
		} else {
			columns[format.Column] = format
		}
	}
	return columns
}

// indexes the RowFormats so that rows[i] gives the correct format for row i
func indexRowFormats(request *models.RenderRequest) []models.RowFormat {
	count := len(request.Data)
	// create default RowFormats
	rows := make([]models.RowFormat, count)
	for i := range rows {
		rows[i] = models.RowFormat{Row: i}
	}
	// replace with actual RowFormats where they exist
	for _, format := range request.RowFormats {
		if format.Row >= count || format.Row < 0 {
			log.Debug("RowFormat specified for non-existent row", log.Data{"filename": request.Filename, "RowFormat": format, "row_count": count})
		} else {
			rows[format.Row] = format
		}
	}
	return rows
}

// creates a map with one cellModel for each cell that requires special handling
func createCellModels(request *models.RenderRequest) map[int]map[int]*cellModel {
	m := make(map[int]map[int]*cellModel)
	for _, format := range request.CellFormats {
		cell := getCellModel(m, format.Row, format.Column)
		cell.colspan = format.Colspan
		cell.rowspan = format.Rowspan
		if len(format.Align) > 0 || len(format.VerticalAlign) > 0 {
			cell.class = strings.Trim(format.Align+" "+format.VerticalAlign, " ")
		}
		// if we have merged cells, find those that need to be skipped in the output
		colspan := min(format.Colspan, 1)
		rowspan := min(format.Rowspan, 1)
		for c := 0; c < colspan; c++ {
			for r := 0; r < rowspan; r++ {
				if (c + r) > 0 {
					otherCell := getCellModel(m, format.Row+r, format.Column+c)
					otherCell.skip = true
				}
			}
		}
	}
	return m
}

func min(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// getCellModel finds the requested cellModel from the map, creating the cellModel and parent map if necessary
func getCellModel(m map[int]map[int]*cellModel, r int, c int) *cellModel {
	row, exists := m[r]
	if !exists {
		row = make(map[int]*cellModel)
		m[r] = row
	}
	cell, exists := row[c]
	if !exists {
		cell = &cellModel{}
		row[c] = cell
	}
	return cell
}
