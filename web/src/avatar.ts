/** Small avatar helpers shared by the chrome, cards, strip, and comments. */

const PALETTE = ["#ce1f2b", "#dfa32a", "#141210", "#a5141f", "#3a3328"] as const;

/** A stable background colour for a member id, from Relay's palette. */
export const avatarColor = (id: string): string => {
  let h = 0;
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) >>> 0;
  return PALETTE[h % PALETTE.length] ?? "#ce1f2b";
};

/** The first letter of a display name, uppercased. */
export const avatarInitial = (name: string): string => (name.trim()[0] ?? "?").toUpperCase();

/** An inline `style` string: background + a readable text colour. */
export const avatarStyle = (id: string): string => {
  const bg = avatarColor(id);
  const fg = bg === "#dfa32a" ? "var(--ink)" : "var(--paper)";
  return `background:${bg};color:${fg}`;
};
