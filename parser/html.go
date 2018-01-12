package parser

import (
	"encoding/json"

	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/go-ns/log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ResponseModel defines the format of the json response contained in the bytes returned from ParseHTML
type ResponseModel struct {
	JSON        models.RenderRequest `json:"json"`
	PreviewHTML string               `json:"preview_html"`
}

// ParseHTML parses the html table in the request and generates correctly formatted JSON
func ParseHTML(request *models.ParseRequest) ([]byte, error) {

	sourceTable, err := parseTableToNode(request.TableHTML)
	if err != nil {
		log.Error(err, log.Data{"message": "Unable to parse TableHTML to table element", "ParseRequest": request})
		return nil, err
	}

	requestJSON := &models.RenderRequest{
		Filename:   request.Filename,
		Title:      request.Title,
		Subtitle:   request.Subtitle,
		Source:     request.Source,
		URI:        request.URI,
		StyleClass: request.StyleClass,
		TableType:  "generated-table",
		Footnotes:  request.Footnotes}

	// todo parse table to requestJSON
	findNode(sourceTable, atom.Tbody)

	previewHTML, err := renderer.RenderHTML(requestJSON)
	if err != nil {
		log.Error(err, log.Data{"message": "Unable to render preview HTML", "ParseRequest": request, "RenderRequest": requestJSON})
		return nil, err
	}
	response := ResponseModel{JSON: *requestJSON, PreviewHTML: string(previewHTML)}

	return marshalResponse(response)
}

// parseTableToNode parses a string of html and returns the single table node, or an error if the html doesn't contain a single table
func parseTableToNode(tableHTML string) (*html.Node, error) {
	nodes, err := html.ParseFragment(strings.NewReader(tableHTML), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, errors.New("table_html could not be parsed into a single element")
	}
	if nodes[0].DataAtom != atom.Table {
		return nil, errors.New("table_html could not be parsed into a table element")
	}
	return nodes[0], nil
}

// marshalResponse marshals the ResponseModel to json, turning off escaping of html
func marshalResponse(response ResponseModel) ([]byte, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(response)
	if err == nil {
		err = writer.Flush()
	}
	return b.Bytes(), err
}

// find an attribute for the node - returns empty string if not found
func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// depth-first search for the first node of the given type
func findNode(n *html.Node, a atom.Atom) *html.Node {
	return findNodeWithAttributes(n, a, nil)
}

// depth-first search for the first node of the given type with the given attributes
func findNodeWithAttributes(n *html.Node, a atom.Atom, attr map[string]string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a && hasAttributes(c, attr) {
			return c
		}
		gc := findNodeWithAttributes(c, a, attr)
		if gc != nil {
			return gc
		}
	}
	return nil
}

// return true if the given node has all the attribute values
func hasAttributes(n *html.Node, attr map[string]string) bool {
	for key, value := range attr {
		if getAttribute(n, key) != value {
			return false
		}
	}
	return true
}

// returns all child nodes of the given type
func findNodes(n *html.Node, a atom.Atom) []*html.Node {
	return findNodesWithAttributes(n, a, nil)
}

// returns all child nodes of the given type with the given attributes
func findNodesWithAttributes(n *html.Node, a atom.Atom, attr map[string]string) []*html.Node {
	var result []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a && hasAttributes(c, attr) {
			result = append(result, c)
		}
		result = append(result, findNodesWithAttributes(c, a, attr)...)
	}
	return result
}
