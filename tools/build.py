#!/usr/bin/env python3
"""Generate the Relay HTML pages from their Markdown sources.

The Markdown in docs/ is the single source of truth:

    docs/vision.md      -> manifesto.html    (the public vision)
    docs/technicals.md  -> technicals.html   (how it works + deployment)

Usage:
    python3 tools/build.py           # regenerate both
    python3 tools/build.py --check   # exit 1 if either is stale (for CI)
"""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

PAGES = [
    {
        "src": ROOT / "docs" / "vision.md",
        "template": ROOT / "tools" / "manifesto.template.html",
        "out": ROOT / "manifesto.html",
        "manback": True,
    },
    {
        "src": ROOT / "docs" / "technicals.md",
        "template": ROOT / "tools" / "technicals.template.html",
        "out": ROOT / "technicals.html",
        "manback": False,
    },
]

MINI_DIAGRAM = """<div class="mini" role="img" aria-label="A browser asks for a path on one host; the public node routes the request down the matching outbound WebSocket to a friend node behind NAT.">
  <div class="flow">
    <div class="net-row"><div class="node"><b>Visitors</b>a browser asks for<small>relay.example.org/raspberry/docs/</small></div></div>
    <div class="flow-link"><span class="flow-label">HTTPS request</span><span class="wire"></span></div>
    <div class="net-row">
      <div class="node pub"><b>Public node</b>relay.example.org
        <span class="friend-list">
          <span><em>/raspberry/</em> &rarr; its outbound WebSocket</span>
          <span><em>/bookshelf/</em> &rarr; its outbound WebSocket</span>
        </span>
      </div>
    </div>
    <div class="flow-link"><span class="flow-label">relayed down the matching tunnel</span><span class="wire"></span></div>
    <div class="net-row">
      <div class="node"><b>Raspberry Pi</b>behind NAT<small>/raspberry/</small></div>
      <div class="node"><b>Laptop</b>behind NAT<small>/bookshelf/</small></div>
    </div>
  </div>
</div>"""


MANBACK = ('<div class="manback"><p>Read it, fork it, run a node. '
           'The web belongs to the people who build it.</p>'
           '<div class="actions">'
           '<a class="btn solid" href="./index.html#deploy">Deploy a site</a>'
           '<a class="btn" href="./index.html#node">Run a node</a>'
           '</div></div>')


def esc(s):
    return (s.replace("&", "&amp;").replace("<", "&lt;")
            .replace(">", "&gt;").replace('"', "&quot;"))


def inline(s):
    """Convert inline Markdown (code, links, bold, italic) to HTML."""
    s = esc(s)
    codes = []

    def stash(m):
        codes.append(m.group(1))
        return "\x00%d\x00" % (len(codes) - 1)

    s = re.sub(r"`([^`]+)`", stash, s)
    s = re.sub(r"\[([^\]]+)\]\(([^)]+)\)", r'<a href="\2">\1</a>', s)
    s = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", s)
    s = re.sub(r"(?<!\*)\*([^*]+)\*(?!\*)", r"<em>\1</em>", s)
    s = re.sub(r"_([^_]+)_", r"<em>\1</em>", s)
    for i, c in enumerate(codes):
        s = s.replace("\x00%d\x00" % i, "<code>%s</code>" % c)
    return s


def slug(s):
    return re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")


def is_sep(cells):
    return bool(cells) and all(re.match(r"^:?-{2,}:?$", c.strip()) for c in cells)


def parse(lines):
    """Tokenize the Markdown subset into blocks."""
    blocks = []
    i = 0
    n = len(lines)
    start = re.compile(r"^(#{1,6}\s|>|\||\s*[-*]\s|\s*\d+\.\s|```|---+$)")
    while i < n:
        line = lines[i]
        if line.strip() == "":
            i += 1
            continue
        if line.startswith("```"):
            info = line[3:].strip()
            i += 1
            buf = []
            while i < n and not lines[i].startswith("```"):
                buf.append(lines[i])
                i += 1
            i += 1
            blocks.append(("fence", info, "\n".join(buf)))
            continue
        m = re.match(r"^(#{1,6})\s+(.*)$", line)
        if m:
            blocks.append(("h", len(m.group(1)), m.group(2).strip()))
            i += 1
            continue
        if re.match(r"^---+$", line.strip()):
            blocks.append(("hr",))
            i += 1
            continue
        if line.lstrip().startswith(">"):
            buf = []
            while i < n and lines[i].lstrip().startswith(">"):
                buf.append(re.sub(r"^\s*>\s?", "", lines[i]))
                i += 1
            blocks.append(("quote", buf))
            continue
        if line.lstrip().startswith("|"):
            rows = []
            while i < n and lines[i].lstrip().startswith("|"):
                cells = [c.strip() for c in lines[i].strip().strip("|").split("|")]
                rows.append(cells)
                i += 1
            blocks.append(("table", rows))
            continue
        m = re.match(r"^\s*\d+\.\s+(.*)$", line)
        if m:
            items = []
            while i < n:
                mm = re.match(r"^\s*\d+\.\s+(.*)$", lines[i])
                if not mm:
                    break
                items.append(mm.group(1).strip())
                i += 1
            blocks.append(("ol", items))
            continue
        m = re.match(r"^\s*[-*]\s+(.*)$", line)
        if m:
            items = []
            while i < n:
                mm = re.match(r"^\s*[-*]\s+(.*)$", lines[i])
                if not mm:
                    break
                items.append(mm.group(1).strip())
                i += 1
            blocks.append(("ul", items))
            continue
        buf = [line.strip()]
        i += 1
        while i < n and lines[i].strip() != "" and not start.match(lines[i]):
            buf.append(lines[i].strip())
            i += 1
        blocks.append(("p", " ".join(buf)))
    return blocks


