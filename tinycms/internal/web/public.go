package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
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
	var footer template.HTML
	if settings.FooterEnabled {
		footer, err = p.render(settings.FooterMarkdown)
		if err != nil {
			http.Error(w, "render error", 500)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = publicTpl.Execute(w, map[string]any{
		"SiteName":           settings.SiteName,
		"Title":              page.Title,
		"MetaDescription":    page.MetaDescription,
		"Body":               body,
		"Footer":             footer,
		"MenuHTML":           renderMenu(settings.Menu),
		"LogoURL":            settings.LogoURL,
		"FaviconURL":         settings.FaviconURL,
		"DefaultTheme":       settings.DefaultTheme,
		"LogoEnabled":        settings.LogoEnabled,
		"FaviconEnabled":     settings.FaviconEnabled,
		"MenuEnabled":        settings.MenuEnabled,
		"FooterEnabled":      settings.FooterEnabled,
		"ThemeToggleEnabled": settings.ThemeToggleEnabled,
		"IconsEnabled":       settings.IconsEnabled,
		"HasMermaid":         strings.Contains(page.Markdown, "```mermaid"),
	})
}

func (p *Public) render(markdown string) (template.HTML, error) {
	markdown, replacements := p.expandRichMarkdown(markdown)
	var b bytes.Buffer
	if err := p.md.Convert([]byte(markdown), &b); err != nil {
		return "", err
	}
	html := b.String()
	for token, replacement := range replacements {
		html = strings.ReplaceAll(html, "<p>"+token+"</p>", replacement)
		html = strings.ReplaceAll(html, token, replacement)
	}
	return template.HTML(html), nil
}

func (p *Public) expandRichMarkdown(markdown string) (string, map[string]string) {
	replacements := map[string]string{}
	markdown = p.expandCards(markdown, replacements)
	markdown = expandIcons(markdown, replacements)
	return markdown, replacements
}

var cardStartRe = regexp.MustCompile(`^:::card(?:\s+(.*))?$`)
var attrRe = regexp.MustCompile(`([a-zA-Z_]+)="([^"]*)"`)
var iconRe = regexp.MustCompile(`\{\{icon:([a-zA-Z0-9 -]+)\}\}`)

func (p *Public) expandCards(markdown string, replacements map[string]string) string {
	lines := strings.Split(markdown, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		matches := cardStartRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
		if matches == nil {
			out = append(out, lines[i])
			continue
		}
		attrs := parseAttrs(matches[1])
		var bodyLines []string
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != ":::"; i++ {
			bodyLines = append(bodyLines, lines[i])
		}
		var body bytes.Buffer
		if err := p.md.Convert([]byte(strings.Join(bodyLines, "\n")), &body); err != nil {
			body.WriteString(template.HTMLEscapeString(strings.Join(bodyLines, "\n")))
		}
		token := fmt.Sprintf("TINYCMS_CARD_%d", len(replacements))
		replacements[token] = renderCard(attrs, body.String())
		out = append(out, token)
	}
	return strings.Join(out, "\n")
}

func expandIcons(markdown string, replacements map[string]string) string {
	return iconRe.ReplaceAllStringFunc(markdown, func(match string) string {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{icon:"), "}}"))
		token := fmt.Sprintf("TINYCMS_ICON_%d", len(replacements))
		replacements[token] = renderIcon(name)
		return token
	})
}

func parseAttrs(raw string) map[string]string {
	attrs := map[string]string{}
	for _, match := range attrRe.FindAllStringSubmatch(raw, -1) {
		attrs[strings.ToLower(match[1])] = strings.TrimSpace(match[2])
	}
	return attrs
}

func renderCard(attrs map[string]string, body string) string {
	title := template.HTMLEscapeString(attrs["title"])
	icon := renderIcon(attrs["icon"])
	var b strings.Builder
	b.WriteString(`<section class="cms-card">`)
	if title != "" || attrs["icon"] != "" {
		b.WriteString(`<header class="cms-card-head">`)
		if attrs["icon"] != "" {
			b.WriteString(icon)
		}
		if title != "" {
			b.WriteString(`<h3>`)
			b.WriteString(title)
			b.WriteString(`</h3>`)
		}
		b.WriteString(`</header>`)
	}
	b.WriteString(`<div class="cms-card-body">`)
	b.WriteString(body)
	b.WriteString(`</div></section>`)
	return b.String()
}

