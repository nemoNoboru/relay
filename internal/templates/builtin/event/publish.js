import { escapeHTML, page, rich, wirePreview, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");

const render = (data) => {
  const title = String(data.get("title") ?? "").trim() || "Event";
  const when = String(data.get("when") ?? "").trim();
  const where = String(data.get("where") ?? "").trim();
  const details = String(data.get("details") ?? "");
  const meta = [when, where].filter(Boolean).map((s) => escapeHTML(s)).join(" · ");
  return page({
    title,
    kicker: "Event",
    content: `<h1>${escapeHTML(title)}</h1>${meta ? `<p class="lead">${meta}</p>` : ""}<div>${rich(details)}</div>`,
  });
};

wirePreview(form, preview, render);
wirePublish(form, go, error, async (data) => {
  const title = String(data.get("title") ?? "").trim() || "Event";
  const details = String(data.get("details") ?? "");
  return {
    template: "event",
    title,
    summary: details.replace(/\s+/g, " ").trim().slice(0, 160),
    files: [{ path: "index.html", content: render(data) }],
  };
});
