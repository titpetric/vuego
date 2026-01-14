package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/cmd/vuego/server"
)

func TestRender_JSONData(t *testing.T) {
	req := server.RenderRequest{
		Template: `<div>{{ name }}</div>`,
		Data:     `{"name": "World"}`,
	}

	html, err := server.Render(context.Background(), nil, req)
	require.NoError(t, err)
	require.Equal(t, "<div>World</div>\n", html)
}

func TestRender_YAMLData(t *testing.T) {
	req := server.RenderRequest{
		Template: `<div>{{ name }}</div>`,
		Data:     "name: World",
	}

	html, err := server.Render(context.Background(), nil, req)
	require.NoError(t, err)
	require.Equal(t, "<div>World</div>\n", html)
}

func TestRender_AdditionalFiles(t *testing.T) {
	req := server.RenderRequest{
		Template: `<vuego include="partial.vuego"></vuego>`,
		Files: map[string]string{
			"partial.vuego": `<span>Included</span>`,
		},
	}

	html, err := server.Render(context.Background(), nil, req)
	require.NoError(t, err)
	require.Equal(t, "<span>Included</span>\n", html)
}

func TestRender_InvalidData(t *testing.T) {
	req := server.RenderRequest{
		Template: `<div>{{ name }}</div>`,
		Data:     `{invalid json`,
	}

	_, err := server.Render(context.Background(), nil, req)
	require.Error(t, err)
}

func TestRenderHandler_POST(t *testing.T) {
	handler := server.RenderHandler(nil)

	body, _ := json.Marshal(server.RenderRequest{
		Template: `<p>{{ message }}</p>`,
		Data:     `{"message": "Hello"}`,
	})

	req := httptest.NewRequest(http.MethodPost, "/render", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp server.RenderResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Empty(t, resp.Error)
	require.Equal(t, "<p>Hello</p>\n", resp.HTML)
}

func TestRenderHandler_MethodNotAllowed(t *testing.T) {
	handler := server.RenderHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp server.RenderResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "method not allowed", resp.Error)
}

func TestRenderHandler_InvalidJSON(t *testing.T) {
	handler := server.RenderHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/render", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp server.RenderResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Contains(t, resp.Error, "invalid JSON:")
}

func TestRender_LayoutWithRelativePath(t *testing.T) {
	req := server.RenderRequest{
		Template: `---
layout: base.vuego
---
<h1>{{ title }}</h1>
<p>Page content</p>`,
		Data: `title: Hello`,
		Files: map[string]string{
			"base.vuego": `<!DOCTYPE html>
<html>
<body>
<main v-html="content"></main>
</body>
</html>`,
		},
	}

	html, err := server.Render(context.Background(), nil, req)
	require.NoError(t, err)
	require.Contains(t, html, "<h1>Hello</h1>")
	require.Contains(t, html, "<p>Page content</p>")
	require.Contains(t, html, "<main>")
	require.Contains(t, html, "</html>")
}
