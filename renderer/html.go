package renderer

import (
	"bytes"
	"context"
	"fmt"

	"regexp"

	"strings"

	h "github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/log.go/v2/log"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	newLine        = regexp.MustCompile(`\n`)
	footnoteLink   = regexp.MustCompile(`\[[0-9]+]`)
	emptyCellModel = &cellModel{}

	// text that will need internationalising at some point:
	sourceText         = "Source: "
	unitsText          = "Units: "
	notesText          = "Notes"
	footnoteHiddenText = "Footnote "

	// a map of the alignments to their css classes
	cssAlignmentMap = map[string]string{
		models.AlignTop:     "align-top",
		models.AlignMiddle:  "align-middle",
		models.AlignBottom:  "align-bottom",
		models.AlignLeft:    "align-left",
		models.AlignCenter:  "align-center",
		models.AlignRight:   "align-right",
		models.AlignJustify: "align-justify",
	}
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
	align   string
	valign  string
}

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(ctx context.Context, request *models.RenderRequest) ([]byte, error) {
	model := createModel(ctx, request)

	figure := h.CreateNode("figure", atom.Figure,
		h.Attr("class", "figure"),
		h.Attr("id", tableID(request)),
		"\n")

	table := addTable(ctx, request, figure)

	addColumnGroup(model, table)
	addRows(ctx, model, table)

	addFooter(ctx, request, figure)

	var buf bytes.Buffer
	html.Render(&buf, figure)
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// tableID returns the id for the table, as used in links etc
func tableID(request *models.RenderRequest) string {
	return "table-" + request.Filename
}

// addTable creates a table node with a caption and adds it to the given node
func addTable(ctx context.Context, request *models.RenderRequest, parent *html.Node) *html.Node {
	table := h.CreateNode("table", atom.Table,
		h.Attr("class", "table"),
		"\n")

	// add title and subtitle as a caption
	if len(request.Title) > 0 || len(request.Subtitle) > 0 {
		caption := h.CreateNode("caption", atom.Caption,
			h.Attr("class", "table__caption"),
			parseValue(ctx, request, request.Title))
		if len(request.Subtitle) > 0 {
			subtitle := h.CreateNode("span", atom.Span,
				h.Attr("class", "table__subtitle"),
				parseValue(ctx, request, request.Subtitle))

			caption.AppendChild(h.CreateNode("br", atom.Br))
			caption.AppendChild(subtitle)

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
func addRows(ctx context.Context, model *tableModel, table *html.Node) {
	for rowIdx, row := range model.request.Data {
		tr := h.CreateNode("tr", atom.Tr)
		table.AppendChild(tr)
		if model.rows[rowIdx].Heading {
			h.AddAttribute(tr, "class", "table__header-row")
			if model.request.KeepHeadersTogether {
				h.AppendAttribute(tr, "class", "table__nowrap")
			}
		}
		if len(model.rows[rowIdx].VerticalAlign) > 0 {
			h.AppendAttribute(tr, "class", mapAlignmentToClass(model.rows[rowIdx].VerticalAlign))
		}
		if len(model.rows[rowIdx].Height) > 0 {
			h.AddAttribute(tr, "style", "height: "+model.rows[rowIdx].Height)
		}
		for colIdx, col := range row {
			addTableCell(ctx, model, tr, col, rowIdx, colIdx)
		}
		table.AppendChild(h.Text("\n"))
	}
}

// adds an individual table cell to the given tr node
func addTableCell(ctx context.Context, model *tableModel, tr *html.Node, colText string, rowIdx int, colIdx int) {
	cell := model.cells[rowIdx][colIdx]
	if cell == nil {
		cell = emptyCellModel
	}
	if cell.skip {
		return
	}
	value := parseValueWithHtml(ctx, model.request, colText)
	hasContent := len(colText) > 0
	var node *html.Node
	if model.rows[rowIdx].Heading && hasContent {
		node = h.CreateNode("th", atom.Th, h.Attr("scope", "col"), value)
		if cell.colspan > 1 {
			h.ReplaceAttribute(node, "scope", "colgroup")
		}
	} else if model.columns[colIdx].Heading && hasContent {
		node = h.CreateNode("th", atom.Th, h.Attr("scope", "row"), value)
		if cell.rowspan > 1 {
			h.ReplaceAttribute(node, "scope", "rowgroup")
		}
		if model.request.KeepHeadersTogether {
			h.AddAttribute(node, "class", "table__nowrap")
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
	if len(cell.align) > 0 {
		h.AppendAttribute(node, "class", mapAlignmentToClass(cell.align))
	} else if len(model.columns[colIdx].Align) > 0 {
		h.AppendAttribute(node, "class", mapAlignmentToClass(model.columns[colIdx].Align))
	}
	if len(cell.valign) > 0 {
		h.AppendAttribute(node, "class", mapAlignmentToClass(cell.valign))
	}
	tr.AppendChild(node)
}

// mapAlignmentToClass converts a VerticalAlign or Align value into a css class
func mapAlignmentToClass(align string) string {
	return cssAlignmentMap[align]
}

// addFooter adds a footer to the given element, containing the source and footnotes
func addFooter(ctx context.Context, request *models.RenderRequest, parent *html.Node) {
	footer := h.CreateNode("footer", atom.Footer,
		h.Attr("class", "figure__footer"),
		"\n")
	if len(request.Units) > 0 {
		footer.AppendChild(h.CreateNode("p", atom.P,
			h.Attr("class", "figure__units"),
			parseValue(ctx, request, unitsText+request.Units)))
		footer.AppendChild(h.Text("\n"))
	}
	if len(request.Source) > 0 {
		footer.AppendChild(h.CreateNode("p", atom.P,
			h.Attr("class", "figure__source"),
			parseValue(ctx, request, sourceText+request.Source)))
		footer.AppendChild(h.Text("\n"))
	}
	if len(request.Footnotes) > 0 {
		footer.AppendChild(h.CreateNode("p", atom.P,
			h.Attr("class", "figure__notes"),
			notesText))
		footer.AppendChild(h.Text("\n"))

		ol := h.CreateNode("ol", atom.Ol,
			h.Attr("class", "figure__footnotes"),
			"\n")
		addFooterItemsToList(ctx, request, ol)
		footer.AppendChild(ol)
		footer.AppendChild(h.Text("\n"))
	}
	parent.AppendChild(footer)
	parent.AppendChild(h.Text("\n"))
}

// addFooterItemsToList adds one li node for each footnote to the given list node
func addFooterItemsToList(ctx context.Context, request *models.RenderRequest, ol *html.Node) {
	for i, note := range request.Footnotes {
		li := h.CreateNode("li", atom.Li,
			h.Attr("id", fmt.Sprintf("table-%s-note-%d", request.Filename, i+1)),
			h.Attr("class", "figure__footnote-item"),
			parseValue(ctx, request, note))
		ol.AppendChild(li)
		ol.AppendChild(h.Text("\n"))
	}
}

// Parses the string to replace \n with <br /> and wrap [1] with a link to the footnote
func parseValue(ctx context.Context, request *models.RenderRequest, value string) []*html.Node {
	hasBr := newLine.MatchString(value)
	hasFootnote := len(request.Footnotes) > 0 && footnoteLink.MatchString(value)
	if hasBr || hasFootnote {
		return replaceValues(ctx, request, value, hasBr, hasFootnote)
	}
	return []*html.Node{{Type: html.TextNode, Data: value}}
}

// Parses the string to replace \n with <br /> and wrap [1] with a link to the footnote
// keeping any HTML tags that are present
func parseValueWithHtml(ctx context.Context, request *models.RenderRequest, value string) []*html.Node {
	hasBr := newLine.MatchString(value)
	hasFootnote := len(request.Footnotes) > 0 && footnoteLink.MatchString(value)
	if hasBr || hasFootnote {
		return replaceValues(ctx, request, value, hasBr, hasFootnote)
	}
	// Create Nodes
	nodes, _ := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body})
	return nodes
}

// replaceValues uses regexp to replace new lines and footnotes with <br/> and <a>.../<a> tags, then parses the result into an array of nodes
func replaceValues(ctx context.Context, request *models.RenderRequest, value string, hasBr bool, hasFootnote bool) []*html.Node {
	original := value
	if hasBr {
		value = newLine.ReplaceAllLiteralString(value, "<br />")
	}
	if hasFootnote {
		for i := range request.Footnotes {
			n := i + 1
			linkText := fmt.Sprintf("<a href=\"#table-%s-note-%d\" class=\"footnote__link\"><span class=\"visuallyhidden\">%s</span>%d</a>", request.Filename, n, footnoteHiddenText, n)
			value = strings.Replace(value, fmt.Sprintf("[%d]", n), linkText, -1)
		}
	}
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		log.Error(ctx, "unable to parse value", err, log.Data{"value": original})
		return []*html.Node{{Type: html.TextNode, Data: original}}
	}
	return nodes
}

// Creates a tableModel containing calculations that are referenced more than once while rendering the table
func createModel(ctx context.Context, request *models.RenderRequest) *tableModel {
	m := tableModel{request: request}
	m.columns = indexColumnFormats(ctx, request)
	m.rows = indexRowFormats(ctx, request)
	m.cells = createCellModels(request)
	return &m
}

// indexes the ColumnFormats so that columns[i] gives the correct format for column i
func indexColumnFormats(ctx context.Context, request *models.RenderRequest) []models.ColumnFormat {
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
			log.Info(ctx, "ColumnFormat specified for non-existent column", log.Data{"file_name": request.Filename, "column_format": format, "column_count": count})
		} else {
			columns[format.Column] = format
		}
	}
	return columns
}

// indexes the RowFormats so that rows[i] gives the correct format for row i
func indexRowFormats(ctx context.Context, request *models.RenderRequest) []models.RowFormat {
	count := len(request.Data)
	// create default RowFormats
	rows := make([]models.RowFormat, count)
	for i := range rows {
		rows[i] = models.RowFormat{Row: i}
	}
	// replace with actual RowFormats where they exist
	for _, format := range request.RowFormats {
		if format.Row >= count || format.Row < 0 {
			log.Info(ctx, "ColumnFormat specified for non-existent column", log.Data{"file_name": request.Filename, "row_format": format, "row_count": count})
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
		cell.align = format.Align
		cell.valign = format.VerticalAlign
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
