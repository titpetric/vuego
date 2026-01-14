package tour_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/cmd/vuego/server/tour"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Test Chapter

## First Lesson

Lesson content here.

@file: first.vuego

---

## Second Lesson

More content.

@file: second.vuego
`),
		},
		"basics/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>Hello {{ name }}</div>`),
		},
		"basics/first.json": &fstest.MapFile{
			Data: []byte(`{"name": "World"}`),
		},
		"basics/second.vuego": &fstest.MapFile{
			Data: []byte(`<div>Goodbye</div>`),
		},
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	module := tour.NewModule()
	router := chi.NewRouter()
	err := module.Mount(context.Background(), router)
	require.NoError(t, err)
	return router
}

func TestModule_ServesIndexPage(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "Vuego tour")
}

func TestModule_LessonEndpoint_ReturnsJSON(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/lesson/interpolation/0", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var lesson tour.Lesson
	err := json.NewDecoder(rec.Body).Decode(&lesson)
	require.NoError(t, err)
	require.NotEmpty(t, lesson.Title)
}

func TestModule_LessonEndpoint_ReturnsHTML(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/lesson/interpolation/0", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "Vuego tour")
}

func TestModule_RenderEndpoint(t *testing.T) {
	handler := newTestHandler(t)

	body := `{"template": "<div>Hello {{ name }}</div>", "data": "{\"name\": \"World\"}"}`
	req := httptest.NewRequest(http.MethodPost, "/render", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Contains(t, resp["html"], "Hello World")
}

func TestModule_RenderEndpoint_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestModule_RenderEndpoint_InvalidJSON(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/render", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Contains(t, resp["error"], "invalid JSON")
}

func TestModule_DonePage(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/done", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "Vuego tour")
}

func TestModule_StaticJS(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/static/tour.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/javascript", rec.Header().Get("Content-Type"))
}

func TestModule_StaticCSS(t *testing.T) {
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/static/tour.css", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/css", rec.Header().Get("Content-Type"))
}
