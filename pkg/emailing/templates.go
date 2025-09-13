package emailing

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// parseEmailTemplate lazily parses a specific email template with all its partials
func parseEmailTemplate(emailFS embed.FS, rootDir string, templateName string, funcs template.FuncMap) (*template.Template, error) {
	// Merge all functions into a single map
	allFuncs := template.FuncMap{
		"now": time.Now,
		"formatTime": func(t time.Time, format string) string {
			return t.Format(format)
		},
		"baseURL": func(r any) string {
			// Helper to get full base URL from request
			if req, ok := r.(*http.Request); ok {
				scheme := "http"
				if req.TLS != nil {
					scheme = "https"
				}
				// Check for X-Forwarded-Proto header
				if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
					scheme = proto
				}
				return fmt.Sprintf("%s://%s", scheme, req.Host)
			}
			return ""
		},
	}
	
	// Add all provided functions
	for name, fn := range funcs {
		allFuncs[name] = fn
	}
	
	// Create a new template set with all functions
	tmpl := template.New(templateName).Funcs(allFuncs)
	
	// First, parse all partials (they're referenced by templates)
	partialsDir := filepath.Join(rootDir, "partials")
	err := fs.WalkDir(emailFS, partialsDir, func(path string, d fs.DirEntry, err error) error {
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

		// Read the partial file
		content, err := emailFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read partial %s: %w", path, err)
		}

		// Parse the partial (it will have {{define}} blocks)
		_, err = tmpl.Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse partial %s: %w", path, err)
		}

		return nil
	})
	
	if err != nil {
		// Partials might not exist, that's ok
		if !strings.Contains(err.Error(), "no such file or directory") {
			return nil, fmt.Errorf("failed to parse partials: %w", err)
		}
	}
	
	// Now parse the main template
	templatePath := filepath.Join(rootDir, templateName)
	content, err := emailFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templateName, err)
	}
	
	// Parse the main template - use the existing tmpl with functions already set
	_, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	return tmpl, nil
}

// renderTemplate renders an email template with the given context
func (c *Collection) renderTemplate(ctx *emailContext) (html string, text string, err error) {
	// Check that templates are configured
	if c.emailFS == (embed.FS{}) {
		return "", "", fmt.Errorf("templates not loaded - call LoadTemplates first")
	}

	// Add default Year function if not provided
	if ctx.funcs["Year"] == nil {
		ctx.funcs["Year"] = func() int { return time.Now().Year() }
	}
	
	// Parse the template lazily with all the functions registered
	tmpl, err := parseEmailTemplate(c.emailFS, c.emailFSDir, ctx.template, ctx.funcs)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render HTML version - pass nil as data since everything is functions
	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to render template: %w", err)
	}

	html = htmlBuf.String()

	// Generate text version from HTML
	text = htmlToText(html)

	return html, text, nil
}

// htmlToText converts HTML to plain text (simplified)
func htmlToText(html string) string {
	text := html

	// Remove style tags and their content
	for {
		start := strings.Index(text, "<style")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "</style>")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+8:]
	}

	// Remove script tags and their content
	for {
		start := strings.Index(text, "<script")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "</script>")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+9:]
	}

	// Replace common HTML entities
	replacements := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  `"`,
		"&#39;":   "'",
		"&ndash;": "-",
		"&mdash;": "--",
		"<br>":    "\n",
		"<br/>":   "\n",
		"<br />":  "\n",
		"</p>":    "\n\n",
		"</div>":  "\n",
		"</li>":   "\n",
		"</h1>":   "\n\n",
		"</h2>":   "\n\n",
		"</h3>":   "\n\n",
		"</h4>":   "\n",
		"</h5>":   "\n",
		"</h6>":   "\n",
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	// Remove remaining HTML tags
	for {
		start := strings.Index(text, "<")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}

	// Clean up whitespace
	lines := strings.Split(text, "\n")
	var cleanLines []string
	emptyCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			emptyCount++
			// Don't allow more than 2 consecutive empty lines
			if emptyCount <= 2 {
				cleanLines = append(cleanLines, "")
			}
		} else {
			emptyCount = 0
			cleanLines = append(cleanLines, line)
		}
	}

	// Remove leading and trailing empty lines
	for len(cleanLines) > 0 && cleanLines[0] == "" {
		cleanLines = cleanLines[1:]
	}
	for len(cleanLines) > 0 && cleanLines[len(cleanLines)-1] == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}

	return strings.Join(cleanLines, "\n")
}