package renderer

import (
	"fmt"

	"bytes"
	"regexp"
	"strconv"

	"encoding/json"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/go-ns/log"
)

var (
	columnNames = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterCount = len(columnNames)

	integerPattern       = regexp.MustCompile("^[1-9]+[0-9]*$")
	floatPattern         = regexp.MustCompile("^[0-9]*\\.[0-9]+$")
	decimalPlacesPattern = regexp.MustCompile("\\.[0-9]+")

	formatGeneral  = 0
	formatInt      = 1
	formatFloat2dp = 2
	formatFloat1dp = "0.0"
	formatFloat3dp = "0.000"
	titleFormat    = &xlsxCellStyle{Font: xlsxFont{Bold: true}}

	// a map of the alignments to their xlsx equivalents
	xlsxAlignmentMap = map[string]string{
		models.AlignTop:     "top",
		models.AlignMiddle:  "center",
		models.AlignBottom:  "", // bottom is the default, and doesn't seem to have a value in excelize
		models.AlignLeft:    "left",
		models.AlignCenter:  "center",
		models.AlignRight:   "right",
		models.AlignJustify: "justify",
	}
)

// xlsxCellStyle holds those cell formatting properties we want to define
type xlsxCellStyle struct {
	NumberFormat       int           `json:"number_format,omitempty"`
	CustomNumberFormat string        `json:"custom_number_format,omitempty"`
	Alignment          xlsxAlignment `json:"alignment,omitempty"`
	Font               xlsxFont      `json:"font,omitempty"`
}
type xlsxAlignment struct {
	Horizontal string `json:"horizontal,omitempty"`
	Vertical   string `json:"vertical,omitempty"`
	WrapText   bool   `json:"wrap_text,omitempty"`
}
type xlsxFont struct {
	Bold bool `json:"bold,omitempty"`
}

type spreadsheetModel struct {
	xlsx         *excelize.File
	styleMap     map[string]int
	cellStyles   map[xlsxCellStyle]int
	currentRow   int
	firstDataRow int
	sheet        string
	tableModel   *tableModel
	request      *models.RenderRequest
}

// RenderXLSX returns an xlsx representation of the table generated from the given request
func RenderXLSX(request *models.RenderRequest) ([]byte, error) {

	xlsx := excelize.NewFile()

	model := &spreadsheetModel{
		request:    request,
		tableModel: createModel(request),
		cellStyles: make(map[xlsxCellStyle]int),
		xlsx:       xlsx,
		currentRow: 0,
		sheet:      "Sheet1",
	}

	insertTitle(model)
	insertData(model)
	insertSource(model)
	insertFootnotes(model)

	// TODO: insert units

	mergeCells(model)

	var buf bytes.Buffer
	xlsx.Write(&buf)
	return buf.Bytes(), nil
}

// insertTitle inserts title and subtitle in the spreadsheet
func insertTitle(model *spreadsheetModel) {
	xlsx := model.xlsx
	request := model.request

	axisRef := getAxisRef(model.currentRow, 0)
	xlsx.SetCellStr(model.sheet, axisRef, request.Title)
	xlsx.SetCellStyle(model.sheet, axisRef, axisRef, getStyleRef(model, titleFormat))
	model.currentRow++

	axisRef = getAxisRef(model.currentRow, 0)
	xlsx.SetCellStr(model.sheet, axisRef, request.Subtitle)
	xlsx.SetCellStyle(model.sheet, axisRef, axisRef, getStyleRef(model, titleFormat))
	model.currentRow++
}

// insertData inserts each cell of the table in the spreadsheet, unless hidden by a merged cell
func insertData(model *spreadsheetModel) {
	xlsx := model.xlsx
	tableModel := model.tableModel
	model.firstDataRow = model.currentRow + 1

	for r, row := range model.request.Data {
		model.currentRow++
		for c := range row {
			if cellIsVisible(tableModel, r, c) {
				value, style := getCellValueAndStyle(model, r, c)
				axisRef := getAxisRef(model.currentRow, c)
				xlsx.SetCellValue(model.sheet, axisRef, value)
				xlsx.SetCellStyle(model.sheet, axisRef, axisRef, style)
			}
		}
	}
	model.currentRow++
}

// cellIsVisible returns true if the cell is visible (not hidden by a merged cell)
func cellIsVisible(tableModel *tableModel, r int, c int) bool {
	return tableModel.cells[r][c] == nil || tableModel.cells[r][c].skip == false
}

// insertSource inserts the source in the spreadsheet
func insertSource(model *spreadsheetModel) {
	xlsx := model.xlsx
	if len(model.request.Source) > 0 {
		model.currentRow++
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 0), sourceText)
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 1), model.request.Source)
	}
}

