import { escapeHTML, page, rich, safeHref, wirePreview, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");

const render = (data) => {
  const url = safeHref(data.get("url") ?? "");
  const title = String(data.get("title") ?? "").trim() || "A link";
  const note = String(data.get("note") ?? "");
  return page({
    title,
    kicker: "Link",
    content: `<h1>${escapeHTML(title)}</h1><div class="card"><p class="url"><a href="${escapeHTML(url)}" rel="nofollow noopener">${escapeHTML(url)}</a></p></div><div>${rich(note)}</div>`,
  });
};

wirePreview(form, preview, render);
wirePublish(form, go, error, async (data) => {
  const url = safeHref(data.get("url") ?? "");
  if (url === "#") throw new Error("Enter an http(s) address.");
  const title = String(data.get("title") ?? "").trim() || "A link";
  const note = String(data.get("note") ?? "");
  return {
    template: "link",
    title,
    summary: String(data.get("note") ?? "").replace(/\s+/g, " ").trim().slice(0, 160),
    files: [{ path: "index.html", content: render(data) }],
  };
});
