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

const BAR = `
<header class="relay-bar">
  <a class="rbar-brand" href="/"><span class="rbar-mark">R</span> RELAY</a>
  <a class="rbar-note" href="/">Hosted with Relay <span>· Know how</span></a>
</header>`;

const BANNER = `
<footer class="relay-banner">
  <div class="rb-art" aria-hidden="true">
    <div class="rb-circle"></div><div class="rb-wedge"></div>
    <div class="rb-bar"></div><div class="rb-word">RELAY</div>
  </div>
  <div class="rb-body">
    <p class="rb-tag">Made by a person,<br><em>not a platform</em></p>
    <p class="rb-sub">This page sits on a server kept by people, not rented from a company. Yours can too. <a href="/">Know how</a>.</p>
    <div class="rb-actions">
      <button class="rb-btn solid" id="rb-share" type="button">Share this page</button>
      <button class="rb-btn" id="rb-qr" type="button">QR code</button>
      <button class="rb-btn" id="rb-card" type="button">Poster</button>
    </div>
    <div class="rb-pop" id="rb-pop" hidden></div>
  </div>
</footer>`;

const STYLE = `:root{--ink:#141210;--paper:#f0e6d2;--paper2:#e6dac0;--red:#ce1f2b;--ochre:#dfa32a}
*{box-sizing:border-box}
body{margin:0;background:var(--paper);color:var(--ink);font:18px/1.65 Georgia,"Times New Roman",serif}
main{max-width:40rem;margin:0 auto;padding:3.5rem 1.25rem 1rem;position:relative}
h1,h2,h3,h4{font-family:Impact,"Haettenschweiler","Franklin Gothic Bold","Arial Black","Liberation Sans Narrow","DejaVu Sans Condensed","Arial Narrow",sans-serif;line-height:1;text-transform:uppercase;letter-spacing:.01em;font-weight:400}
h1{font-size:clamp(2.4rem,7vw,3.8rem);margin:0 0 1.4rem;padding-bottom:.4rem;position:relative}
h1::after{content:"";display:block;height:8px;background:var(--ink);margin-top:.5rem}
h1::before{content:"";position:absolute;left:-2.6rem;top:.55em;width:1.7rem;height:1.1rem;background:var(--red);clip-path:polygon(0 0,100% 50%,0 100%)}
h2{font-size:clamp(1.5rem,4vw,2rem);margin:2.6rem 0 .9rem}
h2::after{content:"";display:block;width:3.4rem;height:5px;background:var(--red);margin-top:.45rem}
h3{font-size:1.25rem;margin:2rem 0 .6rem;color:var(--red)}
p{margin:0 0 1.1rem}
a{color:var(--red);text-decoration:none;border-bottom:2px solid var(--red)}
a:hover{background:var(--red);color:var(--paper)}
img{max-width:100%;height:auto;border:3px solid var(--ink)}
strong{font-weight:700}
hr{border:0;height:6px;background:var(--ink);margin:2.5rem 0}
pre{overflow:auto;background:var(--ink);color:var(--paper);padding:1.1rem;font-size:.85rem;line-height:1.5;border-left:8px solid var(--red)}
code{font-family:ui-monospace,"Courier New",monospace;font-size:.88em;background:var(--ink);color:var(--paper);padding:1px 5px}
pre code{background:none;color:inherit;padding:0}
blockquote{margin:1.8rem 0;padding:1.1rem 1.3rem;background:var(--ink);color:var(--paper);font-family:Impact,"Haettenschweiler","Arial Black","Liberation Sans Narrow",sans-serif;text-transform:uppercase;font-size:clamp(1.1rem,3vw,1.5rem);line-height:1.15;transform:rotate(-.7deg)}
blockquote p{margin:0}
ul,ol{margin:0 0 1.2rem;padding:0;list-style:none}
ul>li{position:relative;padding:.4rem 0 .4rem 1.7rem;border-bottom:1px dashed rgba(20,18,16,.25)}
ul>li::before{content:"→";position:absolute;left:0;color:var(--red);font-family:ui-monospace,monospace}
ol{counter-reset:n}
ol>li{counter-increment:n;position:relative;padding:.4rem 0 .4rem 2.3rem;border-bottom:1px dashed rgba(20,18,16,.25)}
ol>li::before{content:counter(n,decimal-leading-zero);position:absolute;left:0;top:.35rem;font-family:Impact,sans-serif;color:var(--red)}
table{width:100%;border-collapse:collapse;margin:1.5rem 0;font-size:.95rem}
th{font-family:ui-monospace,monospace;font-size:.7rem;letter-spacing:.1em;text-transform:uppercase;text-align:left;border-bottom:3px solid var(--ink);padding:.6rem .5rem;color:var(--red)}
td{padding:.6rem .5rem;border-bottom:1px dashed rgba(20,18,16,.25)}
@media(max-width:560px){h1::before{display:none}}`;