// insertFootnotes inserts footnotes in the spreadsheet
func insertFootnotes(model *spreadsheetModel) {
	xlsx := model.xlsx
	request := model.request

	if len(request.Footnotes) > 0 {
		model.currentRow++
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 0), notesText)
		for i, note := range request.Footnotes {
			model.currentRow++
			xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 0), fmt.Sprintf("%d.", i+1))
			xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 1), note)
		}
	}
}

// mergeCells applies the merge operation to the appropriate cells
func mergeCells(model *spreadsheetModel) {
	xlsx := model.xlsx
	for _, format := range model.request.CellFormats {
		if format.Rowspan > 1 || format.Colspan > 1 {
			topRow := format.Row + model.firstDataRow
			topLeft := getAxisRef(topRow, format.Column)
			rowspan := format.Rowspan
			if rowspan > 0 {
				rowspan--
			}
			colspan := format.Colspan
			if colspan > 0 {
				colspan--
			}
			bottomRight := getAxisRef(topRow+rowspan, format.Column+colspan)
			xlsx.MergeCell(model.sheet, topLeft, bottomRight)
		}
	}
}

// getAxisRef returns the spreadsheet reference for the given cell coordinates, e.g. 'A1' for [0,0]
func getAxisRef(row int, col int) string {
	prefix := ""
	offset := 0
	for col >= letterCount {
		// this will work for the first 26**2 columns - should be enough for any reasonable spreadsheet
		prefix = string(columnNames[offset])
		col -= letterCount
		offset++
	}
	colName := string(columnNames[col])
	return fmt.Sprintf("%s%s%d", prefix, colName, row+1)
}

// getCellValueAndStyle converts the cell value to the appropriate type [string|int|float] and creates the correct cell style for formatting and alignment
func getCellValueAndStyle(model *spreadsheetModel, row int, col int) (interface{}, int) {
	value := model.request.Data[row][col]
	cellContent, cellStyle, err := parseValueAndFormat(value)
	if err != nil {
		log.ErrorC(model.request.Filename, err, log.Data{"_message": "Unable to parse value", "value": value})
		cellContent = value
	}
	align, valign, isHeading := getCellAlignmentAndHeading(model, row, col)
	cellStyle.Alignment.Horizontal = xlsxAlignmentMap[align]
	cellStyle.Alignment.Vertical = xlsxAlignmentMap[valign]
	if isHeading {
		cellStyle.Font.Bold = true
		cellStyle.Alignment.WrapText = true
	}
	return cellContent, getStyleRef(model, cellStyle)
}

// parseValueAndFormat parses the value string into an integer or float if possible, and creates a style with an appropriate number format according to the type and number of decimal places
func parseValueAndFormat(value string) (interface{}, *xlsxCellStyle, error) {
	cellStyle := &xlsxCellStyle{}
	var cellContent interface{}
	var err error
	if integerPattern.MatchString(value) {
		cellContent, err = strconv.Atoi(value)
		cellStyle.NumberFormat = formatInt
	} else if floatPattern.MatchString(value) {
		cellContent, err = strconv.ParseFloat(value, 64)
		switch len(decimalPlacesPattern.FindString(value)) - 1 {
		case 1:
			cellStyle.CustomNumberFormat = formatFloat1dp
		case 2:
			cellStyle.NumberFormat = formatFloat2dp
		case 3:
			cellStyle.CustomNumberFormat = formatFloat3dp
		default:
			cellStyle.NumberFormat = formatGeneral
		}
	} else {
		cellContent = value
	}
	return cellContent, cellStyle, err
}

// getCellAlignmentAndHeading returns the alignment, vertical alignment and whether the cell is a heading
func getCellAlignmentAndHeading(model *spreadsheetModel, row int, col int) (string, string, bool) {
	rowFormat := model.tableModel.rows[row]
	colFormat := model.tableModel.columns[col]
	cellFormat := model.tableModel.cells[row][col]
	align := colFormat.Align
	valign := rowFormat.VerticalAlign
	if cellFormat != nil {
		if len(cellFormat.align) > 0 {
			align = cellFormat.align
		}
		if len(cellFormat.valign) > 0 {
			valign = cellFormat.valign
		}
	}
	isHeading := rowFormat.Heading || colFormat.Heading
	return align, valign, isHeading
}

// getStyleRef finds an existing style with the required properties, creating one if none can be found, and returning the index of that style
func getStyleRef(model *spreadsheetModel, format *xlsxCellStyle) int {
	if i, exists := model.cellStyles[*format]; exists {
		return i
	}
	bytes, e := json.Marshal(*format)
	if e != nil {
		log.ErrorC(model.request.Filename, e, log.Data{"_message": "Unable to marshal an xlsxCellStyle"})
		return 0
	}
	style, err := model.xlsx.NewStyle(string(bytes))
	if err != nil {
		log.Error(err, log.Data{"_message": "Unable create new style for spreadsheet", "value": string(bytes)})
		return 0
	}
	model.cellStyles[*format] = style
	return style
}
