/**
 * Relay admin mock page — this page's only wiring point.
 *
 * Static mock: no live.ts, no poll stream, no API client — the page renders
 * a fixed fixture and never touches the network (the smoke test asserts the
 * bundle never calls fetch). When the mock is promoted to a live page, this
 * file becomes its wiring point.
 */
import { mount } from "svelte";

import { ADMIN_FIXTURE } from "../admin/fixture";
import AdminApp from "../ui/Admin.svelte";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("admin.html is missing #app");
}

mount(AdminApp, { target: root, props: { data: ADMIN_FIXTURE } });