const BANNER_CSS = `.relay-bar{position:sticky;top:0;z-index:9;display:flex;align-items:center;justify-content:space-between;gap:1rem;
  background:var(--ink);color:var(--paper);border-bottom:4px solid var(--red);padding:.6rem 1.1rem;font-family:ui-monospace,monospace}
.rbar-brand{display:flex;align-items:center;gap:.5rem;font-family:Impact,"Arial Narrow",sans-serif;font-size:1.2rem;letter-spacing:.06em;
  color:var(--paper);border:0;text-decoration:none}
.rbar-brand:hover{background:none;color:var(--paper)}
.rbar-mark{display:grid;place-items:center;width:1.55rem;height:1.55rem;background:var(--red);color:var(--paper);font-size:1.05rem}
.rbar-note{font-size:.63rem;letter-spacing:.14em;text-transform:uppercase;color:var(--paper);border:0;text-decoration:none;
  border-bottom:2px solid var(--red);padding-bottom:2px}
.rbar-note:hover{background:var(--red);color:var(--paper)}
.rbar-note span{color:var(--ochre)}
@media(max-width:460px){.rbar-note span{display:none}}

.relay-banner{margin:4rem auto 0;background:var(--ink);color:var(--paper);border-top:8px solid var(--red);
  padding:2rem 1.25rem 2.2rem;display:flex;gap:1.8rem;align-items:center;justify-content:center;flex-wrap:wrap}
.rb-art{position:relative;width:118px;height:118px;flex:none;border:4px solid var(--paper);background:var(--paper2)}
.rb-circle{position:absolute;width:58%;aspect-ratio:1;border-radius:50%;border:4px solid var(--ink);top:12%;left:8%}
.rb-circle::after{content:"";position:absolute;inset:18%;border-radius:50%;background:var(--ochre)}
.rb-wedge{position:absolute;width:120%;height:42%;background:var(--red);left:-8%;top:32%;transform:rotate(-14deg);clip-path:polygon(0 0,100% 50%,0 100%)}
.rb-bar{position:absolute;width:160%;height:9px;background:var(--ink);left:-30%;bottom:17%;transform:rotate(8deg)}
.rb-word{position:absolute;right:4%;top:4%;font-family:Impact,"Arial Narrow",sans-serif;font-size:1rem;letter-spacing:.03em;
  color:var(--ink);background:var(--paper);padding:1px 3px;border-top:3px solid var(--red);border-bottom:3px solid var(--red)}
.rb-body{max-width:26rem}
.rb-tag{font-family:Impact,"Arial Narrow",sans-serif;text-transform:uppercase;font-size:1.75rem;line-height:1;margin:0 0 .5rem}
.rb-tag em{font-style:normal;color:var(--red)}
.rb-sub{margin:0 0 1rem;font-size:.95rem;font-family:Georgia,serif}
.rb-sub a{color:var(--ochre);border-bottom:2px solid var(--red)}
.rb-sub a:hover{background:var(--red);color:var(--paper)}
.rb-actions{display:flex;gap:.55rem;flex-wrap:wrap}
.rb-btn{font-family:ui-monospace,monospace;font-size:.68rem;letter-spacing:.12em;text-transform:uppercase;padding:.6rem 1rem;
  border:3px solid var(--paper);background:transparent;color:var(--paper);cursor:pointer;transition:transform .1s}
.rb-btn.solid{background:var(--red);border-color:var(--red)}
.rb-btn:hover{background:var(--paper);color:var(--ink);border-color:var(--paper);transform:translate(-2px,-2px)}
.rb-pop{margin-top:1.2rem;flex-basis:100%;display:flex;flex-direction:column;align-items:center;gap:.7rem}
.rb-pop img,.rb-pop canvas{display:block;border:4px solid var(--paper);background:var(--paper);max-width:100%;height:auto}
.rb-hint{font-family:ui-monospace,monospace;font-size:.62rem;letter-spacing:.14em;text-transform:uppercase;color:var(--ochre);margin:0}
@media(max-width:560px){.relay-banner{flex-direction:column;align-items:flex-start;gap:1.2rem}.rb-pop{align-items:flex-start}}`;

