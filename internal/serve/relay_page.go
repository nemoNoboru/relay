package serve

import (
	"html"
	"strings"
)

const relayPageStyle = `:root{--ink:#141210;--paper:#f0e6d2;--surface:#fbf6e9;--red:#ce1f2b;--hair:rgba(20,18,16,.16)}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--paper);color:var(--ink);font-family:'Figtree','Nunito',system-ui,sans-serif;line-height:1.55;-webkit-font-smoothing:antialiased}
.bar{display:flex;align-items:center;justify-content:space-between;gap:12px;border-bottom:1px solid var(--hair);padding:14px 20px;background:rgba(240,230,210,.92)}
.brand{display:flex;align-items:center;gap:10px;font-family:'Nunito',system-ui,sans-serif;font-weight:800;font-size:.8rem;letter-spacing:.08em;text-decoration:none;color:var(--ink)}
.brand .mark{display:grid;place-items:center;width:28px;height:28px;background:var(--red);color:var(--paper);border-radius:8px}
.stamp{font-family:'Figtree',system-ui,sans-serif;font-size:.6rem;letter-spacing:.1em;text-transform:uppercase;border:1px solid var(--hair);border-radius:999px;padding:4px 9px;opacity:.7}
main{max-width:820px;margin:0 auto;padding:34px 22px 90px}
h1{font-family:'Nunito',system-ui,sans-serif;font-weight:700;line-height:1.05;font-size:2rem;margin:0 0 10px}
.source{font-family:'Figtree',system-ui,sans-serif;font-size:.72rem;letter-spacing:.04em;opacity:.6;margin:0 0 20px}
.source a{color:var(--ink);border-bottom:1px solid var(--red);text-decoration:none}
.media{display:flex;flex-direction:column;gap:16px}
.media img{border:1px solid var(--hair);border-radius:12px;max-width:100%;height:auto;background:#fff}
.media video{width:100%;max-height:70vh;background:#000;border-radius:12px}
.media audio{width:100%}
.card{background:var(--surface);border:1px solid var(--hair);border-radius:12px;padding:20px;max-width:640px}
.card .cover{width:100%;border-radius:9px;border:1px solid var(--hair);margin-bottom:16px}
.card h1{font-size:1.6rem}
.card .desc{opacity:.85;margin:0 0 18px;max-width:58ch}
.card .go{display:inline-block;font-size:.78rem;font-family:'Figtree',system-ui,sans-serif;color:var(--paper);background:var(--red);border-radius:8px;padding:9px 14px;text-decoration:none;word-break:break-all}`

func relayPage(title, body string) string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + `</title><style>` + relayPageStyle + `</style></head>
<body><header class="bar"><a class="brand" href="/"><span class="mark">R</span> RELAY</a><span class="stamp">relayed</span></header>
<main>` + body + `</main></body></html>`
}

func relayMediaPage(title, source string, media []relayMedia) string {
	var b strings.Builder
	b.WriteString(`<h1>` + html.EscapeString(title) + `</h1>`)
	b.WriteString(`<p class="source">Relayed from <a href="` + html.EscapeString(source) + `">` + html.EscapeString(displayHost(source)) + `</a></p>`)
	b.WriteString(`<div class="media">`)
	for _, m := range media {
		switch m.Kind {
		case relayKindVideo:
			b.WriteString(`<video src="` + m.Path + `" controls playsinline preload="metadata"></video>`)
		case relayKindAudio:
			b.WriteString(`<audio src="` + m.Path + `" controls preload="metadata"></audio>`)
		default:
			b.WriteString(`<img src="` + m.Path + `" alt="">`)
		}
	}
	b.WriteString(`</div>`)
	return relayPage(title, b.String())
}

func relayCardPage(title, description, image, source string) string {
	var b strings.Builder
	b.WriteString(`<div class="card">`)
	if image != "" {
		b.WriteString(`<img class="cover" src="` + html.EscapeString(image) + `" alt="">`)
	}
	b.WriteString(`<h1>` + html.EscapeString(title) + `</h1>`)
	if description != "" {
		b.WriteString(`<p class="desc">` + html.EscapeString(description) + `</p>`)
	}
	b.WriteString(`<a class="go" href="` + html.EscapeString(source) + `">` + html.EscapeString(source) + ` ↗</a>`)
	b.WriteString(`</div>`)
	return relayPage(title, b.String())
}
