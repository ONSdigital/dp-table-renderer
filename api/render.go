package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/gorilla/mux"
)

// Error types
var (
	internalError     = "Failed to process the request due to an internal error"
	badRequest        = "Bad request - Invalid request body"
	unknownRenderType = "Unknown render type"
	statusBadRequest  = "bad request"
)

// Content types
var (
	contentHTML = "text/html"
	contentXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	contentCSV  = "text/csv"
)

func (api *RendererAPI) renderTable(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	renderType := vars["render_type"]
	ctx := r.Context()

	renderRequest, err := models.CreateRenderRequest(ctx, r.Body)
	if err != nil {
		log.Error(ctx, "error with creating model render request", err)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err = renderRequest.ValidateRenderRequest(); err != nil {
		log.Error(ctx, "error with validating model render request", err)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	var bytes []byte

	switch renderType {
	case "html":
		bytes, err = renderer.RenderHTML(ctx, renderRequest)
		setContentType(w, contentHTML)
	case "xlsx":
		bytes, err = renderer.RenderXLSX(ctx, renderRequest)
		setContentType(w, contentXLSX)
	case "csv":
		bytes, err = renderer.RenderCSV(ctx, renderRequest)
		setContentType(w, contentCSV)
	default:
		log.Error(ctx, "Unknown render type", errors.New("Unknown render type"))
		http.Error(w, unknownRenderType, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(ctx, "Unknown render request", err)
		setErrorCode(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(ctx, "failed to write data to connection", err)
		setErrorCode(ctx, w, err)
		return
	}
	log.Info(ctx, "rendered a table", log.Data{"file_name": renderRequest.Filename, "response_bytes": len(bytes)})
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func setErrorCode(ctx context.Context, w http.ResponseWriter, err error) {
	log.Error(ctx, "error code:", err)
	switch err.Error() {
	case "Bad request":
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
