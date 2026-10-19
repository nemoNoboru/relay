import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const input = document.getElementById("images");

const selected = () => [...input.files];

const titleOf = () => String(new FormData(form).get("title") ?? "").trim() || "Gallery";
const captionOf = () => String(new FormData(form).get("caption") ?? "");

const render = (images, caption) =>
  page({
    title: titleOf(),
    kicker: "Gallery",
    content: `<h1>${escapeHTML(titleOf())}</h1><div class="gallery">${images.map((src) => `<img src="${escapeHTML(src)}" alt="">`).join("")}</div><div>${rich(caption)}</div>`,
  });

const update = () => {
  const urls = selected().map((file) => URL.createObjectURL(file));
  preview.srcdoc = render(urls, captionOf());
};

form.addEventListener("input", update);
input.addEventListener("change", update);
update();

wirePublish(form, go, error, async () => {
  const files = selected();
  if (files.length === 0) throw new Error("Choose at least one image.");
  const paths = files.map((file, i) => `assets/${i + 1}-${safeName(file.name)}`);
  const out = [];
  for (let i = 0; i < files.length; i++) {
    out.push({ path: paths[i], base64: await fileToBase64(files[i]) });
  }
  const caption = captionOf();
  out.push({ path: "index.html", content: render(paths, caption) });
  return {
    template: "gallery",
    title: titleOf(),
    summary: caption.replace(/\s+/g, " ").trim().slice(0, 160),
    image: paths[0],
    files: out,
  };
});
