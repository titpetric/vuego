package tour

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/titpetric/platform"
	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/server"
)

//go:embed content
var embeddedContent embed.FS

//go:embed public
var embeddedPublic embed.FS

//go:embed templates
var embeddedTemplates embed.FS

// EmbeddedContentFS returns the embedded tour content filesystem.
func EmbeddedContentFS() fs.FS {
	sub, _ := fs.Sub(embeddedContent, "content")
	return sub
}

// Module represents the tour module for the platform.
type Module struct {
	platform.UnimplementedModule

	tour         *Tour
	tourFS       fs.FS
	indexTmpl    string
	readmeHTML   string
	doneHTML     string
}

// NewModule creates a new tour module using embedded content.
func NewModule() *Module {
	return &Module{}
}

// Name returns the module name.
func (m *Module) Name() string {
	return "vuego-tour"
}

// Mount registers the tour routes.
func (m *Module) Mount(_ context.Context, r platform.Router) error {
	tourFS := EmbeddedContentFS()

	tour, err := ParseTour(tourFS)
	if err != nil {
		return err
	}

	m.tour = tour
	m.tourFS = tourFS

	// Load index template
	tmplData, err := embeddedTemplates.ReadFile("templates/index.vuego")
	if err != nil {
		return err
	}
	m.indexTmpl = string(tmplData)

	// Load and render README
	readmeData, _ := embeddedContent.ReadFile("content/README.md")
	if len(readmeData) > 0 {
		m.readmeHTML = string(readmeData)
	}

	// Load DONE.md
	doneData, _ := embeddedContent.ReadFile("content/DONE.md")
	if len(doneData) > 0 {
		m.doneHTML = string(doneData)
	}

	r.Get("/", m.serveIndex)
	r.Get("/done", m.serveDone)
	r.Get("/lesson/{chapter}/{lesson}", m.serveLessonPage)
	r.Get("/static/tour.js", m.serveJS)
	r.Get("/static/tour.css", m.serveCSS)
	r.Post("/render", m.handleRender)

	return nil
}

func (m *Module) handleLessonPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse /lesson/{chapter}/{lesson} from path
	path := strings.TrimPrefix(r.URL.Path, "/lesson/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid lesson path"})
		return
	}

	lesson := m.tour.GetLessonByName(parts[0], parts[1])
	if lesson == nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "lesson not found"})
		return
	}

	_ = json.NewEncoder(w).Encode(lesson)
}

func (m *Module) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	m.renderTourPage(w, nil, m.readmeHTML, false)
}

func (m *Module) serveDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	m.renderTourPage(w, nil, m.doneHTML, true)
}

func (m *Module) serveLessonPage(w http.ResponseWriter, r *http.Request) {
	chapterName := chi.URLParam(r, "chapter")
	lessonIdx := chi.URLParam(r, "lesson")

	lesson := m.tour.GetLessonByName(chapterName, lessonIdx)

	// Check Accept header - return JSON for API requests
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		if lesson == nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "lesson not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(lesson)
		return
	}

	// Return HTML page
	if lesson == nil {
		lesson = m.tour.FirstLesson()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	m.renderTourPage(w, lesson, "", false)
}

func (m *Module) handleLesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chapterName := chi.URLParam(r, "chapter")
	lessonIdx := chi.URLParam(r, "lesson")

	lesson := m.tour.GetLessonByName(chapterName, lessonIdx)
	if lesson == nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "lesson not found"})
		return
	}

	_ = json.NewEncoder(w).Encode(lesson)
}

func (m *Module) handleRender(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		_ = json.NewEncoder(w).Encode(server.RenderResponse{Error: "method not allowed"})
		return
	}

	var req server.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = json.NewEncoder(w).Encode(server.RenderResponse{Error: "invalid JSON: " + err.Error()})
		return
	}

	html, err := server.Render(r.Context(), m.tourFS, req,
		vuego.WithLessProcessor(),
	)
	if err != nil {
		_ = json.NewEncoder(w).Encode(server.RenderResponse{Error: err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(server.RenderResponse{HTML: html})
}

func (m *Module) serveJS(w http.ResponseWriter, _ *http.Request) {
	data, _ := embeddedPublic.ReadFile("public/tour.js")
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = w.Write(data)
}

func (m *Module) serveCSS(w http.ResponseWriter, _ *http.Request) {
	data, _ := embeddedPublic.ReadFile("public/tour.css")
	w.Header().Set("Content-Type", "text/css")
	_, _ = w.Write(data)
}

func (m *Module) renderTourPage(w http.ResponseWriter, lesson *Lesson, markdown string, isDone bool) {
	data := map[string]any{
		"lesson":   lesson,
		"chapters": m.tour.Chapters,
		"total":    m.tour.LessonCount(),
		"isDone":   isDone,
	}
	if markdown != "" {
		data["readme"] = markdown
	}

	tmpl := vuego.New()
	var buf bytes.Buffer
	_ = tmpl.Fill(data).RenderString(context.Background(), &buf, m.indexTmpl)
	_, _ = buf.WriteTo(w)
}


