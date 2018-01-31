package renderer_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/ONSdigital/dp-table-renderer/testdata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRenderXLSX(t *testing.T) {
	t.Parallel()
	Convey("A Spreadsheet should be rendered without error", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		request, err := models.CreateRenderRequest(reader)
		if err != nil {
			t.Fatal(err)
		}

		resultBytes, e := renderer.RenderXLSX(request)
		So(e, ShouldBeNil)

		xlsx, e := excelize.OpenReader(bytes.NewReader(resultBytes))
		So(e, ShouldBeNil)
		sheetMap := xlsx.GetSheetMap()
		So(len(sheetMap), ShouldEqual, 1)

		// Get all the rows in Sheet1.
		rows := xlsx.GetRows(sheetMap[1])
		So(len(rows), ShouldBeGreaterThanOrEqualTo, len(request.Data))
	})

	Convey("A Spreadsheet should be correctly formatted", t, func() {
		data := [][]string{
			{"Cell 1", "Cell 2", "Cell 3"},
			{"01", "10", "0.01", "23.45"},
			{"Cell 1", "Cell 2", "Cell 3", "Cell 4", "Cell 5"}}
		notes := []string{"Note 1", "Note 2 - this is a slightly longer note"}
		request := models.RenderRequest{Filename: "filename",
			Title:     "This is the Heading",
			Subtitle:  "This is a Subtitle",
			Source:    "Office of National Statistics",
			Units:     "myUnits",
			Data:      data,
			Footnotes: notes}

		resultBytes, e := renderer.RenderXLSX(&request)
		So(e, ShouldBeNil)

		xlsx, e := excelize.OpenReader(bytes.NewReader(resultBytes))
		So(e, ShouldBeNil)
		sheetMap := xlsx.GetSheetMap()
		So(len(sheetMap), ShouldEqual, 1)

		// Get all the rows in Sheet1.
		rows := xlsx.GetRows(sheetMap[1])
		So(len(rows), ShouldBeGreaterThan, len(request.Data))

		So(rows[0][0], ShouldEqual, request.Title)
		So(rows[1][0], ShouldEqual, request.Subtitle)

		rowOffset := getDataRowOffset(&request)
		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Data))
		for r := 0; r < len(request.Data); r++ {
			row := rows[r+rowOffset]
			So(len(row), ShouldBeGreaterThanOrEqualTo, len(request.Data[r]))
			for c, data := range request.Data[r] {
				So(row[c], ShouldEqual, data)
			}
		}

		rowOffset += len(request.Data) + 1

		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, 1)
		So(rows[rowOffset][0], ShouldEqual, "Units: ")
		So(rows[rowOffset][1], ShouldEqual, request.Units)

		rowOffset++

		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, 1)
		So(rows[rowOffset][0], ShouldEqual, "Source: ")
		So(rows[rowOffset][1], ShouldEqual, request.Source)

		rowOffset++

		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Footnotes)+1)
		So(rows[rowOffset][0], ShouldEqual, "Notes")
		for i, note := range request.Footnotes {
			So(rows[rowOffset+i+1][1], ShouldEqual, note)
		}
	})

	Convey("Footnotes in a spreadsheet should be correctly numbered", t, func() {
		data := [][]string{{"Cell 1"}}
		notes := []string{"Note 1", "Note 2 - this is a slightly longer note"}
		request := models.RenderRequest{Filename: "filename",
			Title:     "This is the Heading",
			Subtitle:  "This is a Subtitle",
			Source:    "Office of National Statistics",
			Data:      data,
			Footnotes: notes}

		resultBytes, e := renderer.RenderXLSX(&request)
		So(e, ShouldBeNil)

		xlsx, e := excelize.OpenReader(bytes.NewReader(resultBytes))
		So(e, ShouldBeNil)
		sheetMap := xlsx.GetSheetMap()
		// Get all the rows in Sheet1.
		rows := xlsx.GetRows(sheetMap[1])
		So(len(rows), ShouldBeGreaterThan, len(request.Data))

		So(rows[0][0], ShouldEqual, request.Title)
		So(rows[1][0], ShouldEqual, request.Subtitle)

		rowOffset := getDataRowOffset(&request)
		rowOffset += len(request.Data) + 2

		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Footnotes)+1)
		So(rows[rowOffset][0], ShouldEqual, "Notes")
		for i, note := range request.Footnotes {
			So(rows[rowOffset+i+1][0], ShouldEqual, fmt.Sprintf("%d.", i+1))
			So(rows[rowOffset+i+1][1], ShouldEqual, note)
		}
	})

	Convey("A Spreadsheet should should handle more than 26 columns correctly", t, func() {
		data := [][]string{
			{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
				"aa", "ab", "ac", "ad", "ae", "af", "ag", "ah", "ai", "aj", "ak", "al", "am", "an", "ao", "ap", "aq", "ar", "as", "at", "au", "av", "aw", "ax", "ay", "az",
				"ba", "bb", "bc", "bd"}}
		request := models.RenderRequest{Filename: "filename", Data: data}

		resultBytes, e := renderer.RenderXLSX(&request)
		So(e, ShouldBeNil)

		xlsx, e := excelize.OpenReader(bytes.NewReader(resultBytes))
		So(e, ShouldBeNil)
		sheetMap := xlsx.GetSheetMap()
		So(len(sheetMap), ShouldEqual, 1)

		// Get all the rows in Sheet1.
		rows := xlsx.GetRows(sheetMap[1])
		So(len(rows), ShouldBeGreaterThanOrEqualTo, len(request.Data))

		So(rows[0][0], ShouldEqual, request.Title)
		So(rows[1][0], ShouldEqual, request.Subtitle)

		rowOffset := getDataRowOffset(&request)
		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Data))
		for r := 0; r < len(request.Data); r++ {
			row := rows[r+rowOffset]
			So(len(row), ShouldBeGreaterThanOrEqualTo, len(request.Data[r]))
			for c, data := range request.Data[r] {
				So(row[c], ShouldEqual, data)
			}
		}
	})

	Convey("Cells hidden by a merge should not be present in the spreadsheet", t, func() {
		data := [][]string{
			{"Cell 1A", "hidden", "Cell 1C", "Cell 1D"},
			{"hidden", "hidden", "Cell 2C", "Cell 2D"},
			{"Cell 3A", "Cell 3B", "Cell 3C", "Cell 3D"}}
		formats := []models.CellFormat{
			{Row: 0, Column: 0, Colspan: 2, Rowspan: 2},
			{Row: 0, Column: 1, Colspan: 1, Rowspan: 1}}
		request := models.RenderRequest{Filename: "filename", Data: data, CellFormats: formats}

		resultBytes, e := renderer.RenderXLSX(&request)
		So(e, ShouldBeNil)

		xlsx, e := excelize.OpenReader(bytes.NewReader(resultBytes))
		So(e, ShouldBeNil)
		sheetMap := xlsx.GetSheetMap()
		So(len(sheetMap), ShouldEqual, 1)

		// Get all the rows in Sheet1.
		rows := xlsx.GetRows(sheetMap[1])

		rowOffset := getDataRowOffset(&request)
		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Data))
		So(rows[rowOffset][1], ShouldBeEmpty)
		So(rows[rowOffset+1][0], ShouldBeEmpty)
		So(rows[rowOffset+1][1], ShouldBeEmpty)

		So(rows[rowOffset][3], ShouldEqual, data[0][3])
		So(rows[rowOffset+1][2], ShouldEqual, data[1][2])
	})

}

func getDataRowOffset(request *models.RenderRequest) int {
	offset := 3
	return offset
}
