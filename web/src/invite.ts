/**
 * Read an invitation token from a URL fragment, a full invite link, or a
 * pasted key. The fragment is never sent to the server; this is the only
 * place it is read.
 */
export const readToken = (raw: string): string => {
  const s = raw.trim();
  if (s === "") return "";
  if (s.startsWith("key=") || s.startsWith("token=")) {
    return new URLSearchParams(s).get("key") ?? new URLSearchParams(s).get("token") ?? "";
  }
  try {
    const u = new URL(s, location.origin);
    if (u.hash !== "") {
      const p = new URLSearchParams(u.hash.replace(/^#/, ""));
      return p.get("key") ?? p.get("token") ?? "";
    }
    if (u.pathname === "/" || u.pathname === "") return "";
  } catch {
    return s;
  }
  return s;
};
