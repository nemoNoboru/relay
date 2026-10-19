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
.rbar{background:#000;color:#c8ff00;border-bottom:6px solid #ff2d95;font-family:ui-monospace,monospace}
.rbar a{color:#c8ff00;border-bottom-color:#ff2d95}
main{max-width:48rem;padding-top:2.5rem;padding-bottom:3rem}
.kicker{color:#00e5ff;font-family:ui-monospace,monospace;letter-spacing:.18em}
h1{font-family:Impact,'Arial Black',sans-serif;text-transform:uppercase;font-size:clamp(3rem,10vw,6rem);line-height:.82;color:#fff;text-shadow:5px 5px 0 #ff2d95,10px 10px 0 #00e5ff;border:0;transform:skew(-5deg);margin:0 0 2rem;letter-spacing:.01em}
h1::before{display:none}
main>p{color:#f6f0ff;font-size:1.05rem;line-height:1.6}
.tracks{display:flex;flex-direction:column;gap:18px;margin:2rem 0 2.5rem}
.track{background:#1c1030;border:4px solid #00e5ff;box-shadow:8px 8px 0 #ff2d95;padding:14px 16px;transform:rotate(-.5deg)}
.track:nth-child(even){border-color:#ff2d95;box-shadow:8px 8px 0 #c8ff00;transform:rotate(.6deg)}
.track .num{display:block;font-family:ui-monospace,monospace;font-size:.62rem;letter-spacing:.2em;color:#00e5ff;margin-bottom:4px}
.track:nth-child(even) .num{color:#ff2d95}
.track b{display:block;font-family:Impact,'Arial Black',sans-serif;font-weight:400;text-transform:uppercase;font-size:1.35rem;letter-spacing:.04em;color:#c8ff00;margin-bottom:8px}
.track audio{width:100%;display:block}
footer{border-top:4px solid #ff2d95;color:#c8ff00;font-family:ui-monospace,monospace}
footer a{color:#c8ff00;border-bottom-color:#ff2d95}
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
