/* Relay template runtime — shared helpers for prepublish pages and publish.js.
 * Templates are plain browser modules; the node never runs this code. */

const BASE_CSS = `:root{--ink:#141210;--paper:#f0e6d2;--paper2:#e6d8bc;--red:#ce1f2b;--ochre:#dfa32a;--hair:rgba(20,18,16,.16)}
*{box-sizing:border-box}body{margin:0;background:var(--paper);color:var(--ink);font:16px/1.65 Figtree,system-ui,sans-serif;font-weight:300}
.rbar{display:flex;justify-content:space-between;align-items:center;gap:1rem;background:rgba(240,230,210,.92);color:var(--ink);border-bottom:1px solid var(--hair);padding:.6rem 1.1rem;font-family:Figtree,system-ui,sans-serif;font-size:.62rem;letter-spacing:.08em;text-transform:uppercase}
.rbar a{color:var(--ink);text-decoration:none;opacity:.7}
.rbar a:hover{opacity:1;color:var(--red)}
.rbar b{font-family:Nunito,system-ui,sans-serif;font-weight:700;font-size:.9rem;letter-spacing:.01em;text-transform:lowercase}
main{max-width:42rem;margin:0 auto;padding:3rem 1.25rem 2rem}
h1,h2,h3{font-family:Nunito,system-ui,sans-serif;font-weight:700;line-height:1.1;letter-spacing:.01em}
h1{font-size:clamp(1.8rem,5vw,2.6rem);margin:0 0 1.1rem;padding-bottom:.5rem;border-bottom:1px solid var(--hair)}
h2{font-size:1.4rem;margin:2.2rem 0 .7rem}
p{margin:0 0 1.1rem}.lead{font-size:1.12rem;opacity:.85}
a{color:var(--red);text-decoration:none;border-bottom:1px solid var(--red)}
a:hover{border-color:var(--ink);color:var(--ink)}
img{max-width:100%;height:auto;border-radius:9px}
figure{margin:1.8rem 0}figcaption{font-family:Figtree,system-ui,sans-serif;font-size:.62rem;letter-spacing:.08em;text-transform:uppercase;margin-top:.5rem;opacity:.6}
hr{border:0;height:1px;background:var(--hair);margin:2.4rem 0}
.kicker{font-family:Figtree,system-ui,sans-serif;font-weight:600;font-size:.62rem;letter-spacing:.08em;text-transform:uppercase;color:var(--red);display:block;margin-bottom:.8rem}
.card{border:1px solid var(--hair);border-radius:12px;background:var(--paper2);padding:1.2rem;margin:1.5rem 0}
.card .url{font-family:Figtree,system-ui,sans-serif;font-size:.68rem;letter-spacing:.04em;word-break:break-all}
.gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px;margin:1.5rem 0}
.gallery img{border:1px solid var(--hair);border-radius:9px;aspect-ratio:1;object-fit:cover;width:100%}
footer{max-width:42rem;margin:3rem auto 0;padding:1.2rem 1.25rem 2rem;border-top:1px solid var(--hair);font-family:Figtree,system-ui,sans-serif;font-size:.6rem;letter-spacing:.08em;text-transform:uppercase;opacity:.6}
@media(max-width:560px){h1::before{display:none}}`;

export const escapeHTML = (s) =>
  String(s ?? "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

export const safeHref = (url) => {
  const u = String(url ?? "").trim();
  return /^https?:\/\//i.test(u) || (u.startsWith("/") && !u.startsWith("//")) ? u : "#";
};

/** The small allowed markup for generated post bodies: paragraphs, bold,
 *  italic, links, and line breaks. */
export const rich = (text) => {
  const paragraphs = String(text ?? "").replace(/\r\n/g, "\n").trim().split(/\n{2,}/);
  return paragraphs
    .filter((p) => p.trim() !== "")
    .map((p) => {
      let s = escapeHTML(p.trim());
      s = s.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, label, href) =>
        `<a href="${safeHref(href.replace(/&amp;/g, "&"))}" rel="nofollow noopener">${label}</a>`);
      s = s.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
      s = s.replace(/\*([^*]+)\*/g, "<em>$1</em>");
      s = s.replace(/_([^_]+)_/g, "<em>$1</em>");
      s = s.replace(/\n/g, "<br>");
      return `<p>${s}</p>`;
    })
    .join("");
};

export const shortDate = () => {
  const d = new Date();
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
};

/** A full static post page with the Relay quiet base style. */
export const page = ({ title, kicker = "Post", content, style = "" }) =>
  `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>${escapeHTML(title)}</title><style>${BASE_CSS}${style}</style></head><body><header class="rbar"><a href="/"><b>relay</b></a><a href="/">← the feed</a></header><main><span class="kicker">${escapeHTML(kicker)} · ${shortDate()}</span>${content}</main><footer>Posted on a Relay node · <a href="/">relay</a></footer></body></html>`;

export const showError = (el, message) => {
  if (!el) return;
  el.hidden = false;
  el.textContent = message;
};

export const hideError = (el) => {
  if (el) el.hidden = true;
};

export const safeName = (name) =>
  String(name ?? "file").toLowerCase().replace(/[^a-z0-9._-]+/g, "-").replace(/^-+|-+$/g, "").slice(-80) || "file";

export const fileToBase64 = async (file) => {
  const buf = new Uint8Array(await file.arrayBuffer());
  let binary = "";
  const chunk = 0x8000;
  for (let i = 0; i < buf.length; i += chunk) {
    binary += String.fromCharCode(...buf.subarray(i, i + chunk));
  }
  return btoa(binary);
};

export const publish = async (payload) => {
  const res = await fetch("/_relay/v1/publish", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error?.message ?? `The node answered ${res.status}.`);
  return data;
};

/** Keep a preview iframe in sync with the form. render(data) may be async. */
export const wirePreview = (form, frame, render) => {
  let generation = 0;
  const update = async () => {
    const mine = ++generation;
    try {
      const html = await render(new FormData(form));
      if (mine === generation) frame.srcdoc = html;
    } catch (err) {
      if (mine === generation) frame.srcdoc = `<pre>${escapeHTML(err.message)}</pre>`;
    }
  };
  form.addEventListener("input", update);
  form.addEventListener("change", update);
  void update();
};

/** Wire the submit button, show errors, and go to the post wrapper. */
export const wirePublish = (form, button, errorEl, build) => {
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (button) button.disabled = true;
    hideError(errorEl);
    try {
      const payload = await build(new FormData(form));
      const reply = await publish(payload);
      location.href = reply.post?.postUrl ?? "/";
    } catch (err) {
      showError(errorEl, err.message);
    } finally {
      if (button) button.disabled = false;
    }
  });
};
