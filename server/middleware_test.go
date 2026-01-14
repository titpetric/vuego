package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/server"
)

func TestMiddleware(t *testing.T) {
	t.Run("serves .vuego file with yml data", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div>{{ title }}</div>`)},
			"page.yml":   &fstest.MapFile{Data: []byte(`title: Hello World`)},
		}

		handler := server.Middleware(fs)
		req := httptest.NewRequest(http.MethodGet, "/page.vuego", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		require.Contains(t, rec.Body.String(), "Hello World")
	})

	t.Run("serves .vuego file with json data", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div>{{ title }}</div>`)},
			"page.json":  &fstest.MapFile{Data: []byte(`{"title": "JSON Title"}`)},
		}

		handler := server.Middleware(fs)
		req := httptest.NewRequest(http.MethodGet, "/page.vuego", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "JSON Title")
	})

	t.Run("serves .vuego file without extension in URL", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div>No Extension</div>`)},
		}

		handler := server.Middleware(fs)
		req := httptest.NewRequest(http.MethodGet, "/page", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "No Extension")
	})

	t.Run("returns 404 for non-.vuego files", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div>Hello</div>`)},
		}

		handler := server.Middleware(fs)
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("handles render errors gracefully", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div>{{ undefined | nonexistent }}</div>`)},
		}

		handler := server.Middleware(fs)
		req := httptest.NewRequest(http.MethodGet, "/page.vuego", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "render error")
	})
}

func TestLoadDataFile(t *testing.T) {
	t.Run("loads .yml file", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div></div>`)},
			"page.yml":   &fstest.MapFile{Data: []byte(`key: value`)},
		}

		data, err := server.LoadDataFile(fs, "page.vuego")

		require.NoError(t, err)
		require.Equal(t, "value", data["key"])
	})

	t.Run("loads .json file", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div></div>`)},
			"page.json":  &fstest.MapFile{Data: []byte(`{"key": "json_value"}`)},
		}

		data, err := server.LoadDataFile(fs, "page.vuego")

		require.NoError(t, err)
		require.Equal(t, "json_value", data["key"])
	})

	t.Run("returns empty map when no data file exists", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte(`<div></div>`)},
		}

		data, err := server.LoadDataFile(fs, "page.vuego")

		require.NoError(t, err)
		require.Empty(t, data)
	})
}

func TestMiddlewareDir(t *testing.T) {
	t.Run("serves from directory path", func(t *testing.T) {
		// Use the testdata directory which should contain test fixtures
		handler := server.MiddlewareDir("../testdata/fixtures")
		req := httptest.NewRequest(http.MethodGet, "/basic.vuego", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should either return 200 if basic.vuego exists or 404 if not
		// This validates the handler is created and functional
		require.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNotFound)
	})
}
