import { Effect } from "effect";

import { Client } from "../api/client";
import { establishSession, fetchMe } from "../member/ops";
import { fetchTemplates } from "../templates/ops";
import { assistantPrompt, body, inviteLink, readInvite, toSite, type SiteFile } from "../publish";

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const site = $<HTMLInputElement>("site");
const key = $<HTMLInputElement>("key");
const go = $<HTMLButtonElement>("go");
const drop = $("drop");
const preview = $<HTMLIFrameElement>("preview");
const status = $("status");
let files: ReadonlyArray<SiteFile> = [];

const MESSAGES: Record<string, string> = {
  bad_token: "That invitation doesn't work. Ask for a new invite link.",
  bad_name: "Page names can only use lowercase letters, numbers and hyphens.",
  bad_archive: "Something in there can't be published, like a shortcut or a link.",
  too_large: "That's over 256 MB. Make the images smaller.",
};

const say = (...parts: Array<string | Node>) => {
  status.hidden = false;
  status.replaceChildren(...parts);
};

const keyfield = $("keyfield");
const invited = $("invited");
const invitesection = $("invite");

const run = <A>(eff: Effect.Effect<A, unknown, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())) as Effect.Effect<A, never, never>);

let templatesLoaded = false;
const loadTemplates = async (): Promise<void> => {
  if (templatesLoaded) return;
  try {
    const reply = await run(fetchTemplates());
    if (reply.templates.length === 0) return;
    const grid = $("templates");
    grid.replaceChildren(
      ...reply.templates.map((t) => {
        const card = document.createElement("a");
        card.className = "template-card";
        card.href = `/_relay/templates/${encodeURIComponent(t.id)}/prepublish.html`;
        const tag = document.createElement("em");
        tag.textContent = t.builtin ? "Built in" : "Installed";
        const name = document.createElement("b");
        name.textContent = t.name;
        const desc = document.createElement("span");
        desc.textContent = t.description;
        card.append(tag, name, desc);
        return card;
      }),
    );
    $("templates-section").hidden = false;
    templatesLoaded = true;
  } catch {
    // No session or no templates: the free-form publish card still works.
  }
};

/** The browser already has the member cookie: no key is needed to publish. */
const signedIn = (): void => {
  keyfield.hidden = true;
  key.required = false;
  invited.hidden = false;
  invited.textContent = "You're signed in";
  invitesection.hidden = false;
  void loadTemplates();
};

/** Exchange a token for the member cookie. The page keeps working either
 *  way; a bad token still reports on publish. */
const establish = async (token: string): Promise<boolean> => {
  if (!token) return false;
  try {
    await run(establishSession(token));
    return true;
  } catch {
    return false;
  }
};

/** A rejected invite falls back to the visible key field. */
const inviteFailed = (): void => {
  keyfield.hidden = false;
  key.required = true;
  invitesection.hidden = false;
  say(MESSAGES.bad_token ?? "That invitation doesn't work.");
};

const invite = readInvite(location.hash);
site.value = invite.site;
key.value = invite.key;
keyfield.hidden = !!invite.key;
invited.hidden = !invite.key;
invitesection.hidden = !invite.key;

// Existing publish.html#key=… links still work. Read the token, strip it
// from the address bar, and exchange it for the cookie.
if (invite.key) {
  history.replaceState(null, "", location.pathname + location.search);
  void establish(invite.key).then((ok) => (ok ? signedIn() : inviteFailed()));
} else {
  void run(fetchMe()).then(signedIn).catch(() => {});
}

key.addEventListener("change", () => {
  key.value = readInvite(key.value.split("#")[1] ?? "").key || key.value.trim();
  invitesection.hidden = !key.value;
  void establish(key.value).then((ok) => (ok ? signedIn() : inviteFailed()));
});

const templateForm = $<HTMLFormElement>("template-install");
const templateTar = $<HTMLInputElement>("template-tar");
const templateStatus = $("template-status");
const templateGo = $<HTMLButtonElement>("template-install-go");
const sayTemplate = (...parts: Array<string | Node>): void => {
  templateStatus.hidden = false;
  templateStatus.replaceChildren(...parts);
};
templateForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const file = templateTar.files?.[0];
  if (!file) return;
  templateGo.disabled = true;
  sayTemplate("Installing…");
  try {
    const headers: Record<string, string> = { "Content-Type": "application/x-tar" };
    if (key.value) headers.Authorization = `Bearer ${key.value}`;
    const r = await fetch("/_relay/v1/templates", { method: "POST", headers, body: file });
    const reply = await r.json();
    if (!r.ok) {
      sayTemplate(MESSAGES[reply.error?.code] ?? reply.error?.message ?? "Could not install the template.");
      return;
    }
    templateTar.value = "";
    templatesLoaded = false;
    await loadTemplates();
    sayTemplate(`Installed ${reply.template?.name ?? "template"}.`);
  } catch {
    sayTemplate("Couldn't reach the server. Try again.");
  } finally {
    templateGo.disabled = false;
  }
});

const guest = $<HTMLInputElement>("guest");
const invitestatus = $("invitestatus");
const makeinvite = $<HTMLButtonElement>("makeinvite");
const sayInvite = (...parts: Array<string | Node>) => {
  invitestatus.hidden = false;
  invitestatus.replaceChildren(...parts);
};

$("inviteform").addEventListener("submit", async (e) => {
  e.preventDefault();
  makeinvite.disabled = true;
  sayInvite("Making the invite…");
  try {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (key.value) headers.Authorization = `Bearer ${key.value}`;
    const r = await fetch("/_relay/v1/invites", {
      method: "POST",
      headers,
      body: JSON.stringify({ name: guest.value.trim() }),
    });
    const reply = await r.json();
    if (!r.ok) {
      sayInvite(MESSAGES[reply.error?.code] ?? reply.error?.message ?? "Something went wrong.");
      return;
    }
    const link = inviteLink(location.origin, reply.token);
    const a = Object.assign(document.createElement("a"), { href: link, textContent: "Copy this invite link" });
    a.addEventListener("click", (ev) => {
      ev.preventDefault();
      void navigator.clipboard.writeText(link).then(() => (a.textContent = "Copied"));
    });
    sayInvite("Invite for ", guest.value.trim(), ": ", a);
    guest.value = "";
  } catch {
    sayInvite("Couldn't reach the server. Try again.");
  } finally {
    makeinvite.disabled = false;
  }
});

const copyai = $<HTMLButtonElement>("copyai");
copyai.addEventListener("click", async () => {
  await navigator.clipboard.writeText(assistantPrompt(location.origin, key.value || "YOUR-KEY", site.value));
  copyai.textContent = "Copied";
  setTimeout(() => (copyai.textContent = "Copy instructions for Claude"), 2500);
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
    const headers: Record<string, string> = {};
    if (key.value) headers.Authorization = `Bearer ${key.value}`;
    const r = await fetch(`/_relay/v1/sites/${site.value}/deploy`, {
      method: "POST",
      headers,
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
