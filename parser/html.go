package parser

import (
	"encoding/json"

	"bufio"
	"bytes"
	"errors"
	"strings"

	"fmt"
	"regexp"
	"strconv"

	h "github.com/ONSdigital/dp-table-renderer/htmlutil"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/go-ns/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// parseModel contains values calculated from the parse request that are used to create the ResponseModel
type parseModel struct {
	request       *models.ParseRequest
	tableNode     *html.Node
	cells         [][]*html.Node    // all cells we want to treat as data
	rowClasses    []map[string]int  // all classes defined in a row, keyed by the row index, with a count of the number of cells having that class
	columnClasses []map[string]int  // all classes defined in a column, keyed by the column index, with a count of the number of cells having that class
	alignMap      map[string]string // a map of the classes used for alignment in the input html to the correct Alignment values
	valignMap     map[string]string // a map of the classes used for vertical alignment in the input html to the correct Alignment values
}

// ResponseModel defines the format of the json response contained in the bytes returned from ParseHTML
type ResponseModel struct {
	JSON        models.RenderRequest `json:"render_json"`
	PreviewHTML string               `json:"preview_html"`
}

var (
	widthStylePattern          = regexp.MustCompile(`width: *[0-9]+[^;]+`)
	widthTrailingZeroesPattern = regexp.MustCompile(`\.?0+(%|em)`)
	tableType                  = "table"
	tableVersion               = "2"
)

// ParseHTML parses the html table in the request and generates correctly formatted JSON
func ParseHTML(request *models.ParseRequest) ([]byte, error) {

	sourceTable, err := parseTableToNode(request.TableHTML)
	if err != nil {
		log.Error(err, log.Data{"message": "Unable to parse TableHTML to table element", "ParseRequest": request})
		return nil, err
	}

	model := createParseModel(request, sourceTable)
	requestJSON := &models.RenderRequest{
		Filename:     request.Filename,
		Title:        request.Title,
		Subtitle:     request.Subtitle,
		Source:       request.Source,
		Units:        request.Units,
		TableType:    tableType,
		TableVersion: tableVersion,
		Footnotes:    request.Footnotes}

	rowFormats := createRowFormats(model)
	requestJSON.RowFormats = convertRowFormatsToSlice(rowFormats)

	colFormats := createColumnFormats(model)
	requestJSON.ColumnFormats = convertColumnFormatsToSlice(colFormats)

	requestJSON.CellFormats = createCellFormats(model, rowFormats, colFormats)

	requestJSON.Data = parseData(model.cells)

	previewHTML, err := renderer.RenderHTML(requestJSON)
	if err != nil {
		log.Error(err, log.Data{"message": "Unable to render preview HTML", "ParseRequest": request, "RenderRequest": requestJSON})
		return nil, err
	}
	response := ResponseModel{JSON: *requestJSON, PreviewHTML: string(previewHTML)}

	return marshalResponse(response)
}

// parseTableToNode parses a string of html and returns the single table node, or an error if the html doesn't contain a single table
func parseTableToNode(tableHTML string) (*html.Node, error) {
	nodes, err := html.ParseFragment(strings.NewReader(tableHTML), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, errors.New("table_html could not be parsed into a single element")
	}
	if nodes[0].DataAtom != atom.Table {
		return nil, errors.New("table_html could not be parsed into a table element")
	}
	return nodes[0], nil
}

// createParseModel creates a model from the input request, extracting all properties need to define the output from the input html
func createParseModel(request *models.ParseRequest, tableNode *html.Node) *parseModel {
	model := parseModel{}
	model.request = request
	model.tableNode = tableNode

	model.cells = getCells(tableNode, request.IgnoreFirstRow, request.IgnoreFirstColumn)

	model.rowClasses, model.columnClasses = parseRowAndColumnClasses(model.cells)

	model.alignMap = map[string]string{
		request.AlignmentClasses.Left:    models.AlignLeft,
		request.AlignmentClasses.Center:  models.AlignCenter,
		request.AlignmentClasses.Right:   models.AlignRight,
		request.AlignmentClasses.Justify: models.AlignJustify,
	}

	model.valignMap = map[string]string{
		request.AlignmentClasses.Bottom: models.AlignBottom,
		request.AlignmentClasses.Middle: models.AlignMiddle,
		request.AlignmentClasses.Top:    models.AlignTop,
	}

	// in case the alignment classes were not specified:
	delete(model.alignMap, "")
	delete(model.valignMap, "")

	return &model
}

// getCells returns all td and th elements in the given (table) Node, in a 2d array with one array for each row in the table
func getCells(table *html.Node, ignoreFirstRow bool, ignoreFirstColumn bool) [][]*html.Node {
	var cells [][]*html.Node
	rows := h.FindAllNodes(table, atom.Tr)
	if ignoreFirstRow {
		rows = rows[1:]
	}
	for _, row := range rows {
		columns := h.FindAllNodes(row, atom.Td, atom.Th)
		if ignoreFirstColumn {
			columns = columns[1:]
		}
		cells = append(cells, columns)
	}
	return cells
}

// parseData extracts the text content of each cell in the given array
func parseData(cells [][]*html.Node) [][]string {
	var data [][]string
	for _, row := range cells {
		var rowData []string
		for _, cell := range row {
			rowData = append(rowData, h.GetText(cell))
		}
		data = append(data, rowData)
	}
	return data
}

// parseRowAndColumnClasses iterates through the cells, recording which classes occur in which rows/columns, and how often
func parseRowAndColumnClasses(cells [][]*html.Node) ([]map[string]int, []map[string]int) {
	columnClasses := make([]map[string]int, 0)
	rowClasses := make([]map[string]int, len(cells))
	for r, row := range cells {
		rowClasses[r] = make(map[string]int)
		for c, cell := range row {
			if len(columnClasses) < c+1 {
				columnClasses = append(columnClasses, make(map[string]int))
			}
			// get all the classes of this cell
			classes := strings.Split(h.GetAttribute(cell, "class"), " ")
			for _, class := range classes {
				if len(class) > 0 {
					// increment the counter for this class in row r and column c
					rowClasses[r][class]++
					columnClasses[c][class]++
				}
			}
		}
	}
	return rowClasses, columnClasses
}

// createColumnFormats uses the colgroup element to determine widths, and columnClasses to determine alignment
func createColumnFormats(model *parseModel) map[int]models.ColumnFormat {
	numRows := len(model.cells)
	colFormats := make(map[int]models.ColumnFormat)
	// extract alignment from the column classes
	for i, classes := range model.columnClasses {
		for class, count := range classes {
			if count == numRows && len(model.alignMap[class]) > 0 {
				format := colFormats[i]
				format.Align = model.alignMap[class]
				colFormats[i] = format
			}
		}
	}
	// extract widths from col elements - assume that col elements do not have a colspan
	// TODO handle cases where col elements have colspan
	colgroup := h.FindNodes(model.tableNode, atom.Col)
	if len(colgroup) > 0 && model.request.IgnoreFirstColumn {
		colgroup = colgroup[1:]
	}
	for i, col := range colgroup {
		width := extractWidth(model, col)
		if len(width) > 0 {
			format := colFormats[i]
			format.Width = width
			colFormats[i] = format
		}
	}
	// assign headings
	for i := 0; i < model.request.HeaderCols; i++ {
		format := colFormats[i]
		format.Heading = true
		colFormats[i] = format
	}
	return colFormats
}

// createRowFormats uses rowClasses to determine row vertical alignment
func createRowFormats(model *parseModel) map[int]models.RowFormat {
	rowFormats := make(map[int]models.RowFormat)
	// extract alignment from the row classes
	for i, classes := range model.rowClasses {
		numColumns := len(model.cells[i])
		for class, count := range classes {
			if count == numColumns && len(model.valignMap[class]) > 0 {
				format := rowFormats[i]
				format.VerticalAlign = model.valignMap[class]
				rowFormats[i] = format
			}
		}
	}
	// assign headings
	for i := 0; i < model.request.HeaderRows; i++ {
		format := rowFormats[i]
		format.Heading = true
		rowFormats[i] = format
	}
	return rowFormats
}

// convertRowFormatsToSlice converts the map to an ordered slice
func convertRowFormatsToSlice(rowFormats map[int]models.RowFormat) []models.RowFormat {
	var keys []int
	for k := range rowFormats {
		keys = append(keys, k)
	}
	slice := []models.RowFormat{}
	for key := range keys {
		format := rowFormats[key]
		format.Row = key
		slice = append(slice, format)
	}
	return slice
}

// convertColumnFormatsToSlice converts the map to an ordered slice
func convertColumnFormatsToSlice(colFormats map[int]models.ColumnFormat) []models.ColumnFormat {
	var keys []int
	for k := range colFormats {
		keys = append(keys, k)
	}
	slice := []models.ColumnFormat{}
	for key := range keys {
		format := colFormats[key]
		format.Column = key
		slice = append(slice, format)
	}
	return slice
}

// createCellFormats assigns rowspan and colspan, and align/vertical align if necessary
func createCellFormats(model *parseModel, rowFormats map[int]models.RowFormat, colFormats map[int]models.ColumnFormat) []models.CellFormat {
	cellFormats := []models.CellFormat{}
	for r, row := range model.cells {
		for c, cell := range row {
			format := models.CellFormat{}
			hasData := false
			colspan, _ := strconv.Atoi(h.GetAttribute(cell, "colspan"))
			if colspan > 0 {
				format.Colspan = colspan
				hasData = true
			}
			rowspan, _ := strconv.Atoi(h.GetAttribute(cell, "rowspan"))
			if colspan > 0 {
				format.Rowspan = rowspan
				hasData = true
			}
			classes := strings.Split(h.GetAttribute(cell, "class"), " ")
			for _, class := range classes {
				// specify vertical align if the cell has an alignment different to that of the row
				valign := model.valignMap[class]
				if len(valign) > 0 && valign != rowFormats[r].VerticalAlign {
					format.VerticalAlign = valign
					hasData = true
				}
				// specify align if the cell has an alignment different to that of the column
				align := model.alignMap[class]
				if len(align) > 0 && align != colFormats[c].Align {
					format.Align = align
					hasData = true
				}
			}
			if hasData {
				format.Column = c
				format.Row = r
				cellFormats = append(cellFormats, format)
			}
		}
	}
	return cellFormats
}

// extractWidth extracts width from the style property of the node
func extractWidth(model *parseModel, node *html.Node) string {
	if model.request.CellSizeUnits == "auto" {
		return ""
	}
	width := widthStylePattern.FindString(h.GetAttribute(node, "style"))
	width = strings.Trim(strings.Replace(width, "width:", "", -1), " ")
	width = strings.Replace(width, model.request.ColumnWidthToIgnore, "", -1)
	// replace pixel width with % or em
	if strings.HasSuffix(width, "px") {
		switch units := model.request.CellSizeUnits; units {
		case "%":
			if model.request.CurrentTableWidth > 0 {
				intWidth, err := strconv.Atoi(strings.Trim(width, "px"))
				if err == nil {
					proportion := float32(intWidth) / float32(model.request.CurrentTableWidth)
					width = fmt.Sprintf("%.1f%%", proportion*100.0)
				} else {
					log.ErrorC(model.request.Filename, err, log.Data{"Width not parsable as an integer": width})
				}
			}
		case "em":
			if model.request.SingleEmHeight > 0 {
				intWidth, err := strconv.Atoi(strings.Trim(width, "px"))
				if err == nil {
					width = fmt.Sprintf("%.2fem", (float32(intWidth))/model.request.SingleEmHeight)
				} else {
					log.ErrorC(model.request.Filename, err, log.Data{"Width not parsable as an integer": width})
				}
			}
		}
		// strip unwanted trailing zeroes in decimal places
		width = widthTrailingZeroesPattern.ReplaceAllString(width, "$1")
	}
	return width
}

// marshalResponse marshals the ResponseModel to json, turning off escaping of html
func marshalResponse(response ResponseModel) ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(response)
	if err == nil {
		err = writer.Flush()
	}
	return b.Bytes(), err
}
