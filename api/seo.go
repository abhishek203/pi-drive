package api

import (
	"net/http"
	"strings"
)

func (s *Server) serveRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	robots := strings.ReplaceAll(robotsTxtTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(robots))
}

func (s *Server) serveSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	sm := strings.ReplaceAll(sitemapTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(sm))
}

func (s *Server) serveDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(docsHTMLTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(html))
}

func (s *Server) serveBlog(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")
	tmpl, ok := blogPages[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(tmpl, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(html))
}

func (s *Server) serveVs(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/vs/")
	tmpl, ok := vsPages[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(tmpl, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(html))
}

var robotsTxtTemplate = `User-agent: *
Allow: /

User-agent: GPTBot
Allow: /

User-agent: ChatGPT-User
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: Applebot-Extended
Allow: /

Sitemap: {{SERVER_URL}}/sitemap.xml
`

var sitemapTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>{{SERVER_URL}}/</loc><changefreq>weekly</changefreq><priority>1.0</priority></url>
  <url><loc>{{SERVER_URL}}/docs</loc><changefreq>weekly</changefreq><priority>0.9</priority></url>
  <url><loc>{{SERVER_URL}}/blog/why-agents-need-files</loc><changefreq>monthly</changefreq><priority>0.8</priority></url>
  <url><loc>{{SERVER_URL}}/blog/s3-is-not-a-filesystem</loc><changefreq>monthly</changefreq><priority>0.8</priority></url>
  <url><loc>{{SERVER_URL}}/blog/sharing-files-between-agents</loc><changefreq>monthly</changefreq><priority>0.7</priority></url>
  <url><loc>{{SERVER_URL}}/vs/google-drive</loc><changefreq>monthly</changefreq><priority>0.7</priority></url>
  <url><loc>{{SERVER_URL}}/skill.md</loc><changefreq>weekly</changefreq><priority>0.6</priority></url>
</urlset>`
