package renderer

import (
	"bytes"
	"fmt"

	"github.com/ONSdigital/dp-table-renderer/models"
)

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(request *models.RenderRequest) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("<div class=\"table-renderer\">\n")

	startTable(request, &buf)

	writeTableBody(request, &buf)

	buf.WriteString("</table>\n")
	buf.WriteString("</div>\n")
	return buf.Bytes(), nil
}

func startTable(request *models.RenderRequest, buf *bytes.Buffer) {
	caption := request.Title
	if len(request.Subtitle) > 0 {
		caption += fmt.Sprintf("<br/><span id=\"table_%s_description\">%s</span>", request.Filename, request.Subtitle)
		fmt.Fprintf(buf, "<table id=\"table_%s\" aria-describedby=\"table_%s_description\">\n", request.Filename, request.Filename)
	} else {
		fmt.Fprintf(buf, "<table id=\"table_%s\">\n", request.Filename)
	}
	if len(caption) > 0 {
		fmt.Fprintf(buf, "<caption>%s</caption>\n", caption)
	}
}

func writeTableBody(request *models.RenderRequest, buf *bytes.Buffer) {
	for _, row := range request.Data {
		buf.WriteString("<tr>")
		for _, col := range row {
			fmt.Fprintf(buf, "<td>%s</td>", col)
		}
		buf.WriteString("</tr>\n")
	}
}
