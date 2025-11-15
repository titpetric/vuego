package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing/fstest"

	"github.com/titpetric/vuego"
)

//go:embed static/*
var staticFS embed.FS

type RenderRequest struct {
	Template string `json:"template"`
	Data     string `json:"data"`
}

type RenderResponse struct {
	HTML  string `json:"html,omitempty"`
	Error string `json:"error,omitempty"`
}

// Example represents a single example with template and data.
type Example struct {
	Template string `json:"template"`
	Data     string `json:"data"`
}

// ExamplesResponse contains all available examples.
type ExamplesResponse struct {
	Examples map[string]Example `json:"examples"`
	Error    string             `json:"error,omitempty"`
}

// ExampleButton represents an example button with metadata.
type ExampleButton struct {
	Name        string
	Label       string
	Depth       int
	IsNested    bool
	ButtonClass string
	OnClick     string
}

// FirstExample holds the first example's template and data for initial render.
type FirstExample struct {
	Template string
	Data     string
}

// IndexPageData holds the data for rendering the index template.
type IndexPageData struct {
	Title        string
	Subtitle     string
	ExampleList  []any
	FirstExample FirstExample
}

// NewIndexPageData creates a new IndexPageData with default values.
func NewIndexPageData(examples map[string]Example) IndexPageData {
	exampleList := make([]any, 0, len(examples))
	for name := range examples {
		depth := strings.Count(name, "/")
		label := name
		isNested := depth > 0
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			label = name[idx+1:]
		}

		buttonClass := "example-root"
		if isNested {
			buttonClass = "example-nested"
		}

		exampleList = append(exampleList, ExampleButton{
			Name:        name,
			Label:       label,
			Depth:       depth,
			IsNested:    isNested,
			ButtonClass: buttonClass,
			OnClick:     fmt.Sprintf("loadExample('%s')", name),
		})
	}

	// Sort by depth first (root first), then alphabetically
	sort.Slice(exampleList, func(i, j int) bool {
		ei := exampleList[i].(ExampleButton)
		ej := exampleList[j].(ExampleButton)
		if ei.Depth != ej.Depth {
			return ei.Depth < ej.Depth
		}
		return ei.Name < ej.Name
	})

	// Get first example for initial render
	firstEx := FirstExample{Template: "", Data: "{}"}
	if len(exampleList) > 0 {
		ex := exampleList[0].(ExampleButton)
		if example, ok := examples[ex.Name]; ok {
			firstEx.Template = example.Template
			firstEx.Data = example.Data
		}
	}

	return IndexPageData{
		Title:        "VueGo Playground",
		Subtitle:     "Test your VueGo templates in real-time",
		ExampleList:  exampleList,
		FirstExample: firstEx,
	}
}

