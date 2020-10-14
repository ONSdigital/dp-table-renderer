package renderer

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/log.go/log"
)

// RenderCSV returns a csv representation of the table generated from the given request
func RenderCSV(request *models.RenderRequest) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	model := createModel(request)

	err := writeTitles(writer, request)
	if err != nil {
		return nil, err
	}

	err = writeData(writer, model, request)
	if err != nil {
		return nil, err
	}

	err = writeUnits(writer, request)
	if err != nil {
		return nil, err
	}

	err = writeSource(writer, request)
	if err != nil {
		return nil, err
	}

	err = writeFootnotes(writer, request)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// writeTitles writes the title and subtitle to the csv
func writeTitles(writer *csv.Writer, request *models.RenderRequest) error {
	err := writeRow(writer, request.Title)
	if err != nil {
		log.Event(nil, fmt.Sprintf("unable to write title to csv: %s", request.Title), log.ERROR, log.Error(err))
		return err
	}
	err = writeRow(writer, request.Subtitle)
	if err != nil {
		log.Event(nil, fmt.Sprintf("unable to write subtitle to csv: %s", request.Subtitle), log.ERROR, log.Error(err))
		return err
	}
	return writeEmptyLine(writer, request)
}

// writeData writes each row of the table to the csv writer, replacing cells hidden by a merge with an empty string
func writeData(writer *csv.Writer, model *tableModel, request *models.RenderRequest) error {
	for r, row := range request.Data {
		out := []string{}
		for c, value := range row {
			if cellIsVisible(model, r, c) {
				out = append(out, value)
			} else {
				out = append(out, "")
			}
		}
		err := writer.Write(out)
		if err != nil {
			log.Event(nil, fmt.Sprintf("unable to write row %d", r), log.ERROR, log.Error(err))
			return err
		}
	}
	return writeEmptyLine(writer, request)
}

// writeUnits writes the units as a row in the csv
func writeUnits(writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Units) > 0 {
		err := writeRow(writer, unitsText, request.Units)
		if err != nil {
			log.Event(nil, fmt.Sprintf("unable to write units: %s", request.Units), log.ERROR, log.Error(err))
			return err
		}
	}
	return nil
}

// writeSource writes the source as a row in the csv
func writeSource(writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Source) > 0 {
		err := writeRow(writer, sourceText, request.Source)
		if err != nil {
			log.Event(nil, fmt.Sprintf("unable to write source: %s", request.Source), log.ERROR, log.Error(err))
			return err
		}
	}
	return nil
}

// writeFootnotes writes each footnotes as a row in the csv
func writeFootnotes(writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Footnotes) > 0 {
		err := writeRow(writer, notesText)
		if err != nil {
			log.Event(nil, fmt.Sprintf("unable to write notes header"), log.ERROR, log.Error(err))
			return err
		}
		for i, note := range request.Footnotes {
			err := writeRow(writer, fmt.Sprintf("%d", i+1), note)
			if err != nil {
				log.Event(nil, fmt.Sprintf("unable to write notes %d", i), log.ERROR, log.Error(err))
				return err
			}
		}
	}
	return nil
}

// writeEmptyLine is a convenience method that writes an empty line and logs any error that occurs
func writeEmptyLine(writer *csv.Writer, request *models.RenderRequest) error {
	err := writeRow(writer, "")
	if err != nil {
		log.Event(nil, fmt.Sprintf("unable to write empty line to csv"), log.ERROR, log.Error(err))
		return err
	}
	return nil
}

// writeRow is a convenience method accepting a variadic string instead of a slice
func writeRow(writer *csv.Writer, row ...string) error {
	return writer.Write(row)
}
