package api

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/dp-table-renderer/renderer"
	"github.com/ONSdigital/go-ns/log"
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

	renderRequest, err := models.CreateRenderRequest(r.Body)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err = renderRequest.ValidateRenderRequest(); err != nil {
		log.Error(err, nil)
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	var bytes []byte

	switch renderType {
	case "html":
		bytes, err = renderer.RenderHTML(renderRequest)
		setContentType(w, contentHTML)
	case "xlsx":
		bytes, err = renderer.RenderXLSX(renderRequest)
		setContentType(w, contentXLSX)
	case "csv":
		bytes, err = renderer.RenderCSV(renderRequest)
		setContentType(w, contentCSV)
	default:
		log.Error(errors.New("Unknown render type"), log.Data{"render_type": renderType})
		http.Error(w, unknownRenderType, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err, log.Data{"render_request": renderRequest})
		setErrorCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"render_request": renderRequest})
		setErrorCode(w, err)
		return
	}

	log.InfoC(renderRequest.Filename, "Rendered a table", log.Data{"render_type": renderType, "response_bytes": len(bytes)})
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func setErrorCode(w http.ResponseWriter, err error, typ ...string) {
	log.Debug("error is", log.Data{"error": err})
	switch err.Error() {
	case "Bad request":
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
