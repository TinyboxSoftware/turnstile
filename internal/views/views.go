package views

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.html
var templateFS embed.FS

// Renderer holds the parsed template set for all views.
type Renderer struct {
	tmpl *template.Template
}

// NewRenderer parses all embedded HTML templates and returns a Renderer.
func NewRenderer() (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl}, nil
}

// CatchAllData is the template data for the catch-all route page.
type CatchAllData struct {
	StaticRoot string // e.g. "/_turnstile/static"
	AuthPrefix string // e.g. "/_turnstile"
	LoginURL   string
	LogoutURL  string
	HealthURL  string
}

// UnauthData is the template data for the access-denied page.
type UnauthData struct {
	StaticRoot   string
	Message      string
	LoginURL     string
	ReconsentURL string
}

// RenderCatchAll writes the catch-all error page to w with HTTP 200.
func (r *Renderer) RenderCatchAll(w http.ResponseWriter, data CatchAllData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = r.tmpl.ExecuteTemplate(w, "catchall.html", data)
}

// RenderUnauthenticated writes the access-denied page to w with the given status code.
func (r *Renderer) RenderUnauthenticated(w http.ResponseWriter, data UnauthData, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = r.tmpl.ExecuteTemplate(w, "unauthenticated.html", data)
}
