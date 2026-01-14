package main

import (
	"bytes"
	"context"
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

	yaml "gopkg.in/yaml.v2"

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

// ComponentGroup represents a grouped set of components by folder
type ComponentGroup struct {
	Label      string
	Components []ExampleButton
}

// ExampleButton represents an example button with metadata.
type ExampleButton struct {
	Name        string
	Label       string
	Component   string // Component name with folder prefix (e.g., "form-input")
	Folder      string // Folder path (e.g., "form")
	Depth       int
	IsPage      bool
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
	Title         string
	Subtitle      string
	PageButtons   []ExampleButton
	ComponentList []ComponentGroup
	FirstExample  FirstExample
}

// NewIndexPageData creates a new IndexPageData with default values.
func NewIndexPageData(examples map[string]Example) IndexPageData {
	var pageButtons []ExampleButton
	componentsByFolder := make(map[string][]ExampleButton)

	for name := range examples {
		parts := strings.Split(name, "/")

		// Check if this is a component (starts with "components")
		isComponent := len(parts) > 1 && parts[0] == "components"

		label := parts[len(parts)-1]

		if !isComponent {
			// Root-level page
			pageButtons = append(pageButtons, ExampleButton{
				Name:        name,
				Label:       label,
				Component:   label,
				Folder:      "",
				Depth:       0,
				IsPage:      true,
				ButtonClass: "example-root",
				OnClick:     fmt.Sprintf("loadExample('%s')", name),
			})
		} else {
			// Component in components folder
			// Determine folder group
			var folderGroup string
			if len(parts) == 2 {
				// Root component: components/list.vuego
				folderGroup = "Component"
			} else {
				// Nested component: components/form/input.vuego
				folderGroup = parts[1]
			}

			componentName := buildComponentName(parts)

			componentsByFolder[folderGroup] = append(componentsByFolder[folderGroup], ExampleButton{
				Name:        name,
				Label:       label,
				Component:   componentName,
				Folder:      folderGroup,
				Depth:       len(parts) - 1,
				IsPage:      false,
				ButtonClass: "example-nested",
				OnClick:     fmt.Sprintf("loadExample('%s')", name),
			})
		}
	}

	// Sort pages alphabetically
	sort.Slice(pageButtons, func(i, j int) bool {
		return pageButtons[i].Name < pageButtons[j].Name
	})

	// Sort component groups by folder name (with "Component" first)
	var folderNames []string
	for folder := range componentsByFolder {
		folderNames = append(folderNames, folder)
	}
	sort.Slice(folderNames, func(i, j int) bool {
		// Put "Component" (root components) first
		if folderNames[i] == "Component" {
			return true
		}
		if folderNames[j] == "Component" {
			return false
		}
		return folderNames[i] < folderNames[j]
	})

	var componentList []ComponentGroup
	for _, folder := range folderNames {
		components := componentsByFolder[folder]
		// Sort components alphabetically
		sort.Slice(components, func(i, j int) bool {
			return components[i].Name < components[j].Name
		})
		componentList = append(componentList, ComponentGroup{
			Label:      folder,
			Components: components,
		})
	}

	// Get first example for initial render
	firstEx := FirstExample{Template: "", Data: "{}"}
	if len(pageButtons) > 0 {
		ex := pageButtons[0]
		if example, ok := examples[ex.Name]; ok {
			firstEx.Template = example.Template
			firstEx.Data = example.Data
		}
	} else if len(componentList) > 0 && len(componentList[0].Components) > 0 {
		ex := componentList[0].Components[0]
		if example, ok := examples[ex.Name]; ok {
			firstEx.Template = example.Template
			firstEx.Data = example.Data
		}
	}

	return IndexPageData{
		Title:         "Vuego Playground",
		Subtitle:      "Test your Vuego templates in real-time",
		PageButtons:   pageButtons,
		ComponentList: componentList,
		FirstExample:  firstEx,
	}
}

// buildComponentName builds a kebab-case component name from path parts
// e.g., ["components", "form", "input"] -> "form-input"
func buildComponentName(parts []string) string {
	// Skip first part if it's "components"
	startIdx := 0
	if len(parts) > 0 && parts[0] == "components" {
		startIdx = 1
	}

	if startIdx >= len(parts) {
		return ""
	}

	// Join remaining parts with hyphens
	return strings.Join(parts[startIdx:], "-")
}

func main() {
	// Determine the examples filesystem
	var examplesFS fs.FS
	var examplesDir string
	isEmbedded := true

	if len(os.Args) > 1 {
		// Use command-line argument as the source folder
		examplesDir = os.Args[1]
		examplesFS = os.DirFS(examplesDir)
		isEmbedded = false
		log.Printf("Loading examples from: %s", examplesDir)
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
	tmpl := vuego.NewFS(staticContent, vuego.WithLessProcessor())

	// Handle root path to render index.vuego
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			pageData := NewIndexPageData(examplesMap) // Pass examplesMap
			if err := tmpl.Load("index.vuego").Fill(pageData).Render(r.Context(), w); err != nil {
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

	// API endpoint for server status
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{
			"isEmbedFS": isEmbedded,
		})
	})

	// API endpoint for cheatsheet
	mux.HandleFunc("/api/cheatsheet", func(w http.ResponseWriter, r *http.Request) {
		handleCheatsheet(w, r)
	})

	// API endpoint for saving files
	mux.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		handleSave(w, r, examplesDir, isEmbedded)
	})

	// API endpoint for creating files
	mux.HandleFunc("/api/create", func(w http.ResponseWriter, r *http.Request) {
		handleCreate(w, r, examplesDir, isEmbedded)
	})

	port := ":8080"
	log.Printf("Vuego Playground starting on http://localhost%s", port)
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
		if ext != ".vuego" && ext != ".json" && ext != ".yaml" && ext != ".yml" {
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
		} else {
			continue
		}

		// Read data file (.json) - default to {} if not found
		example.Data = "{}"
		if dataFile, ok := files[".json"]; ok {
			data, err := fs.ReadFile(examplesFS, dataFile)
			if err == nil {
				example.Data = string(data)
			}
		}

		yamlfile, ok := files[".yaml"]
		if !ok {
			yamlfile, ok = files[".yml"]
		}
		if ok {
			data, err := fs.ReadFile(examplesFS, yamlfile)
			if err == nil {
				var v any
				if err := yaml.Unmarshal(data, &v); err != nil {
					log.Printf("error decoding %s: %v (ignored)\n", yamlfile, err)
				} else {
					v = convert(v)
					b, err := json.MarshalIndent(v, "", "  ")
					example.Data = string(b)
					if err != nil {
						log.Printf("Reading %s, got %s, err %s\n", yamlfile, example.Data, err)
					}
				}
			}
		}

		// Only add if we have at least a template
		if example.Template != "" {
			examples[exampleName] = example
		}
	}

	return examples
}

