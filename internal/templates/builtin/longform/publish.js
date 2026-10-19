import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const heroInput = document.getElementById("hero");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Longform";
const bylineOf = (data) => String(data.get("byline") ?? "").trim();
const standfirstOf = (data) => String(data.get("standfirst") ?? "").trim();
const bodyOf = (data) => String(data.get("body") ?? "");
const readingTime = (body) => Math.max(1, Math.ceil(body.trim().split(/\s+/).filter(Boolean).length / 200));

const style = `
body{background:#fbfaf7;color:#1a1a1a;font-family:Georgia,'Times New Roman',serif}
.rbar{background:#fff;color:#111;border-bottom:1px solid #dedbd4;font-family:ui-monospace,monospace;font-size:.6rem}
.rbar a{color:#111;border-bottom-color:#111}
main{max-width:42rem;padding:3.5rem 1.25rem 2rem}
.kicker{color:#777;font-family:ui-monospace,monospace;font-size:.62rem;letter-spacing:.16em}
h1{font-family:Georgia,'Times New Roman',serif;text-transform:none;font-weight:700;font-size:clamp(2.6rem,6vw,4.1rem);line-height:1.04;letter-spacing:-.025em;border:0;margin:0 0 1rem}
h1::before{display:none}
.byline{font-family:ui-monospace,monospace;font-size:.68rem;letter-spacing:.08em;text-transform:uppercase;color:#777;margin:0 0 1.7rem}
.byline b{color:#1a1a1a;font-weight:700}
main>p{font-size:1.22rem;line-height:1.68;margin:0 0 1.45rem}
.lead{font-size:1.4rem;line-height:1.5;color:#444;font-style:normal;margin-bottom:2rem}
figure{margin:2rem 0}figure img{border:0;width:100%}figcaption{color:#777;font-family:ui-monospace,monospace;font-size:.62rem;letter-spacing:.08em;text-transform:uppercase;margin-top:.5rem}
h2{font-family:Georgia,serif;text-transform:none;color:#111;font-size:1.8rem;margin:2.6rem 0 .8rem}
a{color:#1a1a1a;border-bottom:1px solid #1a1a1a}
a:hover{background:#1a1a1a;color:#fbfaf7}
blockquote{border-left:3px solid #1a1a1a;margin:2rem 0;padding:.2rem 0 .2rem 1.2rem;font-size:1.3rem;font-style:italic}
footer{border-top:1px solid #dedbd4;color:#777;font-family:ui-monospace,monospace;font-size:.58rem}
footer a{color:#777;border-bottom-color:#777}
`;

const render = (data, heroSrc) => {
  const body = bodyOf(data);
  const byline = bylineOf(data);
  const standfirst = standfirstOf(data);
  return page({
    title: titleOf(data),
    kicker: "Longform",
    style,
    content: `<h1>${escapeHTML(titleOf(data))}</h1><div class="byline">${byline ? `<b>${escapeHTML(byline)}</b> · ` : ""}${readingTime(body)} min read</div>${standfirst ? `<p class="lead">${rich(standfirst)}</p>` : ""}${heroSrc ? `<figure><img src="${escapeHTML(heroSrc)}" alt=""></figure>` : ""}<div>${rich(body)}</div>`,
  });
};

let heroURL = "";
const update = () => {
  if (heroURL) URL.revokeObjectURL(heroURL);
  heroURL = heroInput.files?.[0] ? URL.createObjectURL(heroInput.files[0]) : "";
  preview.srcdoc = render(dataOf(), heroURL);
};

form.addEventListener("input", update);
heroInput.addEventListener("change", update);
update();

wirePublish(form, go, error, async (data) => {
  const files = [];
  let heroSrc = "";
  const hero = heroInput.files?.[0];
  if (hero) {
    const name = `assets/hero-${safeName(hero.name)}`;
    files.push({ path: name, base64: await fileToBase64(hero) });
    heroSrc = name;
  }
  files.push({ path: "index.html", content: render(data, heroSrc) });
  return {
    template: "longform",
    title: titleOf(data),
    summary: standfirstOf(data).slice(0, 160),
    image: heroSrc,
    files,
  };
});
