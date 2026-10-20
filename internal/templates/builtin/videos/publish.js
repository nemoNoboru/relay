import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const input = document.getElementById("videos");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Videos";
const descriptionOf = (data) => String(data.get("description") ?? "");
const clipName = (file) => String(file.name ?? "clip").replace(/\.[^.]+$/, "");
const selected = () => [...input.files];

const style = `.videos{display:flex;flex-direction:column;gap:16px;margin:1.4rem 0}.clip{border:1px solid var(--hair);border-radius:12px;background:var(--paper2);padding:10px}.clip video{display:block;width:100%;background:#000;border-radius:9px}.clip b{display:block;font-family:Figtree,system-ui,sans-serif;font-size:.68rem;letter-spacing:.08em;text-transform:uppercase;margin:0 0 8px}`;

const render = (data, sources) =>
  page({
    title: titleOf(data),
    kicker: "Videos",
    style,
    content: `<h1>${escapeHTML(titleOf(data))}</h1><div>${rich(descriptionOf(data))}</div><div class="videos">${sources.map((src, i) => `<div class="clip"><b>${escapeHTML(clipName(selected()[i] ?? { name: `Clip ${i + 1}` }))}</b><video controls preload="metadata" src="${escapeHTML(src)}"></video></div>`).join("")}</div>`,
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
  if (files.length === 0) throw new Error("Choose at least one video file.");
  const paths = files.map((file, i) => `assets/${i + 1}-${safeName(file.name)}`);
  const out = [];
  for (let i = 0; i < files.length; i++) {
    out.push({ path: paths[i], base64: await fileToBase64(files[i]) });
  }
  out.push({ path: "index.html", content: render(data, paths) });
  return {
    template: "videos",
    title: titleOf(data),
    summary: descriptionOf(data).replace(/\\s+/g, " ").trim().slice(0, 160),
    files: out,
  };
});
