package renderer

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/go-ns/log"
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

	// TODO: write units

	err = writeData(writer, model, request)
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
		log.ErrorC(request.Filename, err, log.Data{"_message": "Unable to write title to csv", "value": request.Title})
		return err
	}
	err = writeRow(writer, request.Subtitle)
	if err != nil {
		log.ErrorC(request.Filename, err, log.Data{"_message": "Unable to write Subtitle to csv", "value": request.Subtitle})
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
			log.ErrorC(request.Filename, err, log.Data{"_message": fmt.Sprintf("Unable to write row %d", r), "value": row})
			return err
		}
	}
	return writeEmptyLine(writer, request)
}

// writeSource writes the source as a row in the csv
func writeSource(writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Source) > 0 {
		err := writeRow(writer, sourceText, request.Source)
		if err != nil {
			log.ErrorC(request.Filename, err, log.Data{"_message": "Unable to write source", "value": request.Source})
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
			log.ErrorC(request.Filename, err, log.Data{"_message": "Unable to write notes header"})
			return err
		}
		for i, note := range request.Footnotes {
			err := writeRow(writer, fmt.Sprintf("%d", i+1), note)
			if err != nil {
				log.ErrorC(request.Filename, err, log.Data{"_message": fmt.Sprintf("Unable to write notes %d", i), "value": note})
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
		log.ErrorC(request.Filename, err, log.Data{"_message": "Unable to write empty line to csv"})
		return err
	}
	return nil
}

// writeRow is a convenience method accepting a variadic string instead of a slice
func writeRow(writer *csv.Writer, row ...string) error {
	return writer.Write(row)
}