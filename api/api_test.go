package api

import (
	"testing"

	"github.com/gorilla/mux"
	"strings"
	"net/http"
	"net/http/httptest"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
)

var (
	host          = "http://localhost:80"
	renderHtmlUrl = host + "/render/html"
	requestBody   = `{"title":"table_title", "type":"table_type"}`
)

func TestSuccessfullyRenderTable(t *testing.T) {
	t.Parallel()
	Convey("Successfully render an html table", t, func() {
		reader := strings.NewReader(requestBody)
		r, err := http.NewRequest("POST", renderHtmlUrl, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "text/html")
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
		api := routes(host, mux.NewRouter())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldResemble, "Unknown render type\n")
	})

	Convey("When an invalid json message is sent, a bad request is returned", t, func() {
		reader := strings.NewReader("{")
		r, err := http.NewRequest("POST", renderHtmlUrl, reader)
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		api := routes(host, mux.NewRouter())
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)

		bodyBytes, _ := ioutil.ReadAll(w.Body)
		response := string(bodyBytes)
		So(response, ShouldResemble, "Bad request - Invalid request body\n")
	})
}
