package social

import (
	"strings"
	"testing"
)

func TestRenderComment(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "escapes html",
			in:   `<script>alert("x")</script>`,
			want: `<p>&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;</p>`,
		},
		{
			name: "bold and italic",
			in:   `**bold** and *italic*`,
			want: `<p><strong>bold</strong> and <em>italic</em></p>`,
		},
		{
			name: "http link",
			in:   `see [Relay](https://relay.example.org)`,
			want: `<p>see <a href="https://relay.example.org" rel="nofollow noopener">Relay</a></p>`,
		},
		{
			name: "node-relative link",
			in:   `read [about](/why)`,
			want: `<p>read <a href="/why" rel="nofollow noopener">about</a></p>`,
		},
		{
			name: "rejects protocol-relative link",
			in:   `[x](//evil.example)`,
			want: `<p>[x](//evil.example)</p>`,
		},
		{
			name: "two paragraphs become two p tags",
			in:   "one\n\ntwo",
			want: "<p>one</p><p>two</p>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RenderComment(tc.in); got != tc.want {
				t.Fatalf("RenderComment(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRenderCommentLineBreaks(t *testing.T) {
	got := RenderComment("a\nb")
	if !strings.Contains(got, "a<br>b") {
		t.Fatalf("RenderComment line break = %q", got)
	}
}
