package emailing

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// TemplateEngine manages email templates with embedded filesystem support
type TemplateEngine struct {
	templates *template.Template
	funcs     template.FuncMap
	mu        sync.RWMutex
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: template.New(""),
		funcs:     defaultFuncs(),
	}
}

// LoadTemplates loads templates from an embedded filesystem
func (te *TemplateEngine) LoadTemplates(fsys embed.FS, root string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Walk through all template files
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process HTML files
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Read the template file
		content, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Get the template name (filename without extension)
		name := strings.TrimSuffix(filepath.Base(path), ".html")

		// Parse the template with functions
		tmpl := te.templates.New(name).Funcs(te.funcs)
		_, err = tmpl.Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		return nil
	})
}

// LoadTemplatesFromDir loads templates from a directory (for development)
func (te *TemplateEngine) LoadTemplatesFromDir(dir string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Parse all templates in the directory
	tmpl, err := template.New("").Funcs(te.funcs).ParseGlob(filepath.Join(dir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to parse templates from %s: %w", dir, err)
	}

	te.templates = tmpl
	return nil
}

// Render renders a template with the given data
func (te *TemplateEngine) Render(templateName string, data interface{}) (string, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	tmpl := te.templates.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// AddFunc adds a custom template function
func (te *TemplateEngine) AddFunc(name string, fn interface{}) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.funcs[name] = fn
}

// defaultFuncs returns the default template functions
func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Utility functions
		"default": func(def, val any) any {
			if val == nil || val == "" {
				return def
			}
			return val
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}
}

// TemplateData represents common email template data
type TemplateData struct {
	// Common fields
	To       string
	From     string
	FromName string
	Subject  string

	// User fields
	UserName  string
	UserEmail string

	// Application fields
	AppName string
	AppURL  string
	AppLogo string

	// Action fields
	ActionURL    string
	ActionButton string

	// Content fields
	Title     string
	Preheader string
	Body      template.HTML
	Footer    template.HTML

	// Additional data
	Data map[string]interface{}
}

// NewTemplateData creates a new template data with defaults
func NewTemplateData() *TemplateData {
	return &TemplateData{
		Data: make(map[string]interface{}),
	}
}
