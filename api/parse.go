package api

import (
	"net/http"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/parser"
	"github.com/ONSdigital/log.go/v2/log"
)

// Content types
var (
	contentJSON = "application/json"
)

func (api *RendererAPI) parseHTML(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	parseRequest, err := models.CreateParseRequest(ctx, r.Body)
	if err != nil {
		log.Error(ctx, "error occurred when trying to create model parse request", err)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err = parseRequest.ValidateParseRequest(ctx); err != nil {
		log.Error(ctx, "error occurred when trying to validate model parse request", err)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	bytes, err := parser.ParseHTML(ctx, parseRequest)
	setContentType(w, contentJSON)
	if err != nil {
		log.Error(ctx, "error occurred when trying to parse HTML", err)
		setErrorCode(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(bytes); err != nil {
		log.Error(ctx, "error occurred when trying to parse HTML", err)
		setErrorCode(ctx, w, err)
		return
	}
	log.Info(ctx, "parsed an HTML table to JSON", log.Data{"response_bytes": len(bytes)})
}
