package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHandler(t *testing.T) (http.Handler, func()) {
	t.Helper()

	server, client := testRedis(t)
	tmpl := template.Must(template.ParseGlob("templates/*"))
	return newHandler(client, tmpl), server.Close
}

func TestGetDrops(t *testing.T) {
	handler, _ := testHandler(t)

	// Seed through the handler's Redis by posting first.
	form := url.Values{"text": {"<script>alert(1)</script>"}}
	post := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(httptest.NewRecorder(), post)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	body := recorder.Body.String()
	assert.Contains(t, body, "&lt;script&gt;alert(1)&lt;/script&gt;")
	assert.NotContains(t, body, "<script>alert(1)</script>")
}

func TestPostDrop(t *testing.T) {
	handler, _ := testHandler(t)
	form := url.Values{"text": {"hello"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/", recorder.Header().Get("Location"))

	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Contains(t, getRecorder.Body.String(), "hello")
}

func TestPostRejectsEmptyDrop(t *testing.T) {
	handler, _ := testHandler(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("text="))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRoutesRejectUnsupportedMethodAndPath(t *testing.T) {
	handler, _ := testHandler(t)
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodPut, path: "/", want: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/missing", want: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, tt.want, recorder.Code)
		})
	}
}

func TestHandlerReportsRedisFailure(t *testing.T) {
	handler, closeRedis := testHandler(t)
	closeRedis()

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRequest(http.MethodPost, "/", strings.NewReader("text=hello")),
	} {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code, request.Method)
		assert.Contains(t, recorder.Body.String(), http.StatusText(http.StatusInternalServerError), request.Method)
	}
}

func TestHandlerReportsTemplateFailure(t *testing.T) {
	_, client := testRedis(t)
	handler := newHandler(client, template.Must(template.New("other").Parse("unused")))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
