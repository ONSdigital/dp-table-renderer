package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ONSdigital/log.go/v2/log"
)

// A list of errors returned from package
var (
	ErrorReadingBody = errors.New("Failed to read message body")
	ErrorParsingBody = errors.New("Failed to parse json body")
	ErrorNoData      = errors.New("Bad request - Missing data in body")
)

// valid values for alignments in the various formats
var (
	AlignTop     = "Top"
	AlignMiddle  = "Middle"
	AlignBottom  = "Bottom"
	AlignLeft    = "Left"
	AlignCenter  = "Center"
	AlignRight   = "Right"
	AlignJustify = "Justify"
)

// RenderRequest represents a structure for a table render job
type RenderRequest struct {
	Title               string         `json:"title,omitempty"`
	Subtitle            string         `json:"subtitle,omitempty"`
	Source              string         `json:"source,omitempty"`
	TableType           string         `json:"type,omitempty"`
	TableVersion        string         `json:"type_version,omitempty"`
	Filename            string         `json:"filename,omitempty"`
	Units               string         `json:"units,omitempty"`
	KeepHeadersTogether bool           `json:"keep_headers_together"`
	RowFormats          []RowFormat    `json:"row_formats"`
	ColumnFormats       []ColumnFormat `json:"column_formats"`
	CellFormats         []CellFormat   `json:"cell_formats"`
	Data                [][]string     `json:"data"`
	Footnotes           []string       `json:"footnotes"`
}

// ParseRequest represents a request to convert an html table (plus supporting data) into the correct RenderRequest format
type ParseRequest struct {
	Title               string          `json:"title"`
	Subtitle            string          `json:"subtitle"`
	Source              string          `json:"source"`
	Filename            string          `json:"filename"`
	Units               string          `json:"units"`
	KeepHeadersTogether bool            `json:"keep_headers_together"`
	Footnotes           []string        `json:"footnotes"`
	TableHTML           string          `json:"table_html"`
	IgnoreFirstRow      bool            `json:"ignore_first_row"`       // if true, the first row is ignored
	IgnoreFirstColumn   bool            `json:"ignore_first_column"`    // if true, the first cell of each row is ignored
	HeaderRows          int             `json:"header_rows"`            // the number of header rows (th cells) in the output, after ignoring the first row (if applicable)
	HeaderCols          int             `json:"header_cols"`            // the number of header columns (th cells) in each row of the output, after ignoring the first column (if applicable)
	CurrentTableWidth   int             `json:"current_table_width"`    // used to convert column width from pixels to %
	CurrentTableHeight  int             `json:"current_table_height"`   // used to convert row height from pixels to %
	SingleEmHeight      float32         `json:"single_em_height"`       // used to convert height/width from pixels to em. The height of the following: <div style="display: none; font-size: 1em; margin: 0; padding:0; height: auto; line-height: 1; border:0;">m</div>
	CellSizeUnits       string          `json:"cell_size_units"`        // 'em', '%' or 'auto' - the desired unit for widths/heights. Auto causes no widths/heights to be specified
	ColumnWidthToIgnore string          `json:"column_width_to_ignore"` // if the source html applies a default column width that shouldn't be included in the output, specify it here. e.g. '50px'
	AlignmentClasses    ParseAlignments `json:"alignment_classes"`      // The names of classes that should be interpreted as defining alignment of cells
}

// ParseAlignments defines the css classes that should be interpreted as defining the alignment of cells in a table
type ParseAlignments struct {
	Top     string `json:"top"`
	Middle  string `json:"middle"`
	Bottom  string `json:"bottom"`
	Left    string `json:"left"`
	Right   string `json:"right"`
	Center  string `json:"center"`
	Justify string `json:"justify"`
}

// RowFormat allows us to specify that a row contains headings, and provide a style for html
type RowFormat struct {
	Row           int    `json:"row"`                      // the index of the row the format applies to
	VerticalAlign string `json:"vertical_align,omitempty"` // must be Top, Middle or Bottom to be applied
	Heading       bool   `json:"heading,omitempty"`
	Height        string `json:"height,omitempty"`
}

// ColumnFormat allows us to specify that a column contains headings, specify alignment and provide a style for html
type ColumnFormat struct {
	Column  int    `json:"col"`             // the index of the column the format applies to
	Align   string `json:"align,omitempty"` // must be Left, Center or Right to be applied
	Heading bool   `json:"heading,omitempty"`
	Width   string `json:"width,omitempty"`
}

// CellFormat allows us to specify alignment and how to merge cells
type CellFormat struct {
	Row           int    `json:"row"`
	Column        int    `json:"col"`
	Align         string `json:"align,omitempty"`          // must be Left, Center or Right to be applied
	VerticalAlign string `json:"vertical_align,omitempty"` // must be Top, Middle or Bottom to be applied
	Rowspan       int    `json:"rowspan,omitempty"`
	Colspan       int    `json:"colspan,omitempty"`
}

// CreateRenderRequest manages the creation of a RenderRequest from a reader
func CreateRenderRequest(ctx context.Context, reader io.Reader) (*RenderRequest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(ctx, "error reading request body", err)
		return nil, ErrorReadingBody
	}

	var request RenderRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		log.Error(ctx, "error unmarshalling JSON", err)
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

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}

// CreateParseRequest manages the creation of a ParseRequest from a reader
func CreateParseRequest(ctx context.Context, reader io.Reader) (*ParseRequest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(ctx, "error reading body", err)
		return nil, ErrorReadingBody
	}

	var request ParseRequest
	err = json.Unmarshal(bytes, &request)
	if err != nil {
		log.Error(ctx, "error unmarshalling JSON", err)
		return nil, ErrorParsingBody
	}

	// This should be the last check before returning filter
	if len(bytes) == 2 {
		return &request, ErrorNoData
	}

	return &request, nil
}

// ValidateParseRequest checks the content of the request structure
func (pr *ParseRequest) ValidateParseRequest(ctx context.Context) error {

	var missingFields []string

	if len(pr.TableHTML) == 0 {
		missingFields = append(missingFields, "table_html")
	}

	switch units := pr.CellSizeUnits; units {
	case "%":
		if pr.CurrentTableWidth <= 0 {
			log.Info(ctx, "size_units is 'percentage' but current_table_width is not specified - cannot convert from px", log.Data{"file_name": pr.Filename})
		}
	case "em":
		if pr.SingleEmHeight <= 0 {
			log.Info(ctx, "size_units is 'percentage' but current_table_width is not specified - cannot convert from px", log.Data{"file_name": pr.Filename})
		}
	case "auto", "":
		// nothing to do
	default:
		log.Info(ctx, "unknown size unit specified for width", log.Data{"file_name": pr.Filename, "unit": units})
	}

	if missingFields != nil {
		return fmt.Errorf("Missing mandatory fields: %v", missingFields)
	}

	return nil
}
