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

// NotFoundPageData is the template data for the catch-all route page.
type NotFoundPageData struct {
	StaticRoot string // e.g. "/_turnstile/static"
	AuthPrefix string // e.g. "/_turnstile"
	LoginURL   string
	LogoutURL  string
	HealthURL  string
}

// ForbiddenPageData is the template data for the access-denied page.
type ForbiddenPageData struct {
	StaticRoot   string
	LoginURL     string
	ReconsentURL string
}

// ForbiddenPageData is the template data for the access-denied page.
type InternalServerErrorData struct {
	StaticRoot   string
	Message      *string
	LoginURL     string
	ReconsentURL string
}

// generic intenral function for rendering an HTML template
func (r *Renderer) renderHTMLTemplate(w http.ResponseWriter, name string, status int, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := r.tmpl.ExecuteTemplate(w, name, data); err != nil {
		// THROW HERE?
	}
}

// RenderNotFoundPage displays the 404 Not Found error page with metadata about turnstile
func (r *Renderer) RenderNotFoundPage(w http.ResponseWriter, data NotFoundPageData) {
	r.renderHTMLTemplate(w, "404.html", http.StatusForbidden, data)
}

// RenderForbiddenPage returns the 403 Forbidden error page with the specific error information from Railway
func (r *Renderer) RenderForbiddenPage(w http.ResponseWriter, data ForbiddenPageData) {
	r.renderHTMLTemplate(w, "403.html", http.StatusForbidden, data)
}

// RenderInternalServerErrorPage returns the 500 Internal Server Error page with a generic message
func (r *Renderer) RenderInternalServerErrorPage(w http.ResponseWriter, data InternalServerErrorData) {
	r.renderHTMLTemplate(w, "500.html", http.StatusInternalServerError, data)
}
