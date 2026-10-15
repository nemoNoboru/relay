// Copy the canonical relay-deploy skill into web/public so the built owner UI
// (web/dist) serves it at /_relay/skills/relay-deploy/SKILL.md.
//
// The skill's source of truth is .pi/skills/relay-deploy/SKILL.md at the repo
// root; this keeps the hosted copy byte-for-byte identical so agents can point
// at the URL the node serves.
import { copyFile, mkdir } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const src = resolve(here, "..", "..", ".pi", "skills", "relay-deploy", "SKILL.md");
const dest = resolve(here, "..", "public", "skills", "relay-deploy", "SKILL.md");

await mkdir(dirname(dest), { recursive: true });
await copyFile(src, dest);
console.log("copied relay-deploy skill to web/public/skills/relay-deploy/SKILL.md");
