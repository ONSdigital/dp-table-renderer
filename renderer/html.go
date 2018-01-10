package renderer

import (
	"bytes"
	"fmt"

	"regexp"

	"strings"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/go-ns/log"
)

var (
	newLine      = regexp.MustCompile(`\n`)
	footnoteLink = regexp.MustCompile(`\[[0-9+]\]`)
)

// Contains details of the table that need to be calculated once from the request and cached
type tableModel struct {
	request *models.RenderRequest
	columns []models.ColumnFormat
}

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(request *models.RenderRequest) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<div class=\"table-renderer\" id=\"table_%s\">\n", request.Filename)

	model := createModel(request)

	startTable(request, &buf)
	writeColumnGroup(model, &buf)
	writeTableBody(model, &buf)
	buf.WriteString("</table>\n")

	writeFooter(request, &buf)

	buf.WriteString("</div>\n")
	return buf.Bytes(), nil
}

// write the table and caption tags
func startTable(request *models.RenderRequest, buf *bytes.Buffer) {
	caption := request.Title
	if len(request.Subtitle) > 0 {
		caption += fmt.Sprintf("<br /><span id=\"table_%s_description\" class=\"table_subtitle\">%s</span>", request.Filename, parseValue(request, request.Subtitle))
		fmt.Fprintf(buf, "<table aria-describedby=\"table_%s_description\">\n", request.Filename)
	} else {
		fmt.Fprintf(buf, "<table>\n")
	}
	if len(caption) > 0 {
		fmt.Fprintf(buf, "<caption>%s</caption>\n", parseValue(request, caption))
	}
}

// writes the colgroup element if necessary, applying style to columns
func writeColumnGroup(model *tableModel, buf *bytes.Buffer) {
	if len(model.request.ColumnFormats) > 0 {
		buf.WriteString("<colgroup>\n")
		for _, col := range model.columns {
			buf.WriteString("<col")
			if len(col.Align) > 0 {
				fmt.Fprintf(buf, " class=\"%s\"", col.Align)
			}
			if len(col.Width) > 0 {
				fmt.Fprintf(buf, " style=\"width: %s\"", col.Width)
			}
			buf.WriteString(" />")
		}
		buf.WriteString("</colgroup>\n")
	}
}

// write the rows of the table
func writeTableBody(model *tableModel, buf *bytes.Buffer) {
	request := model.request
	for _, row := range request.Data {
		buf.WriteString("<tr>")
		for i, col := range row {
			if model.columns[i].Heading {
				fmt.Fprintf(buf, "<th scope=\"row\">%s</th>", parseValue(request, col))
			} else {
				fmt.Fprintf(buf, "<td>%s</td>", parseValue(request, col))
			}
		}
		buf.WriteString("</tr>\n")
	}
}

// write a footer element with Source and footnotes
func writeFooter(request *models.RenderRequest, buf *bytes.Buffer) {
	buf.WriteString("<footer>\n")
	if len(request.Source) > 0 {
		fmt.Fprintf(buf, "<p class=\"table-source\">Source: %s</p>\n", parseValue(request, request.Source))
	}
	if len(request.Footnotes) > 0 {
		fmt.Fprintf(buf, "<p class=\"table-notes\" id=\"table_%s_notes\">Notes</p>\n", request.Filename)
		buf.WriteString("<ol>\n")
		for i, note := range request.Footnotes {
			fmt.Fprintf(buf, "<li id=\"table_%s_note_%d\">%s</li>\n", request.Filename, i+1, parseValue(request, note))
		}
		buf.WriteString("</ol>\n")
	}
	buf.WriteString("</footer>\n")
}

// Replaces \n with <br /> and wraps [1] with a link to the footnote
func parseValue(request *models.RenderRequest, value string) string {
	value = newLine.ReplaceAllLiteralString(value, "<br />")
	if len(request.Footnotes) > 0 && footnoteLink.MatchString(value) {
		for i := range request.Footnotes {
			n := i + 1
			linkText := fmt.Sprintf("<a aria-describedby=\"table_%s_notes\" href=\"#table_%s_note_%d\" class=\"footnote-link\">[%d]</a>", request.Filename, request.Filename, n, n)
			value = strings.Replace(value, fmt.Sprintf("[%d]", n), linkText, -1)
		}
	}
	return value
}


// Creates a tableModel containing calculations that are referenced more than once while rendering the table
func createModel(request *models.RenderRequest) *tableModel {
	m := tableModel{request:request}
	m.columns = indexColumnFormats(request)
	return &m
}
// indexes the ColumnFormats so that columns[i] gives the correct format for column i
func indexColumnFormats(request *models.RenderRequest) []models.ColumnFormat {
	// find the maximum number of columns in the data - should be the same in every row, but don't trust that
	count := 0
	for i, _ := range request.Data {
		n := len(request.Data[i])
		if (n > count) {
			count = n
		}
	}
	// create default ColumnFormats
	columns := make([]models.ColumnFormat, count)
	for i,_ := range columns {
		columns[i] = models.ColumnFormat{Column:i}
	}
	// replace with actual ColumnFormats where they exist
	for _, format := range request.ColumnFormats {
		if (format.Column >= count) {
			log.Debug("ColumnFormat specified for non-existent column", log.Data{"filename": request.Filename, "ColumnFormat": format, "column_count": count})
		} else {
			columns[format.Column] = format
		}
	}
	return columns
}
