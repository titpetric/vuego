package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
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

func main() {
	// Serve static files
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticContent)))

	// API endpoint for rendering
	http.HandleFunc("/api/render", handleRender)

	port := ":8080"
	log.Printf("VueGo Playground starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRender(w http.ResponseWriter, r *http.Request) {
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

	// Create a virtual filesystem with the template
	templateFS := fstest.MapFS{
		"template.html": &fstest.MapFile{
			Data: []byte(req.Template),
		},
	}

	// Render the template
	vue := vuego.NewVue(templateFS)
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