func convert(i any) any {
	switch x := i.(type) {
	case map[any]any:
		m2 := map[string]any{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []any:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func handleExamples(w http.ResponseWriter, r *http.Request, examplesFS fs.FS) {
	w.Header().Set("Content-Type", "application/json")

	examples := loadExamplesMap(examplesFS)

	_ = json.NewEncoder(w).Encode(ExamplesResponse{
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
		_ = json.NewEncoder(w).Encode(RenderResponse{
			Error: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// Parse the data JSON string into map[string]any
	var data map[string]any
	if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
		_ = json.NewEncoder(w).Encode(RenderResponse{
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

	// Render the template with Less processor support
	tmpl := vuego.NewFS(combinedFS, vuego.WithLessProcessor())
	var buf bytes.Buffer
	if err := tmpl.Load("template.html").Fill(data).Render(context.Background(), &buf); err != nil {
		_ = json.NewEncoder(w).Encode(RenderResponse{
			Error: err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(RenderResponse{
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

// handleCheatsheet renders the footer.vuego component and returns HTML
func handleCheatsheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	tmpl := vuego.NewFS(staticFS, vuego.WithLessProcessor())
	err := tmpl.Load("static/footer.vuego").Fill(map[string]any{}).Render(r.Context(), &buf)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"content": buf.String(),
	})
}

// SaveRequest contains template and data to be saved
type SaveRequest struct {
	Template string `json:"template"`
	Data     string `json:"data"`
	Name     string `json:"name"` // Example name/path (without extension)
}

// SaveResponse contains the save operation result
type SaveResponse struct {
	Error string `json:"error,omitempty"`
	Path  string `json:"path,omitempty"`
}

// handleSave saves the current template and data to files
func handleSave(w http.ResponseWriter, r *http.Request, examplesDir string, isEmbedded bool) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		_ = json.NewEncoder(w).Encode(SaveResponse{
			Error: "Method not allowed",
		})
		return
	}

	if isEmbedded {
		_ = json.NewEncoder(w).Encode(SaveResponse{
			Error: "cannot save to embedded filesystem",
		})
		return
	}

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = json.NewEncoder(w).Encode(SaveResponse{
			Error: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// If no name provided, use default
	exampleName := req.Name
	if exampleName == "" {
		exampleName = "template"
	}

	// Build file paths based on example name
	templatePath := filepath.Join(examplesDir, exampleName+".vuego")
	dataPath := filepath.Join(examplesDir, exampleName+".json")

	// Create directories if they don't exist (for nested paths)
	if dir := filepath.Dir(templatePath); dir != examplesDir {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			_ = json.NewEncoder(w).Encode(SaveResponse{
				Error: fmt.Sprintf("Failed to create directory: %v", err),
			})
			return
		}
	}

	// Save template file
	if err := os.WriteFile(templatePath, []byte(req.Template), 0o644); err != nil {
		_ = json.NewEncoder(w).Encode(SaveResponse{
			Error: fmt.Sprintf("Failed to save template: %v", err),
		})
		return
	}

	// Save data file
	if err := os.WriteFile(dataPath, []byte(req.Data), 0o644); err != nil {
		_ = json.NewEncoder(w).Encode(SaveResponse{
			Error: fmt.Sprintf("Failed to save data: %v", err),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(SaveResponse{
		Path: fmt.Sprintf("%s.vuego and %s.json", exampleName, exampleName),
	})
}

// CreateRequest contains file creation parameters
type CreateRequest struct {
	Name string `json:"name"`
	Type string `json:"type"` // "page" or "component"
}

// CreateResponse contains the creation result
type CreateResponse struct {
	Error string `json:"error,omitempty"`
	Path  string `json:"path,omitempty"`
}

// handleCreate creates a new template or component file
func handleCreate(w http.ResponseWriter, r *http.Request, examplesDir string, isEmbedded bool) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: "Method not allowed",
		})
		return
	}

	if isEmbedded {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: "cannot create files in embedded filesystem",
		})
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// Sanitize filename
	filename := filepath.Base(req.Name)
	if filename == "" || filename == "." || filename == ".." {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: "invalid filename",
		})
		return
	}

	// Add .vuego extension if not already present
	if !strings.HasSuffix(filename, ".vuego") {
		filename = filename + ".vuego"
	}

	var filePath string
	if req.Type == "component" {
		// Create components directory if it doesn't exist
		compDir := filepath.Join(examplesDir, "components")
		if err := os.MkdirAll(compDir, 0o755); err != nil {
			_ = json.NewEncoder(w).Encode(CreateResponse{
				Error: fmt.Sprintf("Failed to create components directory: %v", err),
			})
			return
		}
		filePath = filepath.Join(compDir, filename)
	} else {
		// Page type (default)
		filePath = filepath.Join(examplesDir, filename)
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: "file already exists",
		})
		return
	}

	// Create the file with a basic template
	templateContent := `<div class="` + filename + `">
  <!-- Add your template here -->
</div>
`
	if err := os.WriteFile(filePath, []byte(templateContent), 0o644); err != nil {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: fmt.Sprintf("Failed to create file: %v", err),
		})
		return
	}

	// Also create a corresponding .json data file
	jsonPath := filepath.Join(filepath.Dir(filePath), strings.TrimSuffix(filepath.Base(filePath), ".vuego")+".json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		_ = json.NewEncoder(w).Encode(CreateResponse{
			Error: fmt.Sprintf("Failed to create data file: %v", err),
		})
		return
	}

	// Calculate the example name (relative path without extension)
	// This is used by the frontend to load the created file directly
	var exampleName string
	if req.Type == "component" {
		exampleName = filepath.Join("components", strings.TrimSuffix(filename, ".vuego"))
	} else {
		exampleName = strings.TrimSuffix(filename, ".vuego")
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"path":  filePath,
		"name":  exampleName,
		"error": "",
	})
}
