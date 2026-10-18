/**
 * Publishing from the browser: invite links, Markdown pages, and the tar
 * archive the deploy endpoint expects.
 */
import { marked } from "marked";
// The QR encoder is inlined into every published page so the QR is generated
// locally in the reader's browser — no third-party service, no tracking.
import qrcodeSrc from "qrcode-generator/qrcode.js?raw";

const QR_GLOBAL = qrcodeSrc.replace(/var qrcode;/, "var QRCODE=qrcode;").replace(/\bqrcode\.stringToBytesFuncs\b/g, "qrcode.stringToBytesFuncs");

export interface SiteFile {
  readonly path: string;
  readonly data: Uint8Array;
}

const enc = new TextEncoder();
const dec = new TextDecoder();

/** The key rides in the fragment, which browsers never send to the server. */
export const inviteLink = (origin: string, key: string): string =>
  `${origin}/_relay/publish.html#${new URLSearchParams({ key })}`;

export const readInvite = (hash: string) => {
  const p = new URLSearchParams(hash.replace(/^#/, ""));
  return { key: p.get("key") ?? "", site: p.get("site") ?? "" };
};

/** A message to paste into Claude or another assistant, ready to run. */
export const assistantPrompt = (origin: string, key: string, site: string): string =>
  [
    "Please publish my website with Relay.",
    "",
    "Upload my site's folder as a tar archive with this command:",
    "",
    `tar -cf - -C SITE_FOLDER . | curl --data-binary @- -H "Authorization: Bearer ${key}" ${origin}/_relay/v1/sites/${site || "SITE_NAME"}/deploy`,
    "",
    "- SITE_FOLDER is the folder with my site; it should have an index.html. Turn any Markdown files into HTML pages first.",
    site ? `- The site name is ${site}.` : "- Ask me for a site name: lowercase letters, numbers and hyphens.",
    `- The JSON reply's "url" field is the address. Give me the full link: ${origin} followed by that url.`,
    "- Every upload replaces the whole site.",
    "- The key after Bearer is private. Don't repeat it back to me or put it anywhere else.",
  ].join("\n");

const octal = (n: number, width: number) => n.toString(8).padStart(width, "0") + "\0";

const header = (path: string, size: number): Uint8Array => {
  const h = new Uint8Array(512);
  const put = (s: string, at: number) => h.set(enc.encode(s), at);
  put(path, 0);
  put(octal(0o644, 7), 100);
  put(octal(size, 11), 124);
  put(octal(Math.floor(Date.now() / 1000), 11), 136);
  put("        0", 148); // checksum counts as spaces; then typeflag "0" (file)
  put("ustar\x0000", 257);
  put(octal(h.reduce((s, b) => s + b, 0), 6) + " ", 148);
  return h;
};

/** A ustar archive. Paths must fit the 100-byte name field. */
export const tar = (files: ReadonlyArray<SiteFile>): Uint8Array => {
  const blocks = files.flatMap((f) => [header(f.path, f.data.length), f.data, new Uint8Array(-f.data.length & 511)]);
  const out = new Uint8Array(blocks.reduce((n, b) => n + b.length, 1024));
  blocks.reduce((at, b) => (out.set(b, at), at + b.length), 0);
  return out;
};

const escape = (s: string) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;");

/** The page title, for the share card. */
export const cardTitle = (md: string): string => {
  const m = md.match(/^#\s+(.+)$/m);
  return m ? m[1]!.trim() : "A page on Relay";
};

const BANNER = `
<footer class="relay-banner">
  <div class="rb-art" aria-hidden="true">
    <div class="rb-frame"></div><div class="rb-circle"></div><div class="rb-wedge"></div>
    <div class="rb-bar"></div><div class="rb-word">RELAY</div>
  </div>
  <div class="rb-body">
    <p class="rb-tag">Hosted with Relay</p>
    <p class="rb-sub">Your own corner of the web, on a server kept by people. <a href="/">Learn how</a>.</p>
    <div class="rb-actions">
      <button class="rb-btn" id="rb-share" type="button">Share</button>
      <button class="rb-btn" id="rb-qr" type="button">QR code</button>
      <button class="rb-btn" id="rb-card" type="button">Share card</button>
    </div>
    <div class="rb-pop" id="rb-pop" hidden></div>
  </div>
</footer>`;

const STYLE = `body{margin:0;background:#f0e6d2;color:#141210;font:18px/1.6 Georgia,serif}
main{max-width:40rem;margin:0 auto;padding:3rem 1.25rem}
h1,h2,h3{font-family:Impact,"Arial Narrow",sans-serif;line-height:1.1;text-transform:uppercase}
h1{font-size:2.6rem;border-bottom:6px solid #c8202f;padding-bottom:.3rem}
a{color:#c8202f}img{max-width:100%}
pre{overflow:auto;background:#141210;color:#f0e6d2;padding:1rem}
blockquote{margin:0;padding-left:1rem;border-left:4px solid #c8202f}`;

const BANNER_CSS = `.relay-banner{margin:3rem auto 0;border-top:6px solid #141210;background:#f0e6d2;padding:1.5rem 1.25rem 2rem;display:flex;gap:1.5rem;align-items:center;max-width:37.5rem}
.rb-art{position:relative;width:96px;height:96px;flex:none;border:4px solid #141210;background:#fff}
.rb-circle{position:absolute;width:56%;aspect-ratio:1;border-radius:50%;border:4px solid #141210;top:12%;left:8%}
.rb-circle::after{content:"";position:absolute;inset:20%;border-radius:50%;background:#dfa32a}
.rb-wedge{position:absolute;width:110%;height:42%;background:#ce1f2b;left:-6%;top:32%;transform:rotate(-14deg);clip-path:polygon(0 0,100% 50%,0 100%)}
.rb-bar{position:absolute;width:150%;height:8px;background:#141210;left:-25%;bottom:18%;transform:rotate(8deg)}
.rb-word{position:absolute;right:-6%;top:4%;font-family:Impact,"Arial Narrow",sans-serif;font-size:1.05rem;letter-spacing:.04em;color:#141210;background:#f0e6d2;padding:1px 2px;border-top:3px solid #ce1f2b;border-bottom:3px solid #ce1f2b}
.rb-tag{font-family:Impact,"Arial Narrow",sans-serif;text-transform:uppercase;font-size:1.5rem;line-height:1;margin:0 0 .3rem;color:#141210}
.rb-sub{margin:0 0 .9rem;font-size:.95rem}
.rb-sub a{color:#ce1f2b}
.rb-actions{display:flex;gap:.6rem;flex-wrap:wrap}
.rb-btn{font-family:ui-monospace,monospace;font-size:.7rem;letter-spacing:.1em;text-transform:uppercase;padding:.55rem .9rem;border:3px solid #141210;background:#f0e6d2;color:#141210;cursor:pointer}
.rb-btn:hover{background:#ce1f2b;color:#f0e6d2;border-color:#141210}
.rb-pop{margin-top:1rem}
.rb-pop img,.rb-pop canvas{display:block;border:3px solid #141210;background:#fff;max-width:100%}
@media(max-width:560px){.relay-banner{flex-direction:column;align-items:flex-start;gap:1rem}}`;

const BANNER_JS = `(function(){
var pop=document.getElementById('rb-pop');var url=location.href;
function show(node){pop.hidden=false;pop.replaceChildren(node);}
var title=document.querySelector('h1')?document.querySelector('h1').textContent:'A page on Relay';
document.getElementById('rb-share').onclick=function(){
  if(navigator.share){navigator.share({title:title,url:url}).catch(function(){});}
  else if(navigator.clipboard){navigator.clipboard.writeText(url);this.textContent='Link copied';var b=this;setTimeout(function(){b.textContent='Share';},2000);}
};
document.getElementById('rb-qr').onclick=function(){
  if(typeof QRCODE==='undefined')return;
  var q=QRCODE(0,'M');q.addData(url);q.make();
  var n=q.getModuleCount();var size=8;var c=document.createElement('canvas');c.width=c.height=n*size;
  var x=c.getContext('2d');x.fillStyle='#f0e6d2';x.fillRect(0,0,c.width,c.height);x.fillStyle='#141210';
  for(var r=0;r<n;r++)for(var col=0;col<n;col++)if(q.isDark(r,col))x.fillRect(col*size,r*size,size,size);
  c.style.width='200px';c.style.height='200px';show(c);
};
document.getElementById('rb-card').onclick=function(){
  var c=document.createElement('canvas');c.width=1080;c.height=1350;var x=c.getContext('2d');
  x.fillStyle='#f0e6d2';x.fillRect(0,0,1080,1350);
  x.strokeStyle='#141210';x.lineWidth=6;
  for(var i=0;i<=1080;i+=90){x.globalAlpha=.08;x.beginPath();x.moveTo(i,0);x.lineTo(i,1350);x.stroke();x.beginPath();x.moveTo(0,i);x.lineTo(1080,i);x.stroke();}
  x.globalAlpha=1;
  x.fillStyle='#dfa32a';x.beginPath();x.arc(540,560,330,0,7);x.fill();
  x.strokeStyle='#141210';x.lineWidth=10;x.beginPath();x.arc(540,560,430,0,7);x.stroke();
  x.save();x.translate(540,620);x.rotate(-0.24);x.fillStyle='#ce1f2b';
  x.beginPath();x.moveTo(-760,-90);x.lineTo(760,60);x.lineTo(-760,240);x.closePath();x.fill();x.restore();
  x.fillStyle='#141210';x.save();x.translate(540,760);x.rotate(0.14);x.fillRect(-820,-14,1640,28);x.restore();
  x.fillStyle='#141210';x.font='700 150px Impact,"Arial Narrow",sans-serif';x.textAlign='center';x.textBaseline='middle';
  var t=title;var lines=[];var words=t.split(' ');var line='';
  words.forEach(function(w){if((line+' '+w).trim().length>14&&line){lines.push(line);line=w;}else{line=(line+' '+w).trim();}});if(line)lines.push(line);
  var y0=560-((lines.length-1)*80);lines.forEach(function(l,i){x.fillText(l.toUpperCase(),540,y0+i*160);});
  x.fillStyle='#ce1f2b';x.font='700 44px ui-monospace,monospace';x.fillText('HOSTED WITH RELAY',540,1150);
  c.toBlob(function(b){var a=document.createElement('a');a.href=URL.createObjectURL(b);a.download='relay-card.png';a.click();
    var img=document.createElement('img');img.src=c.toDataURL();show(img);});
};
})();`;

const page = (md: string) =>
  `<!doctype html><html lang="en"><head><meta charset="utf-8">` +
  `<meta name="viewport" content="width=device-width, initial-scale=1">` +
  `<title>${escape(md.match(/^#\s+(.+)$/m)?.[1] ?? "Page")}</title><style>${STYLE}${BANNER_CSS}</style></head>` +
  `<body><main>${marked.parse(md, { async: false })}</main>${BANNER}<script>${QR_GLOBAL};var QRCODE=qrcode;</script><script>${BANNER_JS}</script></body></html>`;

/** Drop a folder and its own name is not part of the site. */
const unwrap = (files: ReadonlyArray<SiteFile>) => {
  const top = files[0]?.path.split("/")[0] + "/";
  return files.every((f) => f.path.startsWith(top)) ? files.map((f) => ({ ...f, path: f.path.slice(top.length) })) : files;
};

/** Markdown becomes pages; a lone Markdown file becomes the home page. */
export const toSite = (dropped: ReadonlyArray<SiteFile>): ReadonlyArray<SiteFile> => {
  const files = unwrap(dropped);
  const isMd = (f: SiteFile) => f.path.endsWith(".md");
  const home = files.filter(isMd).length === 1 && !files.some((f) => f.path === "index.html");
  return files.map((f) =>
    isMd(f) ? { path: home ? "index.html" : f.path.replace(/\.md$/, ".html"), data: enc.encode(page(dec.decode(f.data))) } : f,
  );
};

/** What to upload: a dropped tar as it is, anything else packed. */
export const body = (files: ReadonlyArray<SiteFile>): Uint8Array => {
  const [only] = files;
  return files.length === 1 && only?.path.endsWith(".tar") ? only.data : tar(toSite(files));
};
