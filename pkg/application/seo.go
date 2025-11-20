package application

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// SEOMetadata contains all SEO-related metadata for a page.
// It supports Open Graph, Twitter Cards, and standard meta tags.
type SEOMetadata struct {
	// Basic SEO
	Title       string // Page title (displayed in browser tab and search results)
	Description string // Meta description (160 characters recommended)
	Keywords    []string // Meta keywords (optional, less important for modern SEO)
	Canonical   string   // Canonical URL to prevent duplicate content issues

	// Open Graph (Facebook, LinkedIn, etc.)
	OGTitle       string // Open Graph title (defaults to Title)
	OGDescription string // Open Graph description (defaults to Description)
	OGImage       string // Open Graph image URL (absolute URL required)
	OGImageAlt    string // Alt text for OG image
	OGType        string // Open Graph type (article, website, profile, etc.)
	OGSiteName    string // Site name for Open Graph
	OGURL         string // Open Graph URL (defaults to Canonical)

	// Twitter Card
	TwitterCard        string // Card type: summary, summary_large_image, app, player
	TwitterSite        string // @username of website
	TwitterCreator     string // @username of content creator
	TwitterTitle       string // Twitter title (defaults to OGTitle or Title)
	TwitterDescription string // Twitter description (defaults to OGDescription or Description)
	TwitterImage       string // Twitter image (defaults to OGImage)
	TwitterImageAlt    string // Alt text for Twitter image (defaults to OGImageAlt)

	// Additional metadata
	Author      string   // Article author
	PublishedAt string   // Publication date (ISO 8601 format)
	ModifiedAt  string   // Last modification date (ISO 8601 format)
	Section     string   // Article section/category
	Tags        []string // Article tags

	// Structured data
	JSONLD template.HTML // JSON-LD structured data for rich snippets
}

// NewSEO creates a new SEOMetadata with sensible defaults.
func NewSEO(title, description string) *SEOMetadata {
	return &SEOMetadata{
		Title:       title,
		Description: description,
		OGType:      "website",
		TwitterCard: "summary_large_image",
	}
}

// SetCanonical sets the canonical URL for the page.
// It ensures the URL is properly formatted with the host.
func (s *SEOMetadata) SetCanonical(host, path string) *SEOMetadata {
	// Ensure host starts with https://
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	s.Canonical = host + path
	return s
}

// SetImage sets the Open Graph and Twitter image.
// If the image URL is relative, it will be made absolute using the host.
func (s *SEOMetadata) SetImage(imageURL, alt, host string) *SEOMetadata {
	// Make image URL absolute if it's relative
	if imageURL != "" && !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "https://" + host
		}
		if !strings.HasPrefix(imageURL, "/") {
			imageURL = "/" + imageURL
		}
		imageURL = host + imageURL
	}

	s.OGImage = imageURL
	s.OGImageAlt = alt
	s.TwitterImage = imageURL
	s.TwitterImageAlt = alt
	return s
}

// SetTwitter configures Twitter Card metadata.
func (s *SEOMetadata) SetTwitter(site, creator string) *SEOMetadata {
	// Ensure @ prefix
	if site != "" && !strings.HasPrefix(site, "@") {
		site = "@" + site
	}
	if creator != "" && !strings.HasPrefix(creator, "@") {
		creator = "@" + creator
	}

	s.TwitterSite = site
	s.TwitterCreator = creator
	return s
}

// SetArticle configures article-specific metadata.
func (s *SEOMetadata) SetArticle(author, publishedAt, section string, tags []string) *SEOMetadata {
	s.OGType = "article"
	s.Author = author
	s.PublishedAt = publishedAt
	s.Section = section
	s.Tags = tags
	return s
}

// SetProfile configures profile-specific metadata.
func (s *SEOMetadata) SetProfile() *SEOMetadata {
	s.OGType = "profile"
	return s
}

// WithJSONLD adds JSON-LD structured data.
// The jsonld parameter should be valid JSON-LD markup.
func (s *SEOMetadata) WithJSONLD(jsonld string) *SEOMetadata {
	s.JSONLD = template.HTML(jsonld)
	return s
}

// GetTitle returns the effective title for the page.
// It uses Title if set, otherwise returns a default.
func (s *SEOMetadata) GetTitle() string {
	if s.Title != "" {
		return s.Title
	}
	return "The Skyscape"
}

// GetOGTitle returns the Open Graph title, falling back to Title.
func (s *SEOMetadata) GetOGTitle() string {
	if s.OGTitle != "" {
		return s.OGTitle
	}
	return s.GetTitle()
}

// GetOGDescription returns the Open Graph description, falling back to Description.
func (s *SEOMetadata) GetOGDescription() string {
	if s.OGDescription != "" {
		return s.OGDescription
	}
	return s.Description
}

// GetOGURL returns the Open Graph URL, falling back to Canonical.
func (s *SEOMetadata) GetOGURL() string {
	if s.OGURL != "" {
		return s.OGURL
	}
	return s.Canonical
}

