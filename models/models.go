package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// A list of errors returned from package
var (
	ErrorReadingBody = errors.New("Failed to read message body")
	ErrorParsingBody = errors.New("Failed to parse json body")
	ErrorNoData      = errors.New("Bad request - Missing data in body")
)

// RenderRequest represents a structure for a table render job
type RenderRequest struct {
	Title         string         `json:"title"`
	Subtitle      string         `json:"subtitle"`
	Source        string         `json:"source"`
	TableType     string         `json:"type"`
	Filename      string         `json:"filename"`
	URI           string         `json:"uri"`
	StyleClass    string         `json:"style_class"`
	RowFormats    []RowFormat    `json:"row_formats"`
	ColumnFormats []ColumnFormat `json:"column_formats"`
	CellFormats   []CellFormat   `json:"cell_formats"`
	Data          [][]string     `json:"data"`
	Footnotes     []string       `json:"footnotes"`
}

// ParseRequest represents a request to convert an html table (plus supporting data) into the correct RenderRequest format
type ParseRequest struct {
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle"`
	Source     string   `json:"source"`
	Filename   string   `json:"filename"`
	URI        string   `json:"uri"`
	Footnotes  []string `json:"footnotes"`
	TableHTML  string   `json:"table_html"`
	HeaderRows int      `json:"header_rows"`
	HeaderCols int      `json:"header_cols"`
}

// RowFormat allows us to specify that a row contains headings, and provide a style for html
type RowFormat struct {
	Row           int    `json:"row"` // the index of the row the format applies to
	VerticalAlign string `json:"vertical_align"`
	Heading       bool   `json:"heading"`
	Height        string `json:"height"`
}

// ColumnFormat allows us to specify that a column contains headings, specify alignment and provide a style for html
type ColumnFormat struct {
	Column  int    `json:"col"` // the index of the column the format applies to
	Align   string `json:"align"`
	Heading bool   `json:"heading"`
	Width   string `json:"width"`
}

// CellFormat allows us to specify alignment and style, that a cell contains a heading, and how to merge cells
type CellFormat struct {
	Row           int    `json:"row"`
	Column        int    `json:"col"`
	Align         string `json:"align"`
	VerticalAlign string `json:"vertical_align"`
	Rowspan       int    `json:"rowspan"`
	Colspan       int    `json:"colspan"`
}

// CreateRenderRequest manages the creation of a RenderRequest from a reader
func CreateRenderRequest(reader io.Reader) (*RenderRequest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ErrorReadingBody
	}

	var request RenderRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		return nil, ErrorParsingBody
	}

	// This should be the last check before returning RenderRequest
	if len(bytes) == 2 {
		return &request, ErrorNoData
	}

	return &request, nil
}

// ValidateRenderRequest checks the content of the request structure
func (rr *RenderRequest) ValidateRenderRequest() error {

	var missingFields []string

	if len(rr.Filename) == 0 {
		missingFields = append(missingFields, "filename")
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}

// CreateParseRequest manages the creation of a ParseRequest from a reader
func CreateParseRequest(reader io.Reader) (*ParseRequest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, ErrorReadingBody
	}

	var request ParseRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		return nil, ErrorParsingBody
	}

	// This should be the last check before returning filter
	if len(bytes) == 2 {
		return &request, ErrorNoData
	}

	return &request, nil
}

// ValidateParseRequest checks the content of the request structure
func (pr *ParseRequest) ValidateParseRequest() error {

	var missingFields []string

	if len(pr.Filename) == 0 {
		missingFields = append(missingFields, "filename")
	}
	if len(pr.TableHTML) == 0 {
		missingFields = append(missingFields, "table_html")
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}
