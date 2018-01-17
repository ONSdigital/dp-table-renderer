package api

import (
	"net/http"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/parser"
	"github.com/ONSdigital/go-ns/log"
)

// Content types
var (
	contentJSON = "application/json"
)

func (api *RendererAPI) parseHTML(w http.ResponseWriter, r *http.Request) {

	parseRequest, err := models.CreateParseRequest(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err = parseRequest.ValidateParseRequest(); err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	bytes, err := parser.ParseHTML(parseRequest)
	setContentType(w, contentJSON)
	if err != nil {
		log.Error(err, log.Data{"parse_request": parseRequest})
		setErrorCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(bytes); err != nil {
		log.Error(err, log.Data{"parse_request": parseRequest})
		setErrorCode(w, err)
		return
	}

	log.InfoC(parseRequest.Filename, "Parsed an html table to json", log.Data{"response_bytes": len(bytes)})
}