def split_lead(html):
    """Turn a leading <strong>Lead.</strong> into <b>Lead.</b>."""
    return re.sub(r"^<strong>(.*?)</strong>", r"<b>\1</b>", html, count=1)


def render_table(rows):
    sep = next((r for r in rows if is_sep(r)), None)
    clean = [r for r in rows if not is_sep(r)]
    if not clean:
        return ""
    aligns = []
    if sep:
        for c in sep:
            c = c.strip()
            if c.startswith(":") and c.endswith(":"):
                aligns.append("center")
            elif c.endswith(":"):
                aligns.append("right")
            else:
                aligns.append("left")

    def cell(tag, txt, idx):
        a = aligns[idx] if idx < len(aligns) else "left"
        style = ' style="text-align:%s"' % a if a != "left" else ""
        return "<%s%s>%s</%s>" % (tag, style, inline(txt), tag)

    head, body = clean[0], clean[1:]
    out = ["<table><thead><tr>"]
    out += [cell("th", c, i) for i, c in enumerate(head)]
    out.append("</tr></thead><tbody>")
    for r in body:
        out.append("<tr>" + "".join(cell("td", c, i) for i, c in enumerate(r)) + "</tr>")
    out.append("</tbody></table>")
    return "".join(out)


def render_block(b, section_title):
    kind = b[0]
    if kind == "h":
        lvl = b[1]
        if lvl >= 4:
            return '<h%d id="%s">%s</h%d>' % (lvl, slug(b[2]), inline(b[2]), lvl)
        return ""
    if kind == "p":
        return "<p>%s</p>" % inline(b[1])
    if kind == "quote":
        return '<p class="pq">%s</p>' % "<br>".join(inline(x) for x in b[1])
    if kind == "ol":
        return "<ol>" + "".join("<li>%s</li>" % inline(x) for x in b[1]) + "</ol>"
    if kind == "ul":
        return "<ul>" + "".join("<li>%s</li>" % inline(x) for x in b[1]) + "</ul>"
    if kind == "table":
        return render_table(b[1])
    if kind == "fence":
        if "Public node" in b[2] and "NAT" in b[2]:
            return MINI_DIAGRAM
        return '<pre class="ascii">%s</pre>' % esc(b[2])
    if kind == "hr":
        return "<hr>"
    return ""


def render_section(sec, manback=True):
    title = sec["title"]
    blocks = sec["blocks"]

    if title == "Core Principles":
        rows = [split_lead(inline(b[1])) for b in blocks if b[0] == "p"]
        return '<div class="principles">\n' + "\n".join("<div>%s</div>" % r for r in rows) + "\n</div>"

    # generic; on the manifesto a closing `---` becomes the call to action,
    # and the paragraph after it the sign-off line.
    out = []
    foot_open = False
    for b in blocks:
        if b[0] == "p" and foot_open:
            out.append('<p class="last">%s</p></div>' % inline(b[1]))
            foot_open = False
        elif b[0] == "p":
            out.append("<p>%s</p>" % inline(b[1]))
        elif b[0] == "hr" and manback:
            out.append(MANBACK + '<div class="manfoot">')
            foot_open = True
        else:
            out.append(render_block(b, title))
    if foot_open:
        out.append("</div>")
    return "\n".join(out)


def section_html(sec, manback):
    title = sec["title"]
    sid = slug(title)
    m = re.match(r"^(\d+)\.\s+(.*)$", title)
    if m:
        head = '<h3 id="%s"><span class="n">%s.</span> %s</h3>' % (sid, m.group(1), esc(m.group(2)))
    else:
        head = '<h2 id="%s">%s</h2>' % (sid, inline(title))
    return head + "\n" + render_section(sec, manback)


def toc_html(sections):
    out = ["<h4>Contents</h4>"]
    for s in sections:
        t = s["title"]
        m = re.match(r"^(\d+)\.\s+(.*)$", t)
        label = ("%s · %s" % (m.group(1), m.group(2))) if m else t
        out.append('<a href="#%s">%s</a>' % (slug(t), esc(label)))
    return "\n".join(out)


def build_page(page):
    blocks = parse(page["src"].read_text().splitlines())
    subtitle = "Relay"
    sections = []
    current = None
    for b in blocks:
        if b[0] == "h":
            lvl, txt = b[1], b[2]
            if lvl == 1:
                continue
            if lvl == 2:
                subtitle = txt
                continue
            if lvl == 3:
                current = {"title": txt, "blocks": []}
                sections.append(current)
                continue
            if current is not None:
                current["blocks"].append(("h", lvl, txt))
            continue
        if current is not None:
            current["blocks"].append(b)

    content = "\n\n".join(section_html(s, page["manback"]) for s in sections)
    toc = toc_html(sections)
    return (page["template"].read_text()
            .replace("{{SUBTITLE}}", esc(subtitle))
            .replace("{{TOC}}", toc)
            .replace("{{CONTENT}}", content))


def main():
    check = "--check" in sys.argv
    stale = []
    for page in PAGES:
        out = build_page(page)
        rel = page["out"].relative_to(ROOT)
        if check:
            cur = page["out"].read_text() if page["out"].exists() else ""
            if cur != out:
                stale.append(str(rel))
            else:
                print("%s is up to date." % rel)
        else:
            page["out"].write_text(out)
            print("wrote %s (%d bytes)" % (rel, len(out)))
    if stale:
        for rel in stale:
            print("%s is out of date — run: python3 tools/build.py" % rel, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
