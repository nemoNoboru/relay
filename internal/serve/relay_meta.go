package serve

import (
	"html"
	"regexp"
	"strings"
)

// relayMeta holds the social preview fields extracted from a fetched page.
type relayMeta struct {
	title, description, image string
}

var (
	metaTagRe  = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	metaAttrRe = regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)
	titleTagRe = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
)

// extractMeta reads the social preview fields from a page. It prefers
// Open Graph over Twitter cards over plain <title>/<meta name=description>.
func extractMeta(page string) relayMeta {
	var ogTitle, twTitle, ogDesc, twDesc, metaDesc, ogImage, twImage string
	for _, tag := range metaTagRe.FindAllString(page, -1) {
		attrs := map[string]string{}
		for _, m := range metaAttrRe.FindAllStringSubmatch(tag, -1) {
			name := strings.ToLower(m[1])
			val := m[3]
			if val == "" {
				val = m[4]
			}
			if val == "" {
				val = m[5]
			}
			attrs[name] = html.UnescapeString(val)
		}
		prop := strings.ToLower(attrs["property"])
		if prop == "" {
			prop = strings.ToLower(attrs["name"])
		}
		content := attrs["content"]
		switch prop {
		case "og:title":
			if ogTitle == "" {
				ogTitle = content
			}
		case "twitter:title":
			if twTitle == "" {
				twTitle = content
			}
		case "og:description":
			if ogDesc == "" {
				ogDesc = content
			}
		case "twitter:description":
			if twDesc == "" {
				twDesc = content
			}
		case "description":
			if metaDesc == "" {
				metaDesc = content
			}
		case "og:image", "og:image:url", "og:image:secure_url":
			if ogImage == "" {
				ogImage = content
			}
		case "twitter:image", "twitter:image:src":
			if twImage == "" {
				twImage = content
			}
		}
	}
	htmlTitle := ""
	if m := titleTagRe.FindStringSubmatch(page); m != nil {
		htmlTitle = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return relayMeta{
		title:       firstNonEmpty(ogTitle, twTitle, htmlTitle),
		description: firstNonEmpty(ogDesc, twDesc, metaDesc),
		image:       firstNonEmpty(ogImage, twImage),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
