import { body, readInvite, toSite, type SiteFile } from "../publish";

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const site = $<HTMLInputElement>("site");
const key = $<HTMLInputElement>("key");
const go = $<HTMLButtonElement>("go");
const drop = $("drop");
const preview = $<HTMLIFrameElement>("preview");
const status = $("status");
let files: ReadonlyArray<SiteFile> = [];

const MESSAGES: Record<string, string> = {
  bad_token: "That key doesn't work. Ask for a new invite link.",
  bad_name: "Site names can only use lowercase letters, numbers and hyphens.",
  bad_archive: "Something in there can't be published, like a shortcut or a link.",
  too_large: "That's over 256 MB. Make the images smaller.",
};

const say = (...parts: Array<string | Node>) => {
  status.hidden = false;
  status.replaceChildren(...parts);
};

const invite = readInvite(location.hash);
site.value = invite.site;
key.value = invite.key;
$("keyfield").hidden = !!invite.key;
key.addEventListener("change", () => {
  key.value = readInvite(key.value.split("#")[1] ?? "").key || key.value.trim();
});

const read = async (path: string, f: File): Promise<SiteFile> => ({ path, data: new Uint8Array(await f.arrayBuffer()) });

const entries = (dir: FileSystemDirectoryEntry) =>
  new Promise<FileSystemEntry[]>((ok, fail) => {
    const reader = dir.createReader();
    const all: FileSystemEntry[] = [];
    // readEntries hands out batches until it returns an empty one
    const next = () => reader.readEntries((batch) => (batch.length ? (all.push(...batch), next()) : ok(all)), fail);
    next();
  });

const walk = async (e: FileSystemEntry): Promise<SiteFile[]> =>
  e.isFile
    ? [await read(e.fullPath.slice(1), await new Promise<File>((ok, fail) => (e as FileSystemFileEntry).file(ok, fail)))]
    : (await Promise.all((await entries(e as FileSystemDirectoryEntry)).map(walk))).flat();

const take = (dropped: ReadonlyArray<SiteFile>) => {
  files = dropped;
  const home = toSite(files).find((f) => f.path === "index.html");
  preview.hidden = !home;
  if (home) preview.srcdoc = new TextDecoder().decode(home.data);
  go.disabled = !files.length;
  say(files.length ? `${files.length} file${files.length === 1 ? "" : "s"} ready.` : "Nothing to publish in there.");
};

drop.addEventListener("dragover", (e) => (e.preventDefault(), drop.classList.add("over")));
drop.addEventListener("dragleave", () => drop.classList.remove("over"));
drop.addEventListener("drop", async (e) => {
  e.preventDefault();
  drop.classList.remove("over");
  // entries must be taken before the first await, or the browser drops them;
  // without a folder entry (some browsers) fall back to the plain file
  const dropped = [...(e.dataTransfer?.items ?? [])].map((i) => i.webkitGetAsEntry() ?? i.getAsFile());
  const read1 = (x: FileSystemEntry | File | null) => (x instanceof File ? read(x.name, x).then((f) => [f]) : x ? walk(x) : []);
  take((await Promise.all(dropped.map(read1))).flat());
});

for (const id of ["files", "folder"])
  $<HTMLInputElement>(id).addEventListener("change", async (e) =>
    take(await Promise.all([...((e.target as HTMLInputElement).files ?? [])].map((f) => read(f.webkitRelativePath || f.name, f)))),
  );

$("pub").addEventListener("submit", async (e) => {
  e.preventDefault();
  go.disabled = true;
  say("Publishing…");
  try {
    const r = await fetch(`/_relay/v1/sites/${site.value}/deploy`, {
      method: "POST",
      headers: { Authorization: `Bearer ${key.value}` },
      body: body(files) as BodyInit,
    });
    const reply = await r.json();
    if (!r.ok) return say(MESSAGES[reply.error?.code] ?? reply.error?.message ?? "Something went wrong.");
    const url = new URL(reply.url, location.origin).href;
    say("Live at ", Object.assign(document.createElement("a"), { href: url, textContent: url }));
  } catch {
    say("Couldn't reach the server. Try again.");
  } finally {
    go.disabled = false;
  }
});
