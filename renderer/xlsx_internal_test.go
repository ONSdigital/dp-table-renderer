package renderer

import (
	"context"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ONSdigital/dp-table-renderer/models"
	. "github.com/smartystreets/goconvey/convey"
)

// it's not possible to extract details of cell styling (alignment, font, number format) from an xlsx file in any sensible fashion
// so this test exists to test some of the internal methods of xlsx.go. It's not ideal, but better than no tests relating to style.

var mockContext = context.TODO()

func TestCellStyles(t *testing.T) {
	t.Parallel()

	Convey("Cells should have the correct style applied", t, func() {
		data := [][]string{
			{"Cell 1A", "Cell 1B", "Cell 1C", "Cell 1D"},
			{"Cell 2A", "Cell 2B", "Cell 2C", "Cell 2D"},
			{"1", "1.00", "1.001", "1.1", "1.1234567"}}

		rowFormats := []models.RowFormat{
			{Row: 0, Heading: true, VerticalAlign: "Top"},
			{Row: 1, VerticalAlign: "Top"}}
		colFormats := []models.ColumnFormat{
			{Column: 0, Heading: true, Align: "Center"},
			{Column: 1, Align: "Center"}}
		cellFormats := []models.CellFormat{
			{Row: 0, Column: 0, VerticalAlign: "Bottom", Align: "Left"},
			{Row: 0, Column: 1, VerticalAlign: "Middle"},
			{Row: 1, Column: 1, VerticalAlign: "Bottom", Align: "Left"}}
		request := &models.RenderRequest{Filename: "filename", Data: data, RowFormats: rowFormats, ColumnFormats: colFormats, CellFormats: cellFormats}

		model := &spreadsheetModel{
			request:    request,
			tableModel: createModel(mockContext, request),
			cellStyles: make(map[xlsxCellStyle]int),
			xlsx:       excelize.NewFile(),
			currentRow: 0,
			sheet:      "Sheet1",
		}

		style := invokeGetCellValueAndStyle(model, 0, 0)
		So(style.Alignment.Vertical, ShouldBeEmpty) // bottom alignment is the default, so should be an empty string
		So(style.Alignment.Horizontal, ShouldEqual, "left")
		So(style.Font.Bold, ShouldBeTrue)

		style = invokeGetCellValueAndStyle(model, 0, 1)
		So(style.Alignment.Vertical, ShouldEqual, "center")
		So(style.Alignment.Horizontal, ShouldEqual, "center")
		So(style.Font.Bold, ShouldBeTrue)

		style = invokeGetCellValueAndStyle(model, 0, 2)
		So(style.Alignment.Vertical, ShouldEqual, "top")
		So(style.Alignment.Horizontal, ShouldBeEmpty)
		So(style.Font.Bold, ShouldBeTrue)

		style = invokeGetCellValueAndStyle(model, 1, 0)
		So(style.Alignment.Vertical, ShouldEqual, "top")
		So(style.Alignment.Horizontal, ShouldEqual, "center")
		So(style.Font.Bold, ShouldBeTrue)

		style = invokeGetCellValueAndStyle(model, 1, 1)
		So(style.Alignment.Vertical, ShouldBeEmpty) // bottom alignment is the default, so should be an empty string
		So(style.Alignment.Horizontal, ShouldEqual, "left")
		So(style.Font.Bold, ShouldBeFalse)

		style = invokeGetCellValueAndStyle(model, 1, 2)
		So(style.Alignment.Vertical, ShouldEqual, "top")
		So(style.Alignment.Horizontal, ShouldBeEmpty)
		So(style.Font.Bold, ShouldBeFalse)
		So(style.CustomNumberFormat, ShouldBeEmpty)
		So(style.NumberFormat, ShouldBeZeroValue)

		style = invokeGetCellValueAndStyle(model, 2, 2)
		So(style.Alignment.Vertical, ShouldBeEmpty)
		So(style.Alignment.Horizontal, ShouldBeEmpty)
		So(style.Font.Bold, ShouldBeFalse)

		style = invokeGetCellValueAndStyle(model, 2, 0)
		So(style.CustomNumberFormat, ShouldBeEmpty)
		So(style.NumberFormat, ShouldEqual, 1)

		style = invokeGetCellValueAndStyle(model, 2, 1)
		So(style.CustomNumberFormat, ShouldBeEmpty)
		So(style.NumberFormat, ShouldEqual, 2)

		style = invokeGetCellValueAndStyle(model, 2, 2)
		So(style.CustomNumberFormat, ShouldEqual, "0.000")
		So(style.NumberFormat, ShouldBeZeroValue)

		style = invokeGetCellValueAndStyle(model, 2, 3)
		So(style.CustomNumberFormat, ShouldEqual, "0.0")
		So(style.NumberFormat, ShouldBeZeroValue)

		style = invokeGetCellValueAndStyle(model, 2, 4)
		So(style.CustomNumberFormat, ShouldBeEmpty)
		So(style.NumberFormat, ShouldBeZeroValue)

	})

}

func invokeGetCellValueAndStyle(model *spreadsheetModel, row int, col int) xlsxCellStyle {
	_, styleIndex := getCellValueAndStyle(mockContext, model, row, col)
	style := xlsxCellStyle{}
	for key, value := range model.cellStyles {
		if value == styleIndex {
			style = key
		}
	}
	return style
}
