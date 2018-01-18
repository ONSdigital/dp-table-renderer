package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/go-ns/log"
)

// A list of errors returned from package
var (
	ErrorReadingBody = errors.New("Failed to read message body")
	ErrorParsingBody = errors.New("Failed to parse json body")
	ErrorNoData      = errors.New("Bad request - Missing data in body")
)

// valid values for alignments in the various formats
var (
	AlignTop    = "Top"
	AlignMiddle = "Middle"
	AlignBottom = "Bottom"
	AlignLeft   = "Left"
	AlignCenter = "Center"
	AlignRight  = "Right"
)

// RenderRequest represents a structure for a table render job
type RenderRequest struct {
	Title         string         `json:"title"`
	Subtitle      string         `json:"subtitle"`
	Source        string         `json:"source"`
	TableType     string         `json:"type"`
	TableVersion  string         `json:"type_version"`
	Filename      string         `json:"filename"`
	Units         string         `json:"units"`
	RowFormats    []RowFormat    `json:"row_formats"`
	ColumnFormats []ColumnFormat `json:"column_formats"`
	CellFormats   []CellFormat   `json:"cell_formats"`
	Data          [][]string     `json:"data"`
	Footnotes     []string       `json:"footnotes"`
}

// ParseRequest represents a request to convert an html table (plus supporting data) into the correct RenderRequest format
type ParseRequest struct {
	Title               string          `json:"title"`
	Subtitle            string          `json:"subtitle"`
	Source              string          `json:"source"`
	Filename            string          `json:"filename"`
	Units               string          `json:"units"`
	Footnotes           []string        `json:"footnotes"`
	TableHTML           string          `json:"table_html"`
	IgnoreFirstRow      bool            `json:"ignore_first_row"`       // if true, the first row is ignored
	IgnoreFirstColumn   bool            `json:"ignore_first_column"`    // if true, the first cell of each row is ignored
	HeaderRows          int             `json:"header_rows"`            // the number of header rows (th cells) in the output, after ignoring the first row (if applicable)
	HeaderCols          int             `json:"header_cols"`            // the number of header columns (th cells) in each row of the output, after ignoring the first column (if applicable)
	CurrentTableWidth   int             `json:"current_table_width"`    // used to convert column width from pixels to %
	CurrentTableHeight  int             `json:"current_table_height"`   // used to convert row height from pixels to %
	SingleEmHeight      float32         `json:"single_em_height"`       // used to convert height/width from pixels to em. The height of the following: <div style="display: none; font-size: 1em; margin: 0; padding:0; height: auto; line-height: 1; border:0;">m</div>
	SizeUnits           string          `json:"size_units"`             // 'em' or '%' - the desired unit for widths/heights
	ColumnWidthToIgnore string          `json:"column_width_to_ignore"` // if the source html applies a default column width that shouldn't be included in the output, specify it here. e.g. '50px'
	AlignmentClasses    ParseAlignments `json:"alignment_classes"`      // The names of classes that should be interpreted as defining alignment of cells
}

// ParseAlignments defines the css classes that should be interpreted as defining the alignment of cells in a table
type ParseAlignments struct {
	Top    string `json:"top"`
	Middle string `json:"middle"`
	Bottom string `json:"bottom"`
	Left   string `json:"left"`
	Right  string `json:"right"`
	Center string `json:"center"`
}

// RowFormat allows us to specify that a row contains headings, and provide a style for html
type RowFormat struct {
	Row           int    `json:"row"`            // the index of the row the format applies to
	VerticalAlign string `json:"vertical_align"` // must be Top, Middle or Bottom to be applied
	Heading       bool   `json:"heading"`
	Height        string `json:"height"`
}

// ColumnFormat allows us to specify that a column contains headings, specify alignment and provide a style for html
type ColumnFormat struct {
	Column  int    `json:"col"`   // the index of the column the format applies to
	Align   string `json:"align"` // must be Left, Center or Right to be applied
	Heading bool   `json:"heading"`
	Width   string `json:"width"`
}

// CellFormat allows us to specify alignment and style, that a cell contains a heading, and how to merge cells
type CellFormat struct {
	Row           int    `json:"row"`
	Column        int    `json:"col"`
	Align         string `json:"align"`
	VerticalAlign string `json:"vertical_align"` // must be Top, Middle or Bottom to be applied
	Rowspan       int    `json:"rowspan"`
	Colspan       int    `json:"colspan"`
}

// CreateRenderRequest manages the creation of a RenderRequest from a reader
func CreateRenderRequest(reader io.Reader) (*RenderRequest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(err, log.Data{"request_body": string(bytes)})
		return nil, ErrorReadingBody
	}

	var request RenderRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		log.Error(err, log.Data{"request_body": string(bytes)})
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
		log.Error(err, log.Data{"request_body": string(bytes)})
		return nil, ErrorReadingBody
	}

	var request ParseRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		log.Error(err, log.Data{"request_body": string(bytes)})
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

	switch units := pr.SizeUnits; units {
	case "%":
		if pr.CurrentTableWidth <= 0 {
			log.InfoC(pr.Filename, "size_units is '%' but current_table_width is not specified - cannot convert from px", nil)
		}
	case "em":
		if pr.SingleEmHeight <= 0 {
			log.InfoC(pr.Filename, "size_units is 'em' but single_em_height is not specified - cannot convert from px", nil)
		}
	case "":
		// don't spam the logs
	default:
		log.InfoC(pr.Filename, "Unknown size unit specified for width: "+units, nil)
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}
