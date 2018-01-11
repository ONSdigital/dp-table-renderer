package renderer

import (
	"bytes"
	"fmt"

	"regexp"

	"strings"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/go-ns/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	newLine      = regexp.MustCompile(`\n`)
	footnoteLink = regexp.MustCompile(`\[[0-9+]]`)
)

// Contains details of the table that need to be calculated once from the request and cached
type tableModel struct {
	request *models.RenderRequest
	columns []models.ColumnFormat
}

// RenderHTML returns an HTML representation of the table generated from the given request
func RenderHTML(request *models.RenderRequest) ([]byte, error) {
	model := createModel(request)

	div := createNode("div", atom.Div,
		attr("class", "table-renderer"),
		attr("id", "table_"+request.Filename),
		"\n")

	table := addTable(request, div)

	addColumnGroup(model, table)
	addRows(model, table)

	addFooter(request, div)

	var buf bytes.Buffer
	html.Render(&buf, div)
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// createNode creates an html Node and sets attributes or adds child nodes according to the type of each value
func createNode(data string, dataAtom atom.Atom, values ...interface{}) *html.Node {
	node := &html.Node{
		Type:     html.ElementNode,
		Data:     data,
		DataAtom: dataAtom,
	}
	for _, value := range values {
		switch v := value.(type) {
		case html.Attribute:
			node.Attr = append(node.Attr, v)
		case *html.Node:
			node.AppendChild(v)
		case []*html.Node:
			for _, c := range v {
				node.AppendChild(c)
			}
		case string:
			node.AppendChild(&html.Node{Type: html.TextNode, Data: v})
		}
	}
	return node
}

func setAttribute(node *html.Node, key string, val string) {
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: val})
}

func attr(key string, val string) html.Attribute {
	return html.Attribute{Key: key, Val: val}
}

func text(text string) *html.Node {
	return &html.Node{Type: html.TextNode, Data: text}
}

// createTable creates a table node with a caption
func addTable(request *models.RenderRequest, parent *html.Node) *html.Node {
	table := createNode("table", atom.Table, "\n")

	// add title and subtitle as a caption
	if len(request.Title) > 0 || len(request.Subtitle) > 0 {
		caption := createNode("caption", atom.Caption, parseValue(request, request.Title))
		if len(request.Subtitle) > 0 {
			subtitleID := fmt.Sprintf("table_%s_description", request.Filename)
			subtitle := createNode("span", atom.Span,
				attr("id", subtitleID),
				attr("class", "table-subtitle"),
				parseValue(request, request.Subtitle))

			caption.AppendChild(createNode("br", atom.Br))
			caption.AppendChild(subtitle)

			setAttribute(table, "aria-describedby", subtitleID)
		}
		table.AppendChild(caption)
		table.AppendChild(text("\n"))
	}
	parent.AppendChild(table)
	return table
}

// addColumnGroup adds a columnGroup, if required, to the given table. Cols in the colgroup specify column width and alignment.
func addColumnGroup(model *tableModel, table *html.Node) {
	if len(model.request.ColumnFormats) > 0 {
		colgroup := createNode("colgroup", atom.Colgroup)

		for _, col := range model.columns {
			node := createNode("col", atom.Col)
			if len(col.Align) > 0 {
				setAttribute(node, "class", col.Align)
			}
			if len(col.Width) > 0 {
				setAttribute(node, "style", "width: "+col.Width)
			}
			colgroup.AppendChild(node)
		}

		table.AppendChild(colgroup)
		table.AppendChild(text("\n"))
	}
}

// adds all rows to the table. Rows contain th or td cells as appropriate.
func addRows(model *tableModel, table *html.Node) {
	request := model.request
	for _, row := range request.Data {
		tr := createNode("tr", atom.Tr)
		table.AppendChild(tr)
		for i, col := range row {
			if model.columns[i].Heading {
				tr.AppendChild(createNode("th", atom.Th, attr("scope", "row"), parseValue(request, col)))
			} else {
				tr.AppendChild(createNode("td", atom.Td, parseValue(request, col)))
			}
		}
		table.AppendChild(text("\n"))
	}
}

