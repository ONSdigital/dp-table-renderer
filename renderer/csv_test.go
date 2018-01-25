package renderer_test

import (
	"bytes"
	"testing"

	"encoding/csv"
	"fmt"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/ONSdigital/dp-table-renderer/testdata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRenderCSV(t *testing.T) {
	t.Parallel()
	Convey("A csv should be rendered without error", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		request, err := models.CreateRenderRequest(reader)
		if err != nil {
			t.Fatal(err)
		}

		rows := invokeRenderCSV(request)

		So(len(rows), ShouldBeGreaterThanOrEqualTo, len(request.Data)+len(request.Footnotes))
	})

	Convey("A csv should be correctly formatted", t, func() {
		data := [][]string{
			{"Cell 1", "Cell 2", "Cell 3"},
			{"01", "10", "0.01", "23.45"},
			{"Cell 1", "Cell 2", "Cell 3", "Cell 4", "Cell 5"}}
		notes := []string{"Note 1", "Note 2 - this is a slightly longer note"}
		request := models.RenderRequest{Filename: "filename",
			Title:     "This is the Heading",
			Subtitle:  "This is a Subtitle",
			Source:    "Office of National Statistics",
			Data:      data,
			Footnotes: notes}

		rows := invokeRenderCSV(&request)

		So(rows[0][0], ShouldEqual, request.Title)
		So(rows[1][0], ShouldEqual, request.Subtitle)

		rowOffset := getCSVDataRowOffset(&request)
		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Data))
		for r := 0; r < len(request.Data); r++ {
			row := rows[r+rowOffset]
			So(len(row), ShouldBeGreaterThanOrEqualTo, len(request.Data[r]))
			for c, data := range request.Data[r] {
				So(row[c], ShouldEqual, data)
			}
		}

		rowOffset += len(request.Data)

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

	Convey("Footnotes in a csv should be correctly numbered", t, func() {
		data := [][]string{{"Cell 1"}}
		notes := []string{"Note 1", "Note 2 - this is a slightly longer note"}
		request := models.RenderRequest{Filename: "filename",
			Title:     "This is the Heading",
			Subtitle:  "This is a Subtitle",
			Source:    "Office of National Statistics",
			Data:      data,
			Footnotes: notes}

		rows := invokeRenderCSV(&request)

		So(rows[0][0], ShouldEqual, request.Title)
		So(rows[1][0], ShouldEqual, request.Subtitle)

		rowOffset := getCSVDataRowOffset(&request)
		rowOffset += len(request.Data) + 1

		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Footnotes)+1)
		So(rows[rowOffset][0], ShouldEqual, "Notes")
		for i, note := range request.Footnotes {
			So(rows[rowOffset+i+1][0], ShouldEqual, fmt.Sprintf("%d", i+1))
			So(rows[rowOffset+i+1][1], ShouldEqual, note)
		}
	})

	Convey("Cells hidden by a merge should not be present in the csv", t, func() {
		data := [][]string{
			{"Cell 1A", "hidden", "Cell 1C", "Cell 1D"},
			{"hidden", "hidden", "Cell 2C", "Cell 2D"},
			{"Cell 3A", "Cell 3B", "Cell 3C", "Cell 3D"}}
		formats := []models.CellFormat{
			{Row: 0, Column: 0, Colspan: 2, Rowspan: 2},
			{Row: 0, Column: 1, Colspan: 1, Rowspan: 1}}
		request := models.RenderRequest{Filename: "filename",
			Title:       "This is the Heading",
			Subtitle:    "This is a Subtitle",
			Data:        data,
			CellFormats: formats}

		rows := invokeRenderCSV(&request)

		rowOffset := getCSVDataRowOffset(&request)
		So(len(rows)-rowOffset, ShouldBeGreaterThanOrEqualTo, len(request.Data))
		So(rows[rowOffset][1], ShouldBeEmpty)
		So(rows[rowOffset+1][0], ShouldBeEmpty)
		So(rows[rowOffset+1][1], ShouldBeEmpty)

		So(rows[rowOffset][3], ShouldEqual, data[0][3])
		So(rows[rowOffset+1][2], ShouldEqual, data[1][2])
	})
}

func invokeRenderCSV(request *models.RenderRequest) [][]string {
	resultBytes, e := renderer.RenderCSV(request)
	So(e, ShouldBeNil)

	csv := csv.NewReader(bytes.NewReader(resultBytes))
	csv.FieldsPerRecord = -1
	rows, e := csv.ReadAll()
	So(e, ShouldBeNil)
	return rows
}

// when reading a csv file, empty lines are ignored - this adjusts the row offset accordingly
func getCSVDataRowOffset(request *models.RenderRequest) int {
	return getDataRowOffset(request) - 1
}
