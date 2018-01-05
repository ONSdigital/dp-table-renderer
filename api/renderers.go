package api

import (
	"net/http"
	"github.com/ONSdigital/dp-table-renderer/models"
	"github.com/ONSdigital/go-ns/log"
)

// Error types
var (
	internalError             = "Failed to process the request due to an internal error"
	badRequest                = "Bad request - Invalid request body"
	unauthorised              = "Unauthorised, request lacks valid authentication credentials"
	statusBadRequest          = "bad request"
	statusUnprocessableEntity = "unprocessable entity"
)

// Content types
var (
	contentHTML = "text/html"
)

func (api *RendererAPI) renderTable(w http.ResponseWriter, r *http.Request) {

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

	bytes := []byte("<div><table><thead><tr><th>" + renderRequest.Title + "</th></tr></thead></table></div>")
	setContentType(w, contentHTML)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, log.Data{"render_request": renderRequest})
		setErrorCode(w, err)
		return
	}

	log.Info("Rendered a table", log.Data{"render_request": renderRequest})
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
