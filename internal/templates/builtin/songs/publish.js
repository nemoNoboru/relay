import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const input = document.getElementById("tracks");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Songs";
const descriptionOf = (data) => String(data.get("description") ?? "");
const trackName = (file) => String(file.name ?? "track").replace(/\.[^.]+$/, "");
const selected = () => [...input.files];

const style = `.tracks{list-style:none;display:flex;flex-direction:column;gap:14px;margin:1.4rem 0}.track{border:3px solid var(--ink);padding:12px 14px;background:var(--paper2)}.track b{display:block;font-family:ui-monospace,monospace;font-size:.7rem;letter-spacing:.08em;text-transform:uppercase;margin-bottom:8px}.track audio{width:100%;display:block}`;

const render = (data, sources) =>
  page({
    title: titleOf(data),
    kicker: "Songs",
    style,
    content: `<h1>${escapeHTML(titleOf(data))}</h1><div>${rich(descriptionOf(data))}</div><div class="tracks">${sources.map((src, i) => `<div class="track"><b>${escapeHTML(trackName(selected()[i] ?? { name: `Track ${i + 1}` }))}</b><audio controls preload="metadata" src="${escapeHTML(src)}"></audio></div>`).join("")}</div>`,
  });

const update = () => {
  const urls = selected().map((file) => URL.createObjectURL(file));
  preview.srcdoc = render(dataOf(), urls);
};

form.addEventListener("input", update);
input.addEventListener("change", update);
update();

wirePublish(form, go, error, async (data) => {
  const files = selected();
  if (files.length === 0) throw new Error("Choose at least one audio file.");
  const paths = files.map((file, i) => `assets/${i + 1}-${safeName(file.name)}`);
  const out = [];
  for (let i = 0; i < files.length; i++) {
    out.push({ path: paths[i], base64: await fileToBase64(files[i]) });
  }
  out.push({ path: "index.html", content: render(data, paths) });
  return {
    template: "songs",
    title: titleOf(data),
    summary: descriptionOf(data).replace(/\\s+/g, " ").trim().slice(0, 160),
    files: out,
  };
});
