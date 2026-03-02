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
	tmpl       *template.Template
	staticRoot string // e.g. "/_turnstile/static"
}

// NewRenderer parses all embedded HTML templates and returns a Renderer.
func NewRenderer(staticRoot string) (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl, staticRoot: staticRoot}, nil
}

// NotFoundPageData is the template data for the catch-all route page.
type NotFoundPageData struct {
	StaticRoot string // e.g. "/_turnstile/static"
	AuthPrefix string // e.g. "/_turnstile"
	LoginURL   string
	LogoutURL  string
	HealthURL  string
}

// ErrorPageButton is a single action button rendered on an error page.
type ErrorPageButton struct {
	Label string
	URL   string
}

// ErrorPageData is the template data for the unified error page (error.html).
// It drives the 400, 403, and 500 error pages.
type ErrorPageData struct {
	StaticRoot string
	Title      string // e.g. "Bad Request: 400"
	Subtitle   string // short description under the title
	Message    string // if non-empty, shown in the danger alert box
	Note       string // if non-empty, shown as an extra paragraph below the alert
	Buttons    []ErrorPageButton
}

// generic internal function for rendering an HTML template
func (r *Renderer) renderHTMLTemplate(w http.ResponseWriter, name string, status int, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := r.tmpl.ExecuteTemplate(w, name, data); err != nil {
		// THROW HERE?
	}
}

// RenderErrorPage renders error.html with the provided status code and data.
func (r *Renderer) RenderErrorPage(w http.ResponseWriter, status int, data ErrorPageData) {
	data.StaticRoot = r.staticRoot
	r.renderHTMLTemplate(w, "error.html", status, data)
}

// RenderNotFoundPage displays the 404 Not Found error page with metadata about turnstile
func (r *Renderer) RenderNotFoundPage(w http.ResponseWriter, data NotFoundPageData) {
	data.StaticRoot = r.staticRoot
	r.renderHTMLTemplate(w, "404.html", http.StatusNotFound, data)
}
