import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const input = document.getElementById("images");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Carousel";
const captionOf = (data) => String(data.get("caption") ?? "");
const selected = () => [...input.files];

const style = `.carousel{position:relative;overflow:hidden;border:1px solid var(--hair);border-radius:12px;background:var(--ink);margin:1.5rem 0}.carousel-track{display:flex;transition:transform .35s ease}.carousel-slide{min-width:100%;display:block}.carousel-slide img{display:block;width:100%;height:auto;border:0}.car-prev,.car-next{position:absolute;top:50%;transform:translateY(-50%);border:1px solid var(--hair);background:rgba(20,18,16,.72);color:var(--paper);font:700 1.2rem Figtree,system-ui,sans-serif;padding:.35rem .7rem;border-radius:8px;cursor:pointer;z-index:2}.car-prev{left:10px}.car-next{right:10px}.car-dots{position:absolute;left:0;right:0;bottom:10px;display:flex;justify-content:center;gap:6px;z-index:2}.car-dot{width:10px;height:10px;border:1px solid var(--paper);border-radius:50%;background:transparent;padding:0;cursor:pointer}.car-dot.on{background:var(--red);border-color:var(--red)}`;

const SCRIPT = `(function(){document.querySelectorAll('[data-carousel]').forEach(function(root){var track=root.querySelector('.carousel-track');var slides=root.querySelectorAll('.carousel-slide');var dots=root.querySelector('.car-dots');var at=0;function go(n){at=(n+slides.length)%slides.length;track.style.transform='translateX('+(-at*100)+'%)';dots.querySelectorAll('.car-dot').forEach(function(d,i){d.classList.toggle('on',i===at);});}root.querySelector('.car-prev').onclick=function(){go(at-1);};root.querySelector('.car-next').onclick=function(){go(at+1);};slides.forEach(function(_,i){var b=document.createElement('button');b.className='car-dot'+(i===0?' on':'');b.type='button';b.setAttribute('aria-label','Image '+(i+1));b.onclick=function(){go(i);};dots.appendChild(b);});});})();`;

const render = (sources, caption) =>
  page({
    title: titleOf(dataOf()),
    kicker: "Images carousel",
    style,
    content: `<h1>${escapeHTML(titleOf(dataOf()))}</h1><div class="carousel" data-carousel><div class="carousel-track">${sources.map((src) => `<div class="carousel-slide"><img src="${escapeHTML(src)}" alt=""></div>`).join("")}</div><button class="car-prev" type="button" aria-label="Previous">&#8592;</button><button class="car-next" type="button" aria-label="Next">&#8594;</button><div class="car-dots"></div></div><div>${rich(caption)}</div><script>${SCRIPT}</script>`,
  });

const update = () => {
  const urls = selected().map((file) => URL.createObjectURL(file));
  preview.srcdoc = render(urls, captionOf(dataOf()));
};

form.addEventListener("input", update);
input.addEventListener("change", update);
update();

wirePublish(form, go, error, async (data) => {
  const files = selected();
  if (files.length === 0) throw new Error("Choose at least one image.");
  const paths = files.map((file, i) => `assets/${i + 1}-${safeName(file.name)}`);
  const out = [];
  for (let i = 0; i < files.length; i++) {
    out.push({ path: paths[i], base64: await fileToBase64(files[i]) });
  }
  out.push({ path: "index.html", content: render(paths, captionOf(data)) });
  return {
    template: "carousel",
    title: titleOf(data),
    summary: captionOf(data).replace(/\\s+/g, " ").trim().slice(0, 160),
    image: paths[0],
    files: out,
  };
});