func main() {
	// Determine the examples filesystem
	var examplesFS fs.FS
	if len(os.Args) > 1 {
		// Use command-line argument as the source folder
		examplesFS = os.DirFS(os.Args[1])
		log.Printf("Loading examples from: %s", os.Args[1])
	} else {
		// Use embedded static/examples
		var err error
		examplesFS, err = fs.Sub(staticFS, "static/examples")
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Loading examples from: embedded static/examples")
	}

	// Serve static files
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Load examples upfront
	examplesMap := loadExamplesMap(examplesFS)
	log.Printf("Loaded %d examples", len(examplesMap))

	// Create a custom mux to handle routing
	mux := http.NewServeMux()

	// Create Vue renderer for the index template
	vue := vuego.NewVue(staticContent)

	// Handle root path to render index.vuego
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			pageData := NewIndexPageData(examplesMap) // Pass examplesMap
			if err := vue.Render(w, "index.vuego", pageData); err != nil {
				http.Error(w, fmt.Sprintf("could not render index.vuego: %v", err), http.StatusInternalServerError)
				return
			}
			return
		}
		// For other paths, serve from filesystem (CSS, JS, etc.)
		http.FileServer(http.FS(staticContent)).ServeHTTP(w, r)
	})

	// API endpoint for rendering
	mux.HandleFunc("/api/render", func(w http.ResponseWriter, r *http.Request) {
		handleRender(w, r, examplesFS)
	})

	// API endpoint for loading examples
	mux.HandleFunc("/api/examples", func(w http.ResponseWriter, r *http.Request) {
		handleExamples(w, r, examplesFS)
	})

	port := ":8080"
	log.Printf("VueGo Playground starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

// loadExamplesMap loads all examples from the filesystem into a map, including subdirectories.
func loadExamplesMap(examplesFS fs.FS) map[string]Example {
	examples := make(map[string]Example)

	// Group files by example name (without extension)
	exampleFiles := make(map[string]map[string]string)

	// Walk through all files in the filesystem
	err := fs.WalkDir(examplesFS, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		ext := filepath.Ext(name)

		// Only process .vuego and .json files
		if ext != ".vuego" && ext != ".json" {
			return nil
		}

		// Create a unique key from the file path (removing extension)
		// Use forward slashes in the key for consistency
		dir := path.Dir(filePath)
		baseName := strings.TrimSuffix(name, ext)
		var exampleKey string
		if dir == "." {
			exampleKey = baseName
		} else {
			exampleKey = path.Join(dir, baseName)
		}

		if exampleFiles[exampleKey] == nil {
			exampleFiles[exampleKey] = make(map[string]string)
		}
		exampleFiles[exampleKey][ext] = filePath

		return nil
	})

	if err != nil {
		log.Printf("could not walk examples directory: %v", err)
		return examples
	}

	// Load template and data pairs
	for exampleName, files := range exampleFiles {
		var example Example

		// Read template file (.vuego)
		if templateFile, ok := files[".vuego"]; ok {
			data, err := fs.ReadFile(examplesFS, templateFile)
			if err != nil {
				continue
			}
			example.Template = string(data)
		}

		// Read data file (.json) - default to {} if not found
		example.Data = "{}"
		if dataFile, ok := files[".json"]; ok {
			data, err := fs.ReadFile(examplesFS, dataFile)
			if err == nil {
				example.Data = string(data)
			}
		}

		// Only add if we have at least a template
		if example.Template != "" {
			examples[exampleName] = example
		}
	}

	return examples
}

func handleExamples(w http.ResponseWriter, r *http.Request, examplesFS fs.FS) {
	w.Header().Set("Content-Type", "application/json")

	examples := loadExamplesMap(examplesFS)

	json.NewEncoder(w).Encode(ExamplesResponse{
		Examples: examples,
	})
}

func handleRender(w http.ResponseWriter, r *http.Request, examplesFS fs.FS) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(RenderResponse{
			Error: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// Parse the data JSON string into map[string]any
	var data map[string]any
	if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
		json.NewEncoder(w).Encode(RenderResponse{
			Error: fmt.Sprintf("Invalid data JSON: %v", err),
		})
		return
	}

	// Create a filesystem that combines the examples filesystem with the inline template
	// This allows includes to be resolved from the examples directory
	templateFS := fstest.MapFS{
		"template.html": &fstest.MapFile{
			Data: []byte(req.Template),
		},
	}

	// Create a combined filesystem: first try the temp template, then fall back to examples
	combinedFS := &combinedFilesystem{
		primary:   templateFS,
		secondary: examplesFS,
	}

	// Render the template
	vue := vuego.NewVue(combinedFS)
	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "template.html", data)

	if err != nil {
		json.NewEncoder(w).Encode(RenderResponse{
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(RenderResponse{
		HTML: buf.String(),
	})
}

// combinedFilesystem implements fs.FS by combining a primary and secondary filesystem
type combinedFilesystem struct {
	primary   fs.FS
	secondary fs.FS
}

func (cfs *combinedFilesystem) Open(name string) (fs.File, error) {
	// Try primary first
	f, err := cfs.primary.Open(name)
	if err == nil {
		return f, nil
	}
	// Fall back to secondary
	return cfs.secondary.Open(name)
}

// Ensure combinedFilesystem implements fs.ReadDirFS
func (cfs *combinedFilesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(cfs.secondary, name)
}