// addFooter adds a footer to the given element, containing the source and footnotes
func addFooter(request *models.RenderRequest, parent *html.Node) {
	footer := createNode("footer", atom.Footer, "\n")
	if len(request.Source) > 0 {
		footer.AppendChild(createNode("p", atom.P,
			attr("class", "table-source"),
			parseValue(request, "Source: "+request.Source)))
		footer.AppendChild(text("\n"))
	}
	if len(request.Footnotes) > 0 {
		footer.AppendChild(createNode("p", atom.P,
			attr("class", "table-notes"),
			attr("id", "table_"+request.Filename+"_notes"),
			"Notes"))
		footer.AppendChild(text("\n"))
		ol := createNode("ol", atom.Ol, "\n")

		for i, note := range request.Footnotes {
			ol.AppendChild(createNode("li", atom.Li,
				attr("id", fmt.Sprintf("table_%s_note_%d", request.Filename, i+1)),
				parseValue(request, note)))
			ol.AppendChild(text("\n"))
		}

		footer.AppendChild(ol)
		footer.AppendChild(text("\n"))
	}
	parent.AppendChild(footer)
	parent.AppendChild(text("\n"))
}

// Parses the string to replace \n with <br /> and wrap [1] with a link to the footnote
func parseValue(request *models.RenderRequest, value string) []*html.Node {
	hasBr := newLine.MatchString(value)
	hasFootnote := len(request.Footnotes) > 0 && footnoteLink.MatchString(value)
	if hasBr || hasFootnote {
		return replaceValues(request, value, hasBr, hasFootnote)
	}
	return []*html.Node{{Type: html.TextNode, Data: value}}
}

// replaceValues uses regexp to replace new lines and footnotes with <br/> and <a>.../<a> tags, then parses the result into an array of nodes
func replaceValues(request *models.RenderRequest, value string, hasBr bool, hasFootnote bool) []*html.Node {
	original := value
	if hasBr {
		value = newLine.ReplaceAllLiteralString(value, "<br />")
	}
	if hasFootnote {
		for i := range request.Footnotes {
			n := i + 1
			linkText := fmt.Sprintf("<a aria-describedby=\"table_%s_notes\" href=\"#table_%s_note_%d\" class=\"footnote-link\">[%d]</a>", request.Filename, request.Filename, n, n)
			value = strings.Replace(value, fmt.Sprintf("[%d]", n), linkText, -1)
		}
	}
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		log.ErrorC("Unable to parse value!", err, log.Data{"filename": request.Filename, "value": original})
		return []*html.Node{{Type: html.TextNode, Data: original}}
	}
	return nodes
}

// Creates a tableModel containing calculations that are referenced more than once while rendering the table
func createModel(request *models.RenderRequest) *tableModel {
	m := tableModel{request: request}
	m.columns = indexColumnFormats(request)
	return &m
}

// indexes the ColumnFormats so that columns[i] gives the correct format for column i
func indexColumnFormats(request *models.RenderRequest) []models.ColumnFormat {
	// find the maximum number of columns in the data - should be the same in every row, but don't trust that
	count := 0
	for i := range request.Data {
		n := len(request.Data[i])
		if n > count {
			count = n
		}
	}
	// create default ColumnFormats
	columns := make([]models.ColumnFormat, count)
	for i := range columns {
		columns[i] = models.ColumnFormat{Column: i}
	}
	// replace with actual ColumnFormats where they exist
	for _, format := range request.ColumnFormats {
		if format.Column >= count || format.Column < 0 {
			log.Debug("ColumnFormat specified for non-existent column", log.Data{"filename": request.Filename, "ColumnFormat": format, "column_count": count})
		} else {
			columns[format.Column] = format
		}
	}
	return columns
}