func renderIcon(name string) string {
	className := sanitizeIconClass(name)
	if className == "" {
		return ""
	}
	return `<span class="cms-icon"><i class="` + template.HTMLEscapeString(className) + `" aria-hidden="true"></i></span>`
}

func sanitizeIconClass(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[^a-z0-9 -]`).ReplaceAllString(name, "")
	if name == "" {
		return ""
	}
	if strings.Contains(name, "fa-") {
		return strings.Join(strings.Fields(name), " ")
	}
	return "fa-solid fa-" + strings.ReplaceAll(name, " ", "-")
}

func renderMenu(items []db.NavItem) template.HTML {
	children := map[string][]db.NavItem{}
	for _, item := range items {
		if item.Enabled {
			children[item.ParentID] = append(children[item.ParentID], item)
		}
	}
	var b strings.Builder
	writeMenu(&b, children, "")
	return template.HTML(b.String())
}

func writeMenu(b *strings.Builder, children map[string][]db.NavItem, parentID string) {
	for _, item := range children[parentID] {
		kids := children[item.ID]
		if len(kids) > 0 {
			fmt.Fprintf(b, `<div class="navGroup"><a href="%s"%s>%s</a><div class="subnav">`, template.HTMLEscapeString(item.URL), externalAttrs(item.External), template.HTMLEscapeString(item.Label))
			writeMenu(b, children, item.ID)
			b.WriteString(`</div></div>`)
			continue
		}
		fmt.Fprintf(b, `<a href="%s"%s>%s</a>`, template.HTMLEscapeString(item.URL), externalAttrs(item.External), template.HTMLEscapeString(item.Label))
	}
}

func externalAttrs(external bool) string {
	if external {
		return ` target="_blank" rel="noopener noreferrer"`
	}
	return ""
}

var publicTpl = template.Must(template.New("page").Parse(`<!doctype html><html lang="en" data-theme="{{.DefaultTheme}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}} · {{.SiteName}}</title>{{if .MetaDescription}}<meta name="description" content="{{.MetaDescription}}">{{end}}{{if and .FaviconEnabled .FaviconURL}}<link rel="icon" href="{{.FaviconURL}}">{{end}}{{if .IconsEnabled}}<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css">{{end}}<style>
:root{--bg:#f4f7fb;--paper:#ffffff;--ink:#172033;--muted:#657085;--accent:#2563eb;--line:#d8dee9;--soft:#edf2f7;--shadow:#0f172a14}[data-theme=dark]{--bg:#0f172a;--paper:#172033;--ink:#e5edf8;--muted:#9fb0c7;--accent:#7ab7ff;--line:#2c3b52;--soft:#111827;--shadow:#00000040}*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at top left,var(--soft),var(--bg) 36rem);color:var(--ink);font:16px/1.68 ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,sans-serif}a{color:var(--accent)}img{max-width:100%;height:auto;border-radius:14px}pre{overflow:auto;background:#0b1220;color:#f5f7fb;padding:16px;border-radius:14px}code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}.site{min-height:100vh;display:flex;flex-direction:column}.top{position:sticky;top:0;z-index:5;background:color-mix(in srgb,var(--paper) 88%,transparent);backdrop-filter:blur(16px);border-bottom:1px solid var(--line)}.bar{max-width:1060px;margin:0 auto;padding:14px 22px;display:flex;align-items:center;gap:18px}.brand{display:flex;align-items:center;gap:10px;margin-right:auto;color:var(--ink);font-weight:800;text-decoration:none;letter-spacing:-.02em}.brand img{width:34px;height:34px;object-fit:contain;border-radius:8px}.nav{display:flex;align-items:center;gap:4px}.nav a{color:var(--ink);text-decoration:none;padding:8px 11px;border-radius:999px}.nav a:hover{background:var(--soft)}.navGroup{position:relative}.subnav{display:none;position:absolute;right:0;top:100%;min-width:190px;background:var(--paper);border:1px solid var(--line);border-radius:14px;padding:8px;box-shadow:0 18px 50px var(--shadow)}.navGroup:hover>.subnav,.navGroup:focus-within>.subnav{display:flex;flex-direction:column}.subnav .subnav{position:static;box-shadow:none;border:0;border-left:1px solid var(--line);border-radius:0;margin-left:12px}.theme,.menuBtn{border:1px solid var(--line);background:var(--paper);color:var(--ink);border-radius:999px;padding:8px 11px;cursor:pointer}.menuBtn{display:none}.wrap{width:min(920px,calc(100% - 32px));margin:42px auto;flex:1}.card{background:var(--paper);border:1px solid var(--line);border-radius:24px;box-shadow:0 20px 60px var(--shadow);padding:clamp(24px,5vw,52px)}article h1:first-child{margin-top:0}h1,h2,h3{line-height:1.15;letter-spacing:-.03em}blockquote{border-left:4px solid var(--accent);margin-left:0;padding-left:16px;color:var(--muted)}.cms-icon{display:inline-grid;place-items:center;color:var(--accent);margin-inline:.08em}.cms-card{border:1px solid var(--line);border-radius:20px;background:linear-gradient(180deg,var(--paper),var(--soft));padding:20px;margin:22px 0;box-shadow:0 14px 34px var(--shadow)}.cms-card-head{display:flex;align-items:center;gap:12px;margin-bottom:8px}.cms-card-head .cms-icon{width:38px;height:38px;border-radius:12px;background:color-mix(in srgb,var(--accent) 13%,transparent);font-size:18px}.cms-card h3{margin:0}.cms-card-body>*:first-child{margin-top:0}.cms-card-body>*:last-child{margin-bottom:0}.mermaid{background:var(--soft);color:var(--ink);padding:16px;border-radius:14px;overflow:auto}.foot{border-top:1px solid var(--line);color:var(--muted);padding:28px 22px;text-align:center}.foot>*{max-width:920px;margin-left:auto;margin-right:auto}.foot p{margin:.25rem auto}@media(max-width:720px){.bar{flex-wrap:wrap}.menuBtn{display:inline-block}.nav{display:none;width:100%;flex-direction:column;align-items:stretch;padding-top:10px}.nav.open{display:flex}.nav a{padding:11px 12px}.navGroup{display:flex;flex-direction:column}.subnav{position:static;display:flex;background:transparent;box-shadow:none;border:0;border-left:1px solid var(--line);border-radius:0;margin-left:14px;padding:2px 0 2px 10px}.wrap{margin:24px auto}.card{border-radius:18px}}</style></head><body><div class="site"><header class="top"><div class="bar"><a class="brand" href="/">{{if and .LogoEnabled .LogoURL}}<img src="{{.LogoURL}}" alt="">{{end}}<span>{{.SiteName}}</span></a>{{if .MenuEnabled}}<button class="menuBtn" type="button" aria-expanded="false" aria-controls="nav">Menu</button><nav class="nav" id="nav">{{.MenuHTML}}</nav>{{end}}{{if .ThemeToggleEnabled}}<button class="theme" type="button">Theme</button>{{end}}</div></header><main class="wrap"><section class="card"><article>{{.Body}}</article></section></main>{{if .FooterEnabled}}<footer class="foot">{{.Footer}}</footer>{{end}}</div>{{if .ThemeToggleEnabled}}<script>(function(){var root=document.documentElement;var saved=localStorage.getItem('tinycms-theme');if(saved){root.dataset.theme=saved}var theme=document.querySelector('.theme');if(theme){theme.onclick=function(){var next=root.dataset.theme==='dark'?'light':'dark';root.dataset.theme=next;localStorage.setItem('tinycms-theme',next)}};var btn=document.querySelector('.menuBtn');var nav=document.querySelector('#nav');if(btn&&nav){btn.onclick=function(){var open=nav.classList.toggle('open');btn.setAttribute('aria-expanded',open?'true':'false')}}})()</script>{{else}}<script>(function(){var btn=document.querySelector('.menuBtn');var nav=document.querySelector('#nav');if(btn&&nav){btn.onclick=function(){var open=nav.classList.toggle('open');btn.setAttribute('aria-expanded',open?'true':'false')}}})()</script>{{end}}{{if .HasMermaid}}<script type="module">import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';document.querySelectorAll('pre code.language-mermaid').forEach(function(code){var div=document.createElement('div');div.className='mermaid';div.textContent=code.textContent;code.parentElement.replaceWith(div)});mermaid.initialize({startOnLoad:true,theme:document.documentElement.dataset.theme==='dark'?'dark':'default'});</script>{{end}}</body></html>`))
