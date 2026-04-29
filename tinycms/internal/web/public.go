package web

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"tinycms/internal/db"
)

type Public struct {
	Store    *db.Store
	SiteName string
	md       goldmark.Markdown
}

func NewPublic(st *db.Store, siteName string) *Public {
	return &Public{Store: st, SiteName: siteName, md: goldmark.New(goldmark.WithExtensions(extension.GFM))}
}

func (p *Public) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin" {
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/admin/") || strings.HasPrefix(r.URL.Path, "/uploads/") {
		http.NotFound(w, r)
		return
	}
	routePath := "/" + strings.Trim(r.URL.Path, "/")
	if routePath == "/" {
		routePath = "/"
	}
	page, err := p.Store.GetPublishedByPath(r.Context(), routePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var b bytes.Buffer
	if err := p.md.Convert([]byte(page.Markdown), &b); err != nil {
		http.Error(w, "render error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = publicTpl.Execute(w, map[string]any{"SiteName": p.SiteName, "Title": page.Title, "MetaDescription": page.MetaDescription, "Body": template.HTML(b.String())})
}

var publicTpl = template.Must(template.New("page").Parse(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}} · {{.SiteName}}</title>{{if .MetaDescription}}<meta name="description" content="{{.MetaDescription}}">{{end}}<style>
:root{--bg:#f7f5ef;--paper:#fffefa;--ink:#263238;--muted:#667085;--accent:#3b7a57;--line:#e7e0d2}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.65 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif}main{max-width:860px;margin:48px auto;padding:32px;background:var(--paper);border:1px solid var(--line);border-radius:18px;box-shadow:0 10px 30px #00000010}a{color:var(--accent)}img{max-width:100%;height:auto;border-radius:12px}pre{overflow:auto;background:#182026;color:#f5f5f5;padding:16px;border-radius:12px}code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}header{margin-bottom:28px;color:var(--muted);font-size:14px}.brand{font-weight:700;color:var(--accent);text-decoration:none}h1,h2,h3{line-height:1.2}blockquote{border-left:4px solid var(--accent);margin-left:0;padding-left:16px;color:var(--muted)}</style></head><body><main><header><a class="brand" href="/">{{.SiteName}}</a></header><article>{{.Body}}</article></main></body></html>`))
