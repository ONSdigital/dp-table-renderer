package renderer

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// RenderCSV returns a csv representation of the table generated from the given request
func RenderCSV(ctx context.Context, request *models.RenderRequest) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	model := createModel(ctx, request)

	err := writeTitles(ctx, writer, request)
	if err != nil {
		return nil, err
	}

	err = writeData(ctx, writer, model, request)
	if err != nil {
		return nil, err
	}

	err = writeUnits(ctx, writer, request)
	if err != nil {
		return nil, err
	}

	err = writeSource(ctx, writer, request)
	if err != nil {
		return nil, err
	}

	err = writeFootnotes(ctx, writer, request)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// writeTitles writes the title and subtitle to the csv
func writeTitles(ctx context.Context, writer *csv.Writer, request *models.RenderRequest) error {
	err := writeRow(writer, request.Title)
	if err != nil {
		log.Error(ctx, "unable to write title to csv", err, log.Data{"title": request.Title})
		return err
	}
	err = writeRow(writer, request.Subtitle)
	if err != nil {
		log.Error(ctx, "unable to write subtitle to csv", err, log.Data{"subtitle": request.Subtitle})
		return err
	}
	return writeEmptyLine(ctx, writer, request)
}

// writeData writes each row of the table to the csv writer, replacing cells hidden by a merge with an empty string
func writeData(ctx context.Context, writer *csv.Writer, model *tableModel, request *models.RenderRequest) error {
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
			log.Error(ctx, "unable to write row", err, log.Data{"row": r})
			return err
		}
	}
	return writeEmptyLine(ctx, writer, request)
}

// writeUnits writes the units as a row in the csv
func writeUnits(ctx context.Context, writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Units) > 0 {
		err := writeRow(writer, unitsText, request.Units)
		if err != nil {
			log.Error(ctx, "unable to write units", err, log.Data{"units": request.Units})
			return err
		}
	}
	return nil
}

// writeSource writes the source as a row in the csv
func writeSource(ctx context.Context, writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Source) > 0 {
		err := writeRow(writer, sourceText, request.Source)
		if err != nil {
			log.Error(ctx, "unable to write source", err, log.Data{"source": request.Source})
			return err
		}
	}
	return nil
}

// writeFootnotes writes each footnotes as a row in the csv
func writeFootnotes(ctx context.Context, writer *csv.Writer, request *models.RenderRequest) error {
	if len(request.Footnotes) > 0 {
		err := writeRow(writer, notesText)
		if err != nil {
			log.Error(ctx, "unable to write notes header", err)
			return err
		}
		for i, note := range request.Footnotes {
			err := writeRow(writer, fmt.Sprintf("%d", i+1), note)
			if err != nil {
				log.Error(ctx, "unable to write notes", err, log.Data{"notes": i})
				return err
			}
		}
	}
	return nil
}

// writeEmptyLine is a convenience method that writes an empty line and logs any error that occurs
func writeEmptyLine(ctx context.Context, writer *csv.Writer, request *models.RenderRequest) error {
	err := writeRow(writer, "")
	if err != nil {
		log.Error(ctx, "unable to write empty line to csv", err)
		return err
	}
	return nil
}

// writeRow is a convenience method accepting a variadic string instead of a slice
func writeRow(writer *csv.Writer, row ...string) error {
	return writer.Write(row)
}
