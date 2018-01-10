package renderer

import (
	"bytes"
	"fmt"

	"regexp"

	"strings"

	"github.com/ONSdigital/dp-table-renderer/models"
)

var (
	newLine      = regexp.MustCompile(`\n`)
	footnoteLink = regexp.MustCompile(`\[[0-9+]\]`)
)

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(request *models.RenderRequest) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<div class=\"table-renderer\" id=\"table_%s\">\n", request.Filename)

	startTable(request, &buf)
	writeTableBody(request, &buf)
	buf.WriteString("</table>\n")

	writeFooter(request, &buf)

	buf.WriteString("</div>\n")
	return buf.Bytes(), nil
}

func startTable(request *models.RenderRequest, buf *bytes.Buffer) {
	caption := request.Title
	if len(request.Subtitle) > 0 {
		caption += fmt.Sprintf("<br/><span id=\"table_%s_description\" class=\"table_subtitle\">%s</span>", request.Filename, parseValue(request, request.Subtitle))
		fmt.Fprintf(buf, "<table aria-describedby=\"table_%s_description\">\n", request.Filename)
	} else {
		fmt.Fprintf(buf, "<table>\n")
	}
	if len(caption) > 0 {
		fmt.Fprintf(buf, "<caption>%s</caption>\n", parseValue(request, caption))
	}
}

func writeTableBody(request *models.RenderRequest, buf *bytes.Buffer) {
	for _, row := range request.Data {
		buf.WriteString("<tr>")
		for _, col := range row {
			fmt.Fprintf(buf, "<td>%s</td>", parseValue(request, col))
		}
		buf.WriteString("</tr>\n")
	}
}

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

// Replaces \n with <br/> and wraps [1] with a link to the footnote
func parseValue(request *models.RenderRequest, value string) string {
	value = newLine.ReplaceAllLiteralString(value, "<br/>")
	if len(request.Footnotes) > 0 && footnoteLink.MatchString(value) {
		for i := range request.Footnotes {
			n := i + 1
			linkText := fmt.Sprintf("<a aria-describedby=\"table_%s_notes\" href=\"#table_%s_note_%d\" class=\"footnote-link\">[%d]</a>", request.Filename, request.Filename, n, n)
			value = strings.Replace(value, fmt.Sprintf("[%d]", n), linkText, -1)
		}
	}
	return value
}
