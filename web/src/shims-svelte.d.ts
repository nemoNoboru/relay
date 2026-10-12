/**
 * Module shim so plain `tsc --noEmit` can resolve `.svelte` imports.
 *
 * `svelte-check` (run right after tsc in `npm run typecheck`) does the real
 * strict checking of the `.svelte` files themselves, including the prop
 * types this generic shim cannot express. Together they cover the whole
 * project.
 */
declare module "*.svelte" {
  import type { Component } from "svelte";
  const component: Component<Record<string, unknown>>;
  export default component;
}