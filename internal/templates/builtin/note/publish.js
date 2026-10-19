import { escapeHTML, page, rich, wirePreview, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");

const render = (data) => {
  const title = String(data.get("title") ?? "").trim() || "Untitled";
  const body = String(data.get("body") ?? "");
  return page({
    title,
    kicker: "Note",
    content: `<h1>${escapeHTML(title)}</h1><div>${rich(body)}</div>`,
  });
};

wirePreview(form, preview, render);
wirePublish(form, go, error, async (data) => {
  const title = String(data.get("title") ?? "").trim() || "Untitled";
  const body = String(data.get("body") ?? "");
  const plain = body.replace(/\s+/g, " ").trim().slice(0, 160);
  return {
    template: "note",
    title,
    summary: plain,
    files: [{ path: "index.html", content: render(data) }],
  };
});
