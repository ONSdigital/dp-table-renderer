package jsontohtml

import (
	"encoding/json"

	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-table-renderer/models"
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

	previewHTML, err := RenderHTML(requestJSON)
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
