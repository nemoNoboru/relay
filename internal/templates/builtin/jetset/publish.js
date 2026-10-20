import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const input = document.getElementById("tracks");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Street Radio";
const descriptionOf = (data) => String(data.get("description") ?? "");
const trackName = (file) => String(file.name ?? "track").replace(/\.[^.]+$/, "");
const selected = () => [...input.files];

const style = `
body{background:#100a18;color:#f6f0ff;font-family:'Trebuchet MS',system-ui,sans-serif}
.rbar{background:rgba(16,10,24,.92);color:#c8ff00;border-bottom:1px solid rgba(200,255,0,.3);font-family:Figtree,system-ui,sans-serif}
.rbar a{color:#c8ff00}
main{max-width:48rem;padding-top:2.5rem;padding-bottom:3rem}
.kicker{color:#00e5ff;font-family:Figtree,system-ui,sans-serif;letter-spacing:.16em}
h1{font-family:Nunito,system-ui,sans-serif;font-weight:800;text-transform:uppercase;font-size:clamp(2.6rem,8vw,4.6rem);line-height:.9;color:#fff;border:0;margin:0 0 1.6rem;letter-spacing:.01em}
h1::before{display:none}
main>p{color:#f6f0ff;font-size:1.02rem;line-height:1.6}
.tracks{display:flex;flex-direction:column;gap:14px;margin:2rem 0 2.5rem}
.track{background:#1c1030;border:1px solid rgba(0,229,255,.4);border-radius:12px;padding:14px 16px}
.track:nth-child(even){border-color:rgba(255,45,149,.4)}
.track .num{display:block;font-family:Figtree,system-ui,sans-serif;font-size:.62rem;letter-spacing:.16em;color:#00e5ff;margin-bottom:4px}
.track:nth-child(even) .num{color:#ff2d95}
.track b{display:block;font-family:Nunito,system-ui,sans-serif;font-weight:700;text-transform:uppercase;font-size:1.2rem;letter-spacing:.03em;color:#c8ff00;margin-bottom:8px}
.track audio{width:100%;display:block}
footer{border-top:1px solid rgba(255,45,149,.4);color:#c8ff00;font-family:Figtree,system-ui,sans-serif}
footer a{color:#c8ff00}
`;

const render = (data, sources) =>
  page({
    title: titleOf(data),
    kicker: "Street radio",
    style,
    content: `<h1>${escapeHTML(titleOf(data))}</h1><div>${rich(descriptionOf(data))}</div><div class="tracks">${sources.map((src, i) => `<div class="track"><span class="num">TRACK ${String(i + 1).padStart(2, "0")}</span><b>${escapeHTML(trackName(selected()[i] ?? { name: `Track ${i + 1}` }))}</b><audio controls preload="metadata" src="${escapeHTML(src)}"></audio></div>`).join("")}</div>`,
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
    template: "jetset",
    title: titleOf(data),
    summary: descriptionOf(data).replace(/\s+/g, " ").trim().slice(0, 160),
    files: out,
  };
});
