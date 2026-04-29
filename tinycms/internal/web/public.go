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
	settings, err := p.Store.GetSettings(r.Context(), p.SiteName)
	if err != nil {
		http.Error(w, "settings error", 500)
		return
	}
	body, err := p.render(page.Markdown)
	if err != nil {
		http.Error(w, "render error", 500)
		return
	}
	footer, err := p.render(settings.FooterMarkdown)
	if err != nil {
		http.Error(w, "render error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = publicTpl.Execute(w, map[string]any{
		"SiteName":        settings.SiteName,
		"Title":           page.Title,
		"MetaDescription": page.MetaDescription,
		"Body":            body,
		"Footer":          footer,
		"Menu":            settings.Menu,
		"LogoURL":         settings.LogoURL,
		"FaviconURL":      settings.FaviconURL,
		"DefaultTheme":    settings.DefaultTheme,
	})
}

func (p *Public) render(markdown string) (template.HTML, error) {
	var b bytes.Buffer
	if err := p.md.Convert([]byte(markdown), &b); err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}

var publicTpl = template.Must(template.New("page").Parse(`<!doctype html><html lang="en" data-theme="{{.DefaultTheme}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}} · {{.SiteName}}</title>{{if .MetaDescription}}<meta name="description" content="{{.MetaDescription}}">{{end}}{{if .FaviconURL}}<link rel="icon" href="{{.FaviconURL}}">{{end}}<style>
:root{--bg:#f4f7fb;--paper:#ffffff;--ink:#172033;--muted:#657085;--accent:#2563eb;--line:#d8dee9;--soft:#edf2f7;--shadow:#0f172a14}[data-theme=dark]{--bg:#0f172a;--paper:#172033;--ink:#e5edf8;--muted:#9fb0c7;--accent:#7ab7ff;--line:#2c3b52;--soft:#111827;--shadow:#00000040}*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at top left,var(--soft),var(--bg) 36rem);color:var(--ink);font:16px/1.68 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif}a{color:var(--accent)}img{max-width:100%;height:auto;border-radius:14px}pre{overflow:auto;background:#0b1220;color:#f5f7fb;padding:16px;border-radius:14px}code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}.site{min-height:100vh;display:flex;flex-direction:column}.top{position:sticky;top:0;z-index:5;background:color-mix(in srgb,var(--paper) 88%,transparent);backdrop-filter:blur(16px);border-bottom:1px solid var(--line)}.bar{max-width:1060px;margin:0 auto;padding:14px 22px;display:flex;align-items:center;gap:18px}.brand{display:flex;align-items:center;gap:10px;margin-right:auto;color:var(--ink);font-weight:800;text-decoration:none;letter-spacing:-.02em}.brand img{width:34px;height:34px;object-fit:contain;border-radius:8px}.nav{display:flex;align-items:center;gap:4px}.nav a{color:var(--ink);text-decoration:none;padding:8px 11px;border-radius:999px}.nav a:hover{background:var(--soft)}.theme,.menuBtn{border:1px solid var(--line);background:var(--paper);color:var(--ink);border-radius:999px;padding:8px 11px;cursor:pointer}.menuBtn{display:none}.wrap{width:min(920px,calc(100% - 32px));margin:42px auto;flex:1}.card{background:var(--paper);border:1px solid var(--line);border-radius:24px;box-shadow:0 20px 60px var(--shadow);padding:clamp(24px,5vw,52px)}article h1:first-child{margin-top:0}h1,h2,h3{line-height:1.15;letter-spacing:-.03em}blockquote{border-left:4px solid var(--accent);margin-left:0;padding-left:16px;color:var(--muted)}.foot{border-top:1px solid var(--line);color:var(--muted);padding:28px 22px;text-align:center}.foot>*{max-width:920px;margin-left:auto;margin-right:auto}.foot p{margin:.25rem auto}@media(max-width:720px){.bar{flex-wrap:wrap}.menuBtn{display:inline-block}.nav{display:none;width:100%;flex-direction:column;align-items:stretch;padding-top:10px}.nav.open{display:flex}.nav a{padding:11px 12px}.wrap{margin:24px auto}.card{border-radius:18px}}</style></head><body><div class="site"><header class="top"><div class="bar"><a class="brand" href="/">{{if .LogoURL}}<img src="{{.LogoURL}}" alt="">{{end}}<span>{{.SiteName}}</span></a><button class="menuBtn" type="button" aria-expanded="false" aria-controls="nav">Menu</button><nav class="nav" id="nav">{{range .Menu}}<a href="{{.URL}}"{{if .External}} target="_blank" rel="noopener noreferrer"{{end}}>{{.Label}}</a>{{end}}</nav><button class="theme" type="button">Theme</button></div></header><main class="wrap"><section class="card"><article>{{.Body}}</article></section></main><footer class="foot">{{.Footer}}</footer></div><script>(function(){var root=document.documentElement;var saved=localStorage.getItem('tinycms-theme');if(saved){root.dataset.theme=saved}document.querySelector('.theme').onclick=function(){var next=root.dataset.theme==='dark'?'light':'dark';root.dataset.theme=next;localStorage.setItem('tinycms-theme',next)};var btn=document.querySelector('.menuBtn');var nav=document.querySelector('#nav');btn.onclick=function(){var open=nav.classList.toggle('open');btn.setAttribute('aria-expanded',open?'true':'false')}}())</script></body></html>`))
