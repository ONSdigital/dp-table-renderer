package parser

import (
	"encoding/json"

	"github.com/ONSdigital/dp-table-renderer/models"
)

type responseModel struct {
	json        models.RenderRequest `json:"json"`
	previewHTML string               `json:"preview_html"`
}

// ParseHTML parses the html table in the request and generates correctly formatted json
func ParseHTML(request *models.ParseRequest) ([]byte, error) {
	response := responseModel{}

	return json.Marshal(response)
}
