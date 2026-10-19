package social

import (
	"html"
	"regexp"
	"strings"
)

var (
	commentLinkRE   = regexp.MustCompile(`\[([^\]]+)\]\(((?:https?://|/)[^\s)]+)\)`)
	commentBoldRE   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	commentItalicRE = regexp.MustCompile(`\*([^*]+)\*`)
	commentUnderRE  = regexp.MustCompile(`_([^_]+)_`)
)

// RenderComment renders the allowed comment subset: paragraphs, line breaks,
// bold, italic, and links to http(s) or node-relative URLs. Everything is
// escaped first, so user HTML can never become markup.
func RenderComment(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	paragraphs := strings.Split(strings.TrimSpace(body), "\n\n")
	var out strings.Builder
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		s := html.EscapeString(paragraph)
		s = commentLinkRE.ReplaceAllStringFunc(s, func(m string) string {
			parts := commentLinkRE.FindStringSubmatch(m)
			if len(parts) != 3 {
				return m
			}
			href := parts[2]
			// Reject protocol-relative links; they would leave the node.
			if strings.HasPrefix(href, "//") {
				return m
			}
			return `<a href="` + href + `" rel="nofollow noopener">` + parts[1] + `</a>`
		})
		s = commentBoldRE.ReplaceAllString(s, "<strong>$1</strong>")
		s = commentItalicRE.ReplaceAllString(s, "<em>$1</em>")
		s = commentUnderRE.ReplaceAllString(s, "<em>$1</em>")
		s = strings.ReplaceAll(s, "\n", "<br>")
		out.WriteString("<p>")
		out.WriteString(s)
		out.WriteString("</p>")
	}
	return out.String()
}
