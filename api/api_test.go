package api

import (
	"testing"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	host           = "http://localhost:80"
	requestHTMLURL = host + "/render/html"
	requestXLSXURL = host + "/render/xlsx"
	requestCSVURL  = host + "/render/csv"
	requestBody    = `{"title":"table_title", "filename": "file_name", "type":"table_type"}`
	parseURL       = host + "/parse/html"
	parseBody      = `{"title":"table_title", "filename": "file_name", "table_html":"<table></table>"}`
)

var hcMock = healthcheck.HealthCheck{}

func TestSuccessfullyRenderTable(t *testing.T) {
	t.Parallel()
	Convey("Successfully render an html table", t, func() {
		reader := strings.NewReader(requestBody)
		r, err := http.NewRequest("POST", requestHTMLURL, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "text/html")
		So(w.Body.String(), ShouldContainSubstring, "<table")
		So(w.Body.String(), ShouldContainSubstring, "table_title")
	})

}

func TestSuccessfullyRenderSpreadsheet(t *testing.T) {
	t.Parallel()
	Convey("Successfully render an xlsx spreadsheet", t, func() {
		reader := strings.NewReader(requestBody)
		r, err := http.NewRequest("POST", requestXLSXURL, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		So(len(w.Body.String()), ShouldBeGreaterThan, 0)
	})

}

func TestSuccessfullyRenderCSV(t *testing.T) {
	t.Parallel()
	Convey("Successfully render a csv file", t, func() {
		reader := strings.NewReader(requestBody)
		r, err := http.NewRequest("POST", requestCSVURL, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "text/csv")
		So(len(w.Body.String()), ShouldBeGreaterThan, 0)
	})

}

func TestSuccessfullyParseTable(t *testing.T) {
	t.Parallel()
	Convey("Successfully parse an html table", t, func() {
		reader := strings.NewReader(parseBody)
		r, err := http.NewRequest("POST", parseURL, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		So(w.Body.String(), ShouldContainSubstring, "<table")
		So(w.Body.String(), ShouldContainSubstring, "table_title")
	})

}

func TestRejectInvalidRequest(t *testing.T) {
	t.Parallel()
	Convey("Reject invalid render type in url with StatusNotFound", t, func() {
		reader := strings.NewReader(requestBody)
		r, err := http.NewRequest("POST", host+"/render/foo", reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldResemble, "Unknown render type\n")
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", requestHTMLURL, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(mux.NewRouter(), &hcMock)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})
}