const BANNER_JS = `(function(){
var INK='#141210',PAPER='#f0e6d2',RED='#ce1f2b',OCHRE='#dfa32a';
var DISPLAY='Impact,"Haettenschweiler","Franklin Gothic Bold","Arial Black","Liberation Sans Narrow",sans-serif',MONO='ui-monospace,"Courier New",monospace';
var pop=document.getElementById('rb-pop'),url=location.href;
var h1=document.querySelector('h1');
var title=h1?h1.textContent.trim():'A page on Relay';
var host=location.host+location.pathname.replace(/\\/$/,'');

function show(node,hint){
  pop.hidden=false;pop.replaceChildren(node);
  if(hint){var p=document.createElement('p');p.className='rb-hint';p.textContent=hint;pop.appendChild(p);}
}
function download(canvas,name,label){
  var a=document.createElement('a');a.className='rb-btn';a.textContent=label;a.download=name;
  canvas.toBlob(function(b){a.href=URL.createObjectURL(b);});
  return a;
}
function qr(text){var q=QRCODE(0,'M');q.addData(text);q.make();return q;}

/* QR drawn with square "pixels", a red registration eye and a paper quiet zone. */
function drawQR(x,q,ox,oy,px){
  var n=q.getModuleCount();
  for(var r=0;r<n;r++)for(var c=0;c<n;c++){
    if(!q.isDark(r,c))continue;
    var eye=(r<7&&c<7)||(r<7&&c>=n-7)||(r>=n-7&&c<7);
    x.fillStyle=eye?RED:INK;
    x.fillRect(ox+c*px,oy+r*px,px,px);
  }
}

document.getElementById('rb-share').onclick=function(){
  var b=this;
  if(navigator.share){navigator.share({title:title,url:url}).catch(function(){});return;}
  if(navigator.clipboard){navigator.clipboard.writeText(url);b.textContent='Link copied';
    setTimeout(function(){b.textContent='Share this page';},2000);}
};

document.getElementById('rb-qr').onclick=function(){
  if(typeof QRCODE==='undefined')return;
  var q=qr(url),n=q.getModuleCount(),px=10,pad=px*4,size=n*px+pad*2;
  var c=document.createElement('canvas');c.width=size;c.height=size+88;
  var x=c.getContext('2d');
  x.fillStyle=PAPER;x.fillRect(0,0,c.width,c.height);
  x.fillStyle=INK;x.fillRect(0,size,c.width,88);
  drawQR(x,q,pad,pad,px);
  x.fillStyle=PAPER;x.font='700 22px '+MONO;x.textAlign='center';x.textBaseline='middle';
  x.fillText('SCAN TO READ',c.width/2,size+32);
  x.fillStyle=OCHRE;x.font='16px '+MONO;
  x.fillText(host.length>38?host.slice(0,37)+'…':host,c.width/2,size+60);
  c.style.width='240px';
  var wrap=document.createElement('div');wrap.style.cssText='display:flex;flex-direction:column;align-items:center;gap:.7rem';
  wrap.append(c,download(c,'relay-qr.png','Download QR'));
  show(wrap,'Point a camera at it');
};

document.getElementById('rb-card').onclick=function(){
  var W=1080,H=1350,c=document.createElement('canvas');c.width=W;c.height=H;
  var x=c.getContext('2d');

  x.fillStyle=PAPER;x.fillRect(0,0,W,H);
  x.strokeStyle=INK;x.lineWidth=2;x.globalAlpha=.10;
  for(var i=0;i<=Math.max(W,H);i+=60){
    x.beginPath();x.moveTo(i,0);x.lineTo(i,H);x.stroke();
    x.beginPath();x.moveTo(0,i);x.lineTo(W,i);x.stroke();
  }
  x.globalAlpha=1;

  /* ochre sun, its outline ring, the red wedge and the black bar */
  x.fillStyle=OCHRE;x.beginPath();x.arc(700,430,250,0,7);x.fill();
  x.strokeStyle=INK;x.lineWidth=9;x.beginPath();x.arc(700,430,320,0,7);x.stroke();
  x.save();x.translate(0,470);x.rotate(-0.18);x.fillStyle=RED;
  x.beginPath();x.moveTo(-60,-150);x.lineTo(880,10);x.lineTo(-60,180);x.closePath();x.fill();x.restore();
  x.save();x.translate(540,600);x.rotate(0.1);x.fillStyle=INK;x.fillRect(-640,-11,1280,22);x.restore();

  /* title: shrink until every line fits, then stack from the bottom block */
  var max=W-160,size=132,lines=[];
  for(;size>52;size-=6){
    x.font='700 '+size+'px '+DISPLAY;
    lines=[];var line='',words=title.toUpperCase().split(/\\s+/),ok=true;
    for(var w=0;w<words.length;w++){
      var t=(line?line+' ':'')+words[w];
      if(x.measureText(t).width>max&&line){lines.push(line);line=words[w];}else{line=t;}
      if(x.measureText(line).width>max){ok=false;break;}
    }
    if(line)lines.push(line);
    /* the title band runs from the bar down to the keyline; keep inside it */
    if(ok&&lines.length<=4&&lines.length*size*1.02<=380)break;
  }
  x.textAlign='left';x.textBaseline='alphabetic';
  var lh=size*1.02,top=1010-(lines.length-1)*lh;
  x.fillStyle=INK;
  lines.forEach(function(l,i){x.fillText(l,80,top+i*lh);});

  /* red keyline under the title, then the footer slab */
  x.fillStyle=RED;x.fillRect(80,1040,W-160,12);
  x.fillStyle=INK;x.fillRect(0,1150,W,200);

  var q=qr(url),n=q.getModuleCount(),px=Math.floor(150/n),qs=n*px;
  x.fillStyle=PAPER;x.fillRect(W-80-qs-16,1175-8,qs+32,qs+32);
  drawQR(x,q,W-80-qs,1175,px);

  x.fillStyle=PAPER;x.font='700 46px '+DISPLAY;x.textAlign='left';x.textBaseline='middle';
  x.fillText('HOSTED WITH RELAY',80,1215);
  x.fillStyle=OCHRE;x.font='24px '+MONO;
  x.fillText(host.length>34?host.slice(0,33)+'…':host,80,1268);

  var img=document.createElement('img');img.src=c.toDataURL();img.alt='Share poster for this page';
  img.style.cssText='width:min(340px,100%)';
  var wrap=document.createElement('div');wrap.style.cssText='display:flex;flex-direction:column;align-items:center;gap:.7rem';
  wrap.append(img,download(c,'relay-poster.png','Download poster'));
  show(wrap,'1080×1350 — fits Instagram and TikTok');
};
})();`;

const page = (md: string) =>
  `<!doctype html><html lang="en"><head><meta charset="utf-8">` +
  `<meta name="viewport" content="width=device-width, initial-scale=1">` +
  `<title>${escape(md.match(/^#\s+(.+)$/m)?.[1] ?? "Page")}</title><style>${STYLE}${BANNER_CSS}</style></head>` +
  `<body>${BAR}<main>${marked.parse(md, { async: false })}</main>${BANNER}` +
  `<script>${QR_GLOBAL};var QRCODE=qrcode;</script><script>${BANNER_JS}</script></body></html>`;

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
