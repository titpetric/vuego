package server

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego"
)

// MiddlewareOption configures the middleware behavior.
type MiddlewareOption func(*middlewareConfig)

type middlewareConfig struct {
	loadOptions []vuego.LoadOption
}

// WithLoadOption adds a LoadOption to the middleware's Vue instance.
func WithLoadOption(opt vuego.LoadOption) MiddlewareOption {
	return func(cfg *middlewareConfig) {
		cfg.loadOptions = append(cfg.loadOptions, opt)
	}
}

// Middleware creates an http.Handler that processes .vuego files from the given filesystem.
// It renders .vuego files with accompanying .yml or .json data files.
// Non-.vuego requests are passed through to the next handler or return 404.
func Middleware(contentFS fs.FS, opts ...MiddlewareOption) http.Handler {
	cfg := &middlewareConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &middlewareHandler{
		fs:          contentFS,
		loadOptions: cfg.loadOptions,
	}
}

type middlewareHandler struct {
	fs          fs.FS
	loadOptions []vuego.LoadOption
}

func (h *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(r.URL.Path, "/")

	// Try to serve as .vuego file
	vuegoPath := urlPath
	if !strings.HasSuffix(urlPath, ".vuego") {
		vuegoPath = urlPath + ".vuego"
	}

	// Check if .vuego file exists
	if _, err := fs.Stat(h.fs, vuegoPath); err == nil {
		h.serveVuego(w, r, vuegoPath)
		return
	}

	// Not a .vuego file, return 404
	http.NotFound(w, r)
}

func (h *middlewareHandler) serveVuego(w http.ResponseWriter, r *http.Request, filePath string) {
	// Load data from .yml or .json file
	data, err := LoadDataFile(h.fs, filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load data: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl := vuego.NewFS(h.fs, h.loadOptions...)
	var buf bytes.Buffer
	if err := tmpl.Load(filePath).Fill(data).Render(r.Context(), &buf); err != nil {
		http.Error(w, fmt.Sprintf("render error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

// LoadDataFile loads data from .yml, .yaml, or .json file accompanying a .vuego file.
func LoadDataFile(contentFS fs.FS, vuegoPath string) (map[string]any, error) {
	basePath := strings.TrimSuffix(vuegoPath, ".vuego")
	data := make(map[string]any)

	// Try .yml first, then .yaml, then .json
	for _, ext := range []string{".yml", ".yaml", ".json"} {
		dataPath := basePath + ext
		content, err := fs.ReadFile(contentFS, dataPath)
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", dataPath, err)
		}
		return data, nil
	}

	return data, nil
}

// MiddlewareDir creates an http.Handler that processes .vuego files from the given directory path.
// This is a convenience wrapper around Middleware that uses os.DirFS.
func MiddlewareDir(dir string, opts ...MiddlewareOption) http.Handler {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	return Middleware(os.DirFS(absDir), opts...)
}
