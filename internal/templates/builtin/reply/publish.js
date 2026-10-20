import { escapeHTML, page, rich, safeHref, wirePreview, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");

const render = (data) => {
  const url = safeHref(data.get("url") ?? "");
  const body = String(data.get("body") ?? "");
  return page({
    title: "Reply",
    kicker: "Reply",
    content: `<div class="card"><p class="url"><a href="${escapeHTML(url)}" rel="nofollow noopener">${escapeHTML(url)}</a></p></div><div>${rich(body)}</div>`,
  });
};

wirePreview(form, preview, render);
wirePublish(form, go, error, async (data) => {
  const url = safeHref(data.get("url") ?? "");
  if (url === "#") throw new Error("Enter an http(s) address.");
  const body = String(data.get("body") ?? "");
  return {
    template: "reply",
    title: "Reply",
    summary: body.replace(/\s+/g, " ").trim().slice(0, 160),
    files: [{ path: "index.html", content: render(data) }],
  };
});