// GetTwitterTitle returns the Twitter title, falling back to OG title or Title.
func (s *SEOMetadata) GetTwitterTitle() string {
	if s.TwitterTitle != "" {
		return s.TwitterTitle
	}
	return s.GetOGTitle()
}

// GetTwitterDescription returns the Twitter description, falling back to OG description.
func (s *SEOMetadata) GetTwitterDescription() string {
	if s.TwitterDescription != "" {
		return s.TwitterDescription
	}
	return s.GetOGDescription()
}

// GetTwitterImage returns the Twitter image, falling back to OG image.
func (s *SEOMetadata) GetTwitterImage() string {
	if s.TwitterImage != "" {
		return s.TwitterImage
	}
	return s.OGImage
}

// GetTwitterImageAlt returns the Twitter image alt text, falling back to OG image alt.
func (s *SEOMetadata) GetTwitterImageAlt() string {
	if s.TwitterImageAlt != "" {
		return s.TwitterImageAlt
	}
	return s.OGImageAlt
}

// EscapeHTML escapes HTML in text for safe rendering in meta tags.
func (s *SEOMetadata) EscapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// GenerateJSONLD generates basic JSON-LD structured data for the page.
func (s *SEOMetadata) GenerateJSONLD(schemaType string) template.HTML {
	var jsonld strings.Builder
	jsonld.WriteString(`<script type="application/ld+json">`)
	jsonld.WriteString("\n{")
	jsonld.WriteString(fmt.Sprintf("\n  \"@context\": \"https://schema.org\","))
	jsonld.WriteString(fmt.Sprintf("\n  \"@type\": \"%s\",", schemaType))

	if s.Title != "" {
		jsonld.WriteString(fmt.Sprintf("\n  \"name\": \"%s\",", escapeJSON(s.Title)))
	}

	if s.Description != "" {
		jsonld.WriteString(fmt.Sprintf("\n  \"description\": \"%s\",", escapeJSON(s.Description)))
	}

	if s.OGImage != "" {
		jsonld.WriteString(fmt.Sprintf("\n  \"image\": \"%s\",", s.OGImage))
	}

	if s.Canonical != "" {
		jsonld.WriteString(fmt.Sprintf("\n  \"url\": \"%s\",", s.Canonical))
	}

	if s.Author != "" && schemaType == "Article" {
		jsonld.WriteString(fmt.Sprintf("\n  \"author\": {"))
		jsonld.WriteString(fmt.Sprintf("\n    \"@type\": \"Person\","))
		jsonld.WriteString(fmt.Sprintf("\n    \"name\": \"%s\"", escapeJSON(s.Author)))
		jsonld.WriteString(fmt.Sprintf("\n  },"))
	}

	if s.PublishedAt != "" && schemaType == "Article" {
		jsonld.WriteString(fmt.Sprintf("\n  \"datePublished\": \"%s\",", s.PublishedAt))
	}

	if s.ModifiedAt != "" && schemaType == "Article" {
		jsonld.WriteString(fmt.Sprintf("\n  \"dateModified\": \"%s\",", s.ModifiedAt))
	}

	// Remove trailing comma
	content := jsonld.String()
	if strings.HasSuffix(content, ",") {
		content = content[:len(content)-1]
	}

	return template.HTML(content + "\n}\n</script>")
}

// escapeJSON escapes text for safe inclusion in JSON.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// SEOProvider interface allows controllers to provide SEO metadata.
// Controllers can implement this interface to customize SEO for their pages.
type SEOProvider interface {
	SEO() *SEOMetadata
}

// GetSEO retrieves SEO metadata from a handler if it implements SEOProvider.
// Returns nil if the handler doesn't provide SEO metadata.
func GetSEO(handler Handler) *SEOMetadata {
	if provider, ok := handler.(SEOProvider); ok {
		return provider.SEO()
	}
	return nil
}

// BuildSitemapURL creates a properly formatted sitemap URL entry.
func BuildSitemapURL(loc string, lastmod, changefreq, priority string) string {
	var sb strings.Builder
	sb.WriteString("  <url>\n")
	sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", escapeXML(loc)))

	if lastmod != "" {
		sb.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", lastmod))
	}

	if changefreq != "" {
		sb.WriteString(fmt.Sprintf("    <changefreq>%s</changefreq>\n", changefreq))
	}

	if priority != "" {
		sb.WriteString(fmt.Sprintf("    <priority>%s</priority>\n", priority))
	}

	sb.WriteString("  </url>\n")
	return sb.String()
}

// escapeXML escapes text for safe inclusion in XML.
func escapeXML(s string) string {
	return url.QueryEscape(s)
}

// GenerateSitemap creates a complete sitemap.xml document.
func GenerateSitemap(urls []string) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	sb.WriteString("\n")

	for _, u := range urls {
		sb.WriteString(u)
	}

	sb.WriteString("</urlset>\n")
	return sb.String()
}
