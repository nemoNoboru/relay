/**
 * Fixed fixture data for the Relay admin mock page (admin.html).
 *
 * The mock is a pure function of `AdminData`: fixture in, static HTML out.
 * Plain TS, no wire schema — this is display data, not an API response, so
 * the relay-ui.md schema rules (which cover API responses) do not apply.
 *
 * Key material is never real: fingerprints are truncated (style of
 * node-ui.html) and the key file's private half is masked — the client
 * never models or stores key material.
 */

export interface NodeInfo {
  name: string;
  /** Mono sub-line under the name, e.g. "relay node v0.1.0 · raspberry pi 5". */
  version: string;
  /** The visibility word shown in the vis box, e.g. "Private". */
  visibility: string;
  /** Note under the visibility word, e.g. "behind NAT". */
  visibilityNote: string;
  sessionsUsed: number;
  sessionsMax: number;
  storage: string;
  friendsOnline: number;
  friendsTotal: number;
  uptime: string;
}

export interface PublicFriend {
  name: string;
  /** Mono note under the name, e.g. "public · 312 friends". */
  note: string;
  latencyMs: number;
  selected: boolean;
}

export interface Friend {
  name: string;
  /** Mono role under the name, e.g. "your public friend". */
  role: string;
  /** Path prefix the friend owns, e.g. "/raspberry/". */
  prefix: string;
  online: boolean;
}

export interface IssuedKey {
  name: string;
  /** Truncated fingerprint, e.g. "9f2c…c41a". */
  fingerprint: string;
  created: string;
  status: "claimed" | "open";
}

export interface ClaimedName {
  name: string;
  /** Full relay address, e.g. "relay.cloudpunk.org/raspberry/bookshelf/". */
  address: string;
  /** Truncated fingerprint of the key that owns the name. */
  owner: string;
  size: string;
}

export interface Rule {
  tag: "Allow" | "Review" | "Deny";
  text: string;
}

/** A relay key file as displayed by the mock. The private half is masked. */
export interface KeyFile {
  relay_key: string;
  name: string;
  node: string;
  public_key: string;
  private_key: string;
}

export interface AdminData {
  node: NodeInfo;
  publicFriends: PublicFriend[];
  /** Address the node would be reachable at through the selected friend. */
  resultAddress: string;
  friends: Friend[];
  /** The just-created key file (display-only), or null when none was made. */
  keyFile: KeyFile | null;
  keys: IssuedKey[];
  names: ClaimedName[];
  rules: Rule[];
}

export const ADMIN_FIXTURE: AdminData = {
  node: {
    name: "home-pi",
    version: "relay node v0.1.0 · raspberry pi 5",
    visibility: "Private",
    visibilityNote: "behind NAT",
    sessionsUsed: 12,
    sessionsMax: 64,
    storage: "38.2 MB",
    friendsOnline: 2,
    friendsTotal: 3,
    uptime: "41 d",
  },
  publicFriends: [
    { name: "relay.cloudpunk.org", note: "public · 312 friends", latencyMs: 41, selected: true },
    { name: "relay.openrelay.net", note: "public · 87 friends", latencyMs: 96, selected: false },
  ],
  resultAddress: "relay.cloudpunk.org/raspberry/",
  friends: [
    { name: "relay.cloudpunk.org", role: "your public friend", prefix: "/raspberry/", online: true },
    { name: "bookshelf", role: "on your network", prefix: "/bookshelf/", online: true },
    { name: "workshop", role: "on your network", prefix: "/workshop/", online: false },
  ],
  keyFile: {
    relay_key: "1",
    name: "zine",
    node: "home-pi",
    public_key: "rly_pk_9f2c…c41a",
    private_key: "•••••••• (masked — shown once)",
  },
  keys: [
    { name: "bookshelf", fingerprint: "9f2c…c41a", created: "14 Mar 2026", status: "claimed" },
    { name: "zine", fingerprint: "b7e1…08fd", created: "2 Mar 2026", status: "open" },
  ],
  names: [
    {
      name: "bookshelf",
      address: "relay.cloudpunk.org/raspberry/bookshelf/",
      owner: "9f2c…c41a",
      size: "1.2 MB",
    },
  ],
  rules: [
    { tag: "Allow", text: "plain text, static HTML, small pages" },
    { tag: "Review", text: "anything with links or scripts" },
    { tag: "Deny", text: "personal data, malware, spam" },
  ],
};

/** The same node with nothing connected, keyed, or claimed. */
export const EMPTY_ADMIN: AdminData = {
  node: {
    name: "home-pi",
    version: "relay node v0.1.0 · raspberry pi 5",
    visibility: "Private",
    visibilityNote: "behind NAT",
    sessionsUsed: 0,
    sessionsMax: 64,
    storage: "0 B",
    friendsOnline: 0,
    friendsTotal: 0,
    uptime: "0 d",
  },
  publicFriends: [],
  resultAddress: "",
  friends: [],
  keyFile: null,
  keys: [],
  names: [],
  rules: [
    { tag: "Allow", text: "plain text, static HTML, small pages" },
    { tag: "Review", text: "anything with links or scripts" },
    { tag: "Deny", text: "personal data, malware, spam" },
  ],
};
