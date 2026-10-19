import { escapeHTML, fileToBase64, page, rich, safeName, wirePublish } from "/_relay/template/runtime.js";

const form = document.getElementById("form");
const preview = document.getElementById("preview");
const error = document.getElementById("error");
const go = document.getElementById("go");
const heroInput = document.getElementById("hero");

const dataOf = () => new FormData(form);
const titleOf = (data) => String(data.get("title") ?? "").trim() || "Article";
const standfirstOf = (data) => String(data.get("standfirst") ?? "").trim();
const bodyOf = (data) => String(data.get("body") ?? "");

const render = (data, heroSrc) =>
  page({
    title: titleOf(data),
    kicker: "Article",
    content: `<h1>${escapeHTML(titleOf(data))}</h1>${standfirstOf(data) ? `<p class="lead">${rich(standfirstOf(data))}</p>` : ""}${heroSrc ? `<figure><img src="${escapeHTML(heroSrc)}" alt=""></figure>` : ""}<div>${rich(bodyOf(data))}</div>`,
  });

let heroURL = "";
const update = () => {
  if (heroURL) URL.revokeObjectURL(heroURL);
  heroURL = heroInput.files?.[0] ? URL.createObjectURL(heroInput.files[0]) : "";
  preview.srcdoc = render(new FormData(form), heroURL);
};

form.addEventListener("input", update);
heroInput.addEventListener("change", update);
update();

wirePublish(form, go, error, async (data) => {
  const title = titleOf(data);
  const body = bodyOf(data);
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
    template: "article",
    title,
    summary: standfirstOf(data).slice(0, 160),
    image: heroSrc,
    files,
  };
});
