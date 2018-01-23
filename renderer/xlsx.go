package renderer

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ONSdigital/dp-table-renderer/models"
	"bytes"
	"regexp"
	"strconv"
	"github.com/go-ns/log"
)

const (
	intStyle      = "int"
	floatStyle1dp = "float_1dp"
	floatStyle2dp = "float_2dp"
	floatStyle3dp = "float_3dp"
	headerStyle   = "header"
	titleStyle    = "title"
	defaultStyle  = "default"
)

var (
	columnNames = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterCount = len(columnNames)

	integerPattern       = regexp.MustCompile("^[1-9]+[0-9]*$")
	floatPattern         = regexp.MustCompile("^[0-9]*\\.[0-9]+$")
	decimalPlacesPattern = regexp.MustCompile("\\.[0-9]+")

	styleDefinitions = map[string]string{
		defaultStyle:  `{"number_format": 0}`,
		intStyle:      `{"number_format": 1}`,
		floatStyle1dp: `{"custom_number_format": "0.0"}`,
		floatStyle2dp: `{"number_format": 2}`,
		floatStyle3dp: `{"custom_number_format": "0.000"}`,
		headerStyle:   `{"alignment":{"wrap_text":true}, "font":{"bold":true} }`,
		titleStyle:    `{"font":{"bold":true} }`,
	}
)

type spreadsheetModel struct {
	xlsx         *excelize.File
	styleMap     map[string]int
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
		xlsx:       xlsx,
		styleMap:   createCellStyles(xlsx),
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
	xlsx.SetCellStyle(model.sheet, axisRef, axisRef, model.styleMap[titleStyle])
	model.currentRow++

	axisRef = getAxisRef(model.currentRow, 0)
	xlsx.SetCellStr(model.sheet, axisRef, request.Subtitle)
	xlsx.SetCellStyle(model.sheet, axisRef, axisRef, model.styleMap[titleStyle])
	model.currentRow++
}

// insertData inserts each cell of the table in the spreadsheet, unless hidden by a merged cell
func insertData(model *spreadsheetModel) {
	xlsx := model.xlsx
	tableModel := model.tableModel
	model.firstDataRow = model.currentRow + 1

	for r, row := range model.request.Data {
		model.currentRow ++
		for c, col := range row {
			isVisible := tableModel.cells[r][c] == nil || tableModel.cells[r][c].skip == false
			if isVisible {
				value, style := parseCellValue(col, model.styleMap)
				axisRef := getAxisRef(model.currentRow, c)
				xlsx.SetCellValue(model.sheet, axisRef, value)
				if tableModel.rows[r].Heading || tableModel.columns[c].Heading {
					xlsx.SetCellStyle(model.sheet, axisRef, axisRef, model.styleMap[headerStyle])
				} else if style > 0 {
					xlsx.SetCellStyle(model.sheet, axisRef, axisRef, style)
				}
			}
		}
	}
	model.currentRow ++
}

// insertSource inserts the source in the spreadsheet
func insertSource(model *spreadsheetModel) {
	xlsx := model.xlsx
	if len(model.request.Source) > 0 {
		model.currentRow ++
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 0), sourceText)
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 1), model.request.Source)
	}
}

// insertFootnotes inserts footnotes in the spreadsheet
func insertFootnotes(model *spreadsheetModel) {
	xlsx := model.xlsx
	request := model.request

	if len(request.Footnotes) > 0 {
		model.currentRow ++
		xlsx.SetCellStr(model.sheet, getAxisRef(model.currentRow, 0), notesText)
		for i, note := range request.Footnotes {
			model.currentRow ++
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
				rowspan --
			}
			colspan := format.Colspan
			if colspan > 0 {
				colspan --
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

// parseCellValue parses the value as an integer or a float if appropriate, returning the parsed value and appropriate excel style
func parseCellValue(value string, styleMap map[string]int) (interface{}, int) {
	var result interface{}
	var err error
	style := 0
	if integerPattern.MatchString(value) {
		result, err = strconv.Atoi(value)
		style = styleMap[intStyle]
	} else if floatPattern.MatchString(value) {
		result, err = strconv.ParseFloat(value, 64)
		switch len(decimalPlacesPattern.FindString(value)) - 1 {
		case 1:
			style = styleMap[floatStyle1dp]
		case 2:
			style = styleMap[floatStyle2dp]
		case 3:
			style = styleMap[floatStyle3dp]
		default:
			style = styleMap[defaultStyle]
		}
	} else {
		result = value
	}
	if err != nil {
		log.Error(err, log.Data{"_message": "Unable to parse value", "value": value})
		result = value
		style = 0
	}
	return result, style
}

// createCellStyles creates styles numbers and text the given spreadsheet, returning a map with those styles
func createCellStyles(xlsx *excelize.File) map[string]int {
	// TODO: refactor so that styles are keyed by the full alignment definition and can be added to to create new definitions
	// setting a style replaces any style already defined for that cell, so all style definitions must include number format, font, horizontal and vertical alignment
	styles := make(map[string]int)
	for key, value := range styleDefinitions {
		style, err := xlsx.NewStyle(value)
		if err != nil {
			log.Error(err, log.Data{"_message": "Unable create " + key + " style for spreadsheet"})
		} else {
			styles[key] = style
		}
	}
	return styles
}
