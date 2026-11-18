import type { Backend, Handlers } from "../backend";
import type { View } from "../store.svelte";
import * as M from "../../../bindings/mychat/models";
import * as C from "../../../bindings/mychat/internal/chat/models";
import * as Cfg from "../../../bindings/mychat/internal/config/models";
import * as SL from "../../../bindings/mychat/internal/steamlocal/models";
import photoUrl from "../../assets/preview-bg.jpg";
import gifUrl from "../../assets/sample.gif";

const PICS = {
  photo: { url: photoUrl, width: 1280, height: 720, kind: "upload", title: "" },
  tall: { url: photoUrl, width: 720, height: 1280, kind: "upload", title: "" },
  gif: { url: gifUrl, width: 120, height: 68, kind: "gif", title: "Blue gradient" },
  broken: { url: "/media/missing", width: 800, height: 450, kind: "upload", title: "" },
  link: { url: photoUrl, width: 0, height: 0, kind: "link", title: "" },
} as const;

type Pic = keyof typeof PICS;

const mediaOf = (pic: Pic, id: number): C.Media[] => [
  new C.Media({ id: `mock-${pic}-${id}`, ...PICS[pic] }),
];

const LINK = "https://example.com/clutch.png";

const YOUTUBE = "https://www.youtube.com/watch?v=dQw4w9WgXcQ";
const TWEET = "https://x.com/jack/status/20";

const EMBEDS: Record<string, C.Embed> = {
  [YOUTUBE]: new C.Embed({
    url: YOUTUBE,
    site: "youtube",
    siteName: "YouTube",
    author: "Rick Astley",
    title: "Rick Astley - Never Gonna Give You Up (Official Video) (4K Remaster)",
    video: true,
    thumb: new C.Media({ id: "mock-yt", url: photoUrl, width: 320, height: 180, kind: "link" }),
  }),
  [TWEET]: new C.Embed({
    url: TWEET,
    site: "x",
    siteName: "X",
    author: "@jack",
    title: "jack",
    text: "just setting up my twttr",
  }),
};

const embedIn = (text: string): C.Embed | null =>
  Object.entries(EMBEDS).find(([url]) => text.includes(url))?.[1] ?? null;

interface Friend {
  id: string;
  name: string;
  altName?: string;
  online: boolean;
  inGame: string;
  avatar: string;
}

const CS2 = "Counter-Strike 2";

function avatar(name: string, hue: number): string {
  const letter = Array.from(name)[0]?.toUpperCase() ?? "?";
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
    `<rect width="64" height="64" fill="hsl(${hue} 50% 34%)"/>` +
    `<circle cx="48" cy="12" r="24" fill="hsl(${(hue + 40) % 360} 75% 58%)" opacity="0.55"/>` +
    `<path d="M0 64L26 30L44 64z" fill="hsl(${(hue + 320) % 360} 60% 22%)" opacity="0.6"/>` +
    `<text x="32" y="43" font-family="Segoe UI, sans-serif" font-size="30" font-weight="700" text-anchor="middle" fill="#fff">${letter}</text>` +
    `</svg>`;
  return "data:image/svg+xml," + encodeURIComponent(svg);
}

const BROKEN_AVATAR = "data:image/png;base64,AAAA";

const FRIENDS: Friend[] = [
  { id: "76561198000000001", name: "sant1", online: true, inGame: CS2, avatar: avatar("sant1", 265) },
  { id: "76561198000000002", name: "Mira", online: true, inGame: CS2, avatar: avatar("Mira", 150) },
  { id: "76561198000000003", name: "Дмитрий", online: true, inGame: "", avatar: avatar("Дмитрий", 20) },
  { id: "76561198000000004", name: "samurai_jack_the_unbelievably_long_nickname", online: false, inGame: "", avatar: BROKEN_AVATAR },
  { id: "76561198000000005", name: "サクラ", online: true, inGame: "Dota 2", avatar: avatar("サクラ", 330) },
  { id: "76561198000000006", name: "Sandro", online: false, inGame: "", avatar: "" },
  { id: "76561198000000007", name: "أحمد", online: true, inGame: CS2, avatar: avatar("أحمد", 200) },
  { id: "76561198000000008", name: "s1mple_fan", online: true, inGame: "", avatar: avatar("s1mple_fan", 45) },
  { id: "76561198000000009", name: "Stas (work)", altName: "st4s_", online: true, inGame: CS2, avatar: avatar("st4s_", 95) },
];

const OWN_AVATAR = avatar("sant1", 210);

const peer = (f: Friend): Pick<C.Entry, "peerId" | "peerName" | "peerAvatar"> => ({
  peerId: f.id,
  peerName: f.name,
  peerAvatar: f.avatar,
});

const STYLES = ["fade", "slide-up", "slide-down", "slide-left", "slide-right", "typewriter"];
const SPEEDS: Record<string, number> = { fast: 120, normal: 200, slow: 320 };

const LONG =
  "this is a deliberately long message to show how a pop-up clamps its text to two lines and ends with an ellipsis instead of growing or pushing anything around";

const INCOMING: [number, string, Pic?][] = [
  [1, "we need one more for premier, you in?"],
  [1, YOUTUBE],
  [0, `lmao ${TWEET}`],
  [1, "look at this lineup", "photo"],
  [0, "gg, that last round was insane"],
  [0, "", "gif"],
  [2, "ты где? заходи в лобби"],
  [2, "жду"],
  [2, "", "tall"],
  [3, LONG],
  [3, `this one ${LINK}`, "link"],
  [4, "おつかれ！また明日"],
  [4, "", "broken"],
  [0, "one more?"],
];

const STILL: [number, string, Pic?][] = [
  [0, "gg, that last round was insane"],
  [0, "look at this", "photo"],
  [1, "", "broken"],
  [1, "we need one more for premier, you in?"],
  [0, "", "gif"],
];

const HELP = [
  "/r [msg] – reply to the last whisper",
  "/w name [msg] – whisper a friend",
  "Tab – cycle recent people · ↑↓ – sent history",
  "PgUp/PgDn – scroll · Esc – close",
];

const WHISPER = /^\/(w|whisper|t|tell|msg|m)$/i;

function hotkeyLabel(h: Cfg.Hotkey): string {
  const key = h.code.replace(/^(Key|Digit)/, "").replace(/^Numpad/, "Num ");
  return (h.ctrl ? "Ctrl + " : "") + (h.alt ? "Alt + " : "") + (h.shift ? "Shift + " : "") + key;
}

function fakeQr(seed: number): string[] {
  const n = 33;
  const m = Array.from({ length: n }, () => new Array<number>(n).fill(-1));
  const finder = (r: number, c: number): void => {
    for (let y = -1; y <= 7; y++) {
      for (let x = -1; x <= 7; x++) {
        if (r + y < 0 || r + y >= n || c + x < 0 || c + x >= n) continue;
        const edge = y === 0 || y === 6 || x === 0 || x === 6;
        const core = y >= 2 && y <= 4 && x >= 2 && x <= 4;
        const inside = y >= 0 && y <= 6 && x >= 0 && x <= 6;
        m[r + y][c + x] = inside && (edge || core) ? 1 : 0;
      }
    }
  };
  finder(0, 0);
  finder(0, n - 7);
  finder(n - 7, 0);
  for (let y = -2; y <= 2; y++) {
    for (let x = -2; x <= 2; x++) {
      m[26 + y][26 + x] = Math.max(Math.abs(x), Math.abs(y)) === 1 ? 0 : 1;
    }
  }
  for (let i = 8; i < n - 8; i++) m[6][i] = m[i][6] = i % 2 === 0 ? 1 : 0;
  let s = seed >>> 0;
  const bit = (): number => {
    s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
    return s >>> 31;
  };
  return m.map((row) => row.map((v) => (v < 0 ? bit() : v)).join(""));
}

export function createMockBackend(view: View, params: URLSearchParams): Backend {
  const mock = params.get("mock") ?? (view === "overlay" ? "composer" : "online");
  const flag = (name: string, fallback: boolean): boolean =>
    params.has(name) ? params.get(name) !== "0" : fallback;
  const scale = Number(params.get("scale"));
  const max = Math.round(Number(params.get("max")));

  let handlers: Handlers | null = null;
  const emit = <K extends keyof Handlers>(name: K, ...args: Parameters<Handlers[K]>): void => {
    const h = handlers?.[name] as ((...a: Parameters<Handlers[K]>) => void) | undefined;
    h?.(...args);
  };

  let cfg = new Cfg.Config({
    hotkey: new Cfg.Hotkey({ code: "KeyI" }),
    closeAfterSend: false,
    doubleEnterClose: true,
    overlayStyle: params.get("look") === "minimal" ? "minimal" : "legacy",
    showAccountName: false,
    animations: flag("anim", true),
    animationStyle: STYLES.includes(params.get("style") ?? "") ? (params.get("style") as string) : "slide-up",
    animationSpeed: (params.get("speed") ?? "") in SPEEDS ? (params.get("speed") as string) : "normal",
    showAvatars: flag("avatars", true),
    blur: flag("blur", false),
    showToasts: true,
    toastSeconds: 6,
    toastLayout: params.get("layout") === "stack" ? "stack" : "group",
    toastMax: max >= 1 && max <= 5 ? max : view === "overlay" ? 5 : 3,
    messageSound: true,
    showImages: flag("images", true),
    imageLinks: flag("links", false),
    linkPreviews: flag("previews", true),
    bottomOffset: 150,
    scale: scale >= 0.75 && scale <= 1.5 ? scale : 1,
    machineSeed: "mock",
  });
  const configView = (): M.ConfigView =>
    new M.ConfigView({ ...cfg, hotkeyLabel: hotkeyLabel(cfg.hotkey) });

  const EXPIRED = "Steam's saved login for sant1_cs has expired. Sign in to Steam again.";
  const noLogin = (account: string): string =>
    `Steam has no saved login for ${account}. Sign in to Steam with "Remember me" on`;

  const loggedIn = view === "overlay" || mock === "online" || mock === "offline";
  const source = ["qr", "password"].includes(params.get("source") ?? "")
    ? (params.get("source") as string)
    : "steam";
  let status = new M.Status(
    !loggedIn
      ? { state: "needs_login", detail: mock === "expired" ? EXPIRED : "" }
      : mock === "offline"
        ? { state: "offline", detail: "Connection lost. Trying again in 5 s.", accountName: "sant1_cs", source, personaName: "sant1", avatar: OWN_AVATAR }
        : { state: "online", accountName: "sant1_cs", source, personaName: "sant1", avatar: OWN_AVATAR },
  );
  let login = new M.LoginState(
    mock === "expired"
      ? { phase: "idle", error: EXPIRED }
      : mock === "error"
        ? { phase: "error", method: "password", account: "sant1_cs", error: "Wrong account name or password." }
        : mock === "guard"
          ? { phase: "guard", method: "password", account: "sant1_cs", guards: ["device_code", "device_confirmation"] }
          : mock === "email"
            ? { phase: "guard", method: "password", account: "sant1_cs", guards: ["email_code"], emailHint: "gmail.com" }
            : { phase: "idle" },
  );
  let loginTimers: number[] = [];

  const steamInstalled = mock !== "nosteam";
  const accounts: SL.Account[] = steamInstalled
    ? [
        new SL.Account({ accountName: "sant1_cs", personaName: "sant1", steamId: "76561198000000100", mostRecent: true, hasLogin: mock !== "expired" }),
        new SL.Account({ accountName: "expired", personaName: "Old Smurf", steamId: "76561198000000101", hasLogin: true }),
        new SL.Account({ accountName: "mira_alt", personaName: "Mira", steamId: "76561198000000102", hasLogin: false }),
      ]
    : [];
  let watching = false;
  let watchTimer = 0;

  const localLogins = (): M.LocalLogins =>
    new M.LocalLogins({
      steamInstalled,
      accounts: accounts.map((a) => new SL.Account({ ...a })),
      watching,
      error: steamInstalled ? "" : "Steam is not installed on this PC.",
    });
  const emitLocal = (): void => emit("localLogins", localLogins());

  const setStatus = (s: Partial<M.Status>): void => {
    status = new M.Status({ ...status, ...s });
    emit("status", status);
  };
  const setLogin = (l: Partial<M.LoginState>): void => {
    login = new M.LoginState(l);
    emit("login", login);
  };
  const stopLogin = (): void => {
    loginTimers.forEach((t) => clearTimeout(t));
    loginTimers = [];
  };
  const stopWatch = (): void => {
    clearTimeout(watchTimer);
    watching = false;
  };
  const finishLogin = (method: string, account: string): void => {
    const persona = accounts.find((a) => a.accountName === account)?.personaName || account;
    status = new M.Status({ state: "connecting", accountName: account, source: method });
    emit("status", status);
    setLogin({ phase: "done", method, account });
    loginTimers = [
      window.setTimeout(
        () => setStatus({ state: "online", personaName: persona, avatar: avatar(persona, 210) }),
        300,
      ),
    ];
  };
  const watch = (newAccount: boolean): void => {
    stopWatch();
    watching = true;
    emitLocal();
    watchTimer = window.setTimeout(() => {
      watching = false;
      let found = newAccount ? undefined : accounts.find((a) => !a.hasLogin);
      if (!found) {
        found = new SL.Account({ accountName: "new_login", personaName: "Fresh Start", steamId: "76561198000000103" });
        accounts.unshift(found);
      }
      for (const a of accounts) a.mostRecent = a === found;
      found.hasLogin = true;
      emitLocal();
      finishLogin("steam", found.accountName);
    }, 3000);
  };

  let mode = "hidden";
  let log: C.Entry[] = [];
  let nextId = 1;
  let chars: string[] = [];
  let caret = 0;
  let channel = new C.Channel({ kind: "none" });
  let lastFrom = "";
  let mru: string[] = [];
  let history: string[] = [];
  let histPos = 0;
  let sugg: Friend[] = [];
  let selected = 0;
  let error = "";
  let scroll = 0;
  let toastTimer = 0;
  let toasts: C.Entry[] = [];
  let viewer: C.Media | null = null;
  let lastEmptyEnter = 0;
  const still = mock === "toast";

  const friend = (id: string): Friend | undefined => FRIENDS.find((f) => f.id === id);
  const touch = (id: string): void => {
    mru = [id, ...mru.filter((v) => v !== id)].slice(0, 20);
  };

  const composer = (): C.ComposerState => {
    const ch = new C.Channel({ ...channel });
    if (ch.kind === "whisper") ch.online = friend(ch.targetId)?.online ?? false;
    let hint = "";
    if (sugg.length > 0) hint = "Tab or Enter to pick";
    else if (channel.kind === "none" && lastFrom) hint = "/r to reply · /w name";
    else if (channel.kind === "none") hint = "/w name to whisper · /help";
    return new C.ComposerState({
      text: chars.join(""),
      caret,
      channel: ch,
      hint,
      suggestions: sugg.map((f) => new C.Suggestion({ ...f })),
      selected,
      error,
      scroll,
    });
  };
  const changed = (): void => emit("composer", composer());

  const add = (e: Partial<C.Entry>): C.Entry => {
    const entry = new C.Entry({ id: nextId++, time: Date.now(), ...e });
    log.push(entry);
    emit("entry", entry);
    return entry;
  };
  const update = (entry: C.Entry, statusText: string, detail = ""): void => {
    const next = new C.Entry({ ...entry, status: statusText, detail });
    log = log.map((e) => (e.id === next.id ? next : e));
    emit("entryUpdated", next);
  };

  const modeInfo = (): M.ModeInfo =>
    new M.ModeInfo({
      mode,
      animations: cfg.animations,
      animationStyle: cfg.animationStyle,
      animationSpeed: cfg.animationSpeed,
      hideDelayMs: cfg.animations ? SPEEDS[cfg.animationSpeed] : 0,
      blur: cfg.blur,
      showAvatars: cfg.showAvatars,
      overlayStyle: cfg.overlayStyle,
      viewer,
    });
  const setMode = (m: string): void => {
    mode = m;
    if (m !== "toast") viewer = null;
    emit("mode", modeInfo());
  };

  const startToastTimer = (): void => {
    clearTimeout(toastTimer);
    if (still || viewer) return;
    toastTimer = window.setTimeout(() => {
      if (mode === "toast") setMode("hidden");
    }, cfg.toastSeconds * 1000);
  };

  let spots = new Map<string, M.Spot>();
  const openPicture = (m: C.Media): void => {
    if (mode === "hidden" || viewer) return;
    viewer = m;
    clearTimeout(toastTimer);
    emit("mode", modeInfo());
  };
  const replyTo = (id: string): void => {
    const f = friend(id);
    if (!f) return;
    if (mode !== "composer") openComposer();
    if (sugg.length > 0) {
      chars = [];
      caret = 0;
      sugg = [];
      selected = 0;
    }
    error = "";
    setWhisper(f);
    changed();
  };
  const clickSpot = (key: string): void => {
    const s = spots.get(key);
    if (s?.kind === "picture") openPicture(s.media);
    else if (s?.kind === "name") replyTo(s.peerId);
    else if (s?.kind === "link") window.open(s.url, "_blank", "noopener");
  };
  const closePicture = (): void => {
    if (!viewer) return;
    viewer = null;
    if (mode === "toast") startToastTimer();
    emit("mode", modeInfo());
  };

  const resetLine = (): void => {
    chars = [];
    caret = 0;
    sugg = [];
    selected = 0;
    error = "";
    scroll = 0;
    histPos = history.length;
  };

  const openComposer = (): void => {
    if (mode === "composer") return;
    clearTimeout(toastTimer);
    toasts = [];
    resetLine();
    changed();
    setMode("composer");
  };
  const closeComposer = (): void => {
    if (mode !== "composer") return;
    setMode("hidden");
    resetLine();
    changed();
  };

  const toast = (entry: C.Entry): void => {
    if (!cfg.showToasts || mode === "composer") return;
    clearTimeout(toastTimer);
    if (mode === "hidden") toasts = [];
    mode = "toast";
    toasts = [...toasts, entry].slice(-cfg.toastMax);
    startToastTimer();
    emit("toasts", new M.Toasts({ entries: [...toasts] }));
    emit("mode", modeInfo());
  };

  const withPics = flag("pics", true);

  const attach = (entry: C.Entry, pic: Pic): void => {
    replace(new C.Entry({ ...entry, text: entry.text.replace(LINK, "").trim(), media: mediaOf(pic, entry.id) }));
  };

  const attachEmbed = (entry: C.Entry, embed: C.Embed): void => {
    const text = entry.text.trim() === embed.url ? "" : entry.text;
    replace(new C.Entry({ ...entry, text, embed }));
  };

  const replace = (next: C.Entry): void => {
    if (!log.some((e) => e.id === next.id)) return;
    log = log.map((e) => (e.id === next.id ? next : e));
    emit("entryUpdated", next);
    if (toasts.some((e) => e.id === next.id)) {
      toasts = toasts.map((e) => (e.id === next.id ? next : e));
      emit("toasts", new M.Toasts({ entries: [...toasts] }));
    }
  };

  const incoming = (from: Friend, text: string, pic?: Pic): void => {
    const linked = pic === "link";
    if (!withPics) pic = undefined;
    const media = pic && !linked ? mediaOf(pic, nextId) : [];
    const entry = add({ kind: "in", ...peer(from), text: text || (pic ? "" : "(picture)"), media });
    lastFrom = from.id;
    touch(from.id);
    changed();
    toast(entry);
    if (pic && linked && cfg.showImages && cfg.imageLinks) {
      window.setTimeout(() => attach(entry, "link"), 1200);
    }
    const embed = !pic && cfg.linkPreviews ? embedIn(text) : null;
    if (embed) window.setTimeout(() => attachEmbed(entry, embed), 800);
  };

  const setWhisper = (f: Friend): void => {
    channel = new C.Channel({ kind: "whisper", targetId: f.id, targetName: f.name, avatar: f.avatar });
  };

  const match = (q: string): { all: Friend[]; exact: Friend[] } => {
    const query = q.toLowerCase();
    const groups: Friend[][] = [[], [], []];
    for (const f of FRIENDS) {
      const names = [f.name, f.altName ?? ""].filter(Boolean).map((n) => n.toLowerCase());
      if (names.includes(query)) groups[0].push(f);
      else if (names.some((n) => n.startsWith(query))) groups[1].push(f);
      else if (names.some((n) => n.includes(query))) groups[2].push(f);
    }
    const onlineFirst = (g: Friend[]): Friend[] => [...g.filter((f) => f.online), ...g.filter((f) => !f.online)];
    return { all: groups.flatMap(onlineFirst), exact: groups[0] };
  };

  const afterEdit = (): void => {
    histPos = history.length;
    sugg = [];
    const t = chars.join("");
    const sp = t.indexOf(" ");
    if (!t.startsWith("/") || sp < 0) return;
    const cmd = t.slice(0, sp).toLowerCase();
    const arg = t.slice(sp + 1);
    if ((cmd === "/r" || cmd === "/reply") && arg === "") {
      const f = friend(lastFrom);
      if (!f) {
        error = "No one has whispered you yet";
        return;
      }
      setWhisper(f);
      chars = [];
      caret = 0;
    } else if (WHISPER.test(cmd)) {
      const q = arg.trim();
      if (q === "") {
        sugg = mru.map(friend).filter((f): f is Friend => Boolean(f)).slice(0, 6);
        selected = 0;
        return;
      }
      const { all, exact } = match(q);
      if (arg.endsWith(" ") && (exact.length === 1 || all.length === 1)) {
        setWhisper(exact.length === 1 ? exact[0] : all[0]);
        chars = [];
        caret = 0;
        return;
      }
      sugg = all.slice(0, 6);
      if (selected >= sugg.length) selected = 0;
    }
  };

  const insert = (text: string): void => {
    const typed = Array.from(text.replace(/[\r\n\t]+/g, " ")).slice(0, 1000 - chars.length);
    chars.splice(caret, 0, ...typed);
    caret += typed.length;
  };

  const pick = (): void => {
    setWhisper(sugg[Math.min(selected, sugg.length - 1)]);
    chars = [];
    caret = 0;
    sugg = [];
    selected = 0;
  };

  const sendWhisper = (to: Friend, text: string): boolean => {
    const entry = add({ kind: "out", ...peer(to), text, status: "pending" });
    touch(to.id);
    chars = [];
    caret = 0;
    const fails = /\bfail/i.test(text);
    window.setTimeout(() => {
      if (fails) update(entry, "failed", "Steam did not answer");
      else update(entry, "sent");
    }, fails ? 1500 : 600);
    if (!fails && to.online) {
      window.setTimeout(() => incoming(to, "ok 👍"), 2200);
    }
    return true;
  };

  const command = (t: string): boolean => {
    const fields = t.split(/\s+/);
    const cmd = fields[0].toLowerCase();
    const rest = t.slice(fields[0].length).trim();
    const clear = (): void => {
      chars = [];
      caret = 0;
    };
    if (cmd === "/r" || cmd === "/reply") {
      const f = friend(lastFrom);
      if (!f) {
        error = "No one has whispered you yet";
        return false;
      }
      setWhisper(f);
      if (rest === "") return clear(), false;
      return sendWhisper(f, rest);
    }
    if (WHISPER.test(cmd)) {
      if (fields.length < 2) {
        error = "Usage: /w name message";
        return false;
      }
      const { all } = match(fields[1]);
      if (all.length !== 1) {
        error =
          all.length === 0
            ? `No friend named "${fields[1]}"`
            : `More than one friend matches "${fields[1]}"`;
        return false;
      }
      setWhisper(all[0]);
      const msg = fields.slice(2).join(" ");
      if (msg === "") return clear(), false;
      return sendWhisper(all[0], msg);
    }
    if (cmd === "/help" || cmd === "/?") {
      HELP.forEach((text) => add({ kind: "system", text }));
      return clear(), false;
    }
    if (cmd === "/clear") {
      log = [];
      emit("logCleared");
      return clear(), false;
    }
    error = `Unknown command ${cmd} (try /help)`;
    return false;
  };

  const enter = (): string => {
    const t = chars.join("").trim();
    if (sugg.length > 0 && WHISPER.test(t.split(/\s+/)[0])) {
      pick();
      return "";
    }
    if (t === "") return "close";
    history = [...history, t].slice(-50);
    histPos = history.length;
    if (t.startsWith("/")) return command(t) ? "sent" : "";
    const to = friend(channel.targetId);
    if (channel.kind === "whisper" && to) return sendWhisper(to, t) ? "sent" : "";
    error = "Pick who to message: /w name or /r";
    return "";
  };

  const wordLeft = (i: number): number => {
    while (i > 0 && chars[i - 1] === " ") i--;
    while (i > 0 && chars[i - 1] !== " ") i--;
    return i;
  };
  const wordRight = (i: number): number => {
    while (i < chars.length && chars[i] === " ") i++;
    while (i < chars.length && chars[i] !== " ") i++;
    return i;
  };

  const recall = (dir: number): void => {
    if (history.length === 0) return;
    histPos = Math.min(Math.max(histPos + dir, 0), history.length);
    chars = Array.from(history[histPos] ?? "");
    caret = chars.length;
  };

  const onKey = (e: KeyboardEvent): void => {
    if (viewer && e.key === "Escape") {
      e.preventDefault();
      closePicture();
      return;
    }
    if (mode !== "composer") {
      const h = cfg.hotkey;
      if (e.code === h.code && e.ctrlKey === h.ctrl && e.shiftKey === h.shift && e.altKey === h.alt) {
        e.preventDefault();
        openComposer();
      }
      return;
    }
    if (e.ctrlKey && e.code === "KeyV") {
      e.preventDefault();
      navigator.clipboard?.readText().then(
        (text) => {
          insert(text);
          afterEdit();
          changed();
        },
        () => undefined,
      );
      return;
    }
    const printable = e.key.length > 0 && Array.from(e.key).length === 1 && !e.ctrlKey && !e.altKey;
    const known = printable || /^(Enter|Escape|Backspace|Delete|Arrow|Home|End|Page|Tab)/.test(e.key);
    if (!known) return;
    e.preventDefault();
    error = "";

    if (printable) {
      insert(e.key);
      afterEdit();
      changed();
      return;
    }
    let action = "";
    switch (e.key) {
      case "Enter":
        action = enter();
        break;
      case "Escape":
        action = "close";
        break;
      case "Backspace": {
        const start = e.ctrlKey ? wordLeft(caret) : Math.max(caret - 1, 0);
        chars.splice(start, caret - start);
        caret = start;
        afterEdit();
        break;
      }
      case "Delete":
        chars.splice(caret, 1);
        afterEdit();
        break;
      case "ArrowLeft":
        caret = e.ctrlKey ? wordLeft(caret) : Math.max(caret - 1, 0);
        break;
      case "ArrowRight":
        caret = e.ctrlKey ? wordRight(caret) : Math.min(caret + 1, chars.length);
        break;
      case "Home":
        caret = 0;
        break;
      case "End":
        caret = chars.length;
        break;
      case "ArrowUp":
        if (sugg.length > 0) selected = (selected - 1 + sugg.length) % sugg.length;
        else recall(-1);
        break;
      case "ArrowDown":
        if (sugg.length > 0) selected = (selected + 1) % sugg.length;
        else recall(1);
        break;
      case "PageUp":
        scroll = Math.min(scroll + 5, Math.max(log.length - 1, 0));
        break;
      case "PageDown":
        scroll = Math.max(scroll - 5, 0);
        break;
      case "Tab":
        if (sugg.length > 0) {
          pick();
        } else if (mru.length === 0) {
          error = "No recent conversations";
        } else {
          const at = channel.kind === "whisper" ? mru.indexOf(channel.targetId) : -1;
          const next = e.shiftKey ? (at <= 0 ? mru.length - 1 : at - 1) : (at + 1) % mru.length;
          const f = friend(mru[next]);
          if (f) setWhisper(f);
        }
        break;
    }
    changed();
    if (action === "close" && e.key === "Enter" && !cfg.closeAfterSend && cfg.doubleEnterClose) {
      const now = performance.now();
      const second = now - lastEmptyEnter < 400;
      lastEmptyEnter = second ? 0 : now;
      if (!second) return;
    }
    if (action === "close" || (action === "sent" && cfg.closeAfterSend)) closeComposer();
  };

  const seedLog = (): void => {
    const now = Date.now();
    const at = (minutesAgo: number): number => now - minutesAgo * 60_000;
    const [sant1, mira, dmitri, long, sakura, sandro, ahmed] = FRIENDS;
    const lines: (Partial<C.Entry> & { pic?: Pic })[] = [
      { time: at(34), kind: "system", text: "Connected to Steam as sant1." },
      { time: at(31), kind: "in", ...peer(dmitri), text: "привет, го катку?" },
      { time: at(30), kind: "out", ...peer(dmitri), text: "через 10 минут", status: "sent" },
      { time: at(22), kind: "in", ...peer(sakura), text: "こんにちは！今夜プレイする？" },
      { time: at(18), kind: "in", ...peer(ahmed), text: "مرحبا، هل تلعب الليلة؟" },
      { time: at(14), kind: "in", ...peer(sandro), text: "back in 5, grabbing food" },
      { time: at(12), kind: "in", ...peer(long), text: "did you see the new operation trailer? https://store.steampowered.com/news/app/730/view/000000000000000000 the maps look great and I think we should try the co-op missions this weekend" },
      { time: at(10), kind: "in", ...peer(sakura), text: "", pic: "gif" },
      { time: at(9), kind: "in", ...peer(sandro), text: "", pic: "broken" },
      { time: at(7), kind: "in", ...peer(mira), text: "rotate B, they stacked A again" },
      { time: at(7), kind: "in", ...peer(mira), text: "this spot, remember it", pic: "photo" },
      { time: at(6), kind: "out", ...peer(mira), text: "on my way, smoke window", status: "sent" },
      { time: at(5), kind: "out", ...peer(mira), text: "", pic: "tall", status: "sent" },
      { time: at(4), kind: "out", ...peer(dmitri), text: "ready when you are", status: "failed", detail: "Steam did not answer" },
      { time: at(3), kind: "error", text: "Steam connection lost. Reconnecting…" },
      { time: at(2), kind: "system", text: "Back online." },
      { time: at(2), kind: "in", ...peer(sant1), text: `lmao ${TWEET}` },
      { time: at(1), kind: "in", ...peer(mira), text: `this is how they hold B ${YOUTUBE}` },
      { time: at(1), kind: "in", ...peer(sant1), text: "one more after this?" },
    ];
    log = lines
      .filter((l) => withPics || !l.pic)
      .map(({ pic, ...l }) => {
        const id = nextId++;
        const embed = cfg.linkPreviews ? embedIn(l.text ?? "") : null;
        return new C.Entry({ id, ...l, media: pic ? mediaOf(pic, id) : [], embed });
      });
    mru = [sant1.id, mira.id, dmitri.id, long.id, sakura.id, ahmed.id];
    lastFrom = sant1.id;
  };

  const scenario = (): void => {
    if (view !== "overlay") {
      if (mock === "watching") watch(true);
      return;
    }
    window.addEventListener("keydown", onKey);

    if (mock === "toast") {
      const long = params.has("long") && params.get("long") !== "0";
      window.setTimeout(() => {
        STILL.forEach(([who, text, pic], i) =>
          incoming(FRIENDS[who], long && i === 3 ? LONG : text, pic),
        );
      }, 300);
      return;
    }

    if (mock === "toaststream") {
      let n = 0;
      const next = (): void => {
        const [who, text, pic] = INCOMING[n++ % INCOMING.length];
        incoming(FRIENDS[who], text, pic);
        window.setTimeout(next, 2000);
      };
      window.setTimeout(next, 400);
      return;
    }

    setWhisper(FRIENDS[0]);
    window.setTimeout(() => {
      openComposer();
      if (mock === "suggest") {
        channel = new C.Channel({ kind: "none" });
        insert("/w s");
        afterEdit();
        selected = 1;
        error = 'More than one friend matches "s"';
        changed();
        return;
      }
      const pending = add({ kind: "out", ...peer(FRIENDS[0]), text: "sure, queue it", status: "pending" });
      window.setTimeout(() => update(pending, "sent"), 1800);
      window.setTimeout(() => incoming(FRIENDS[1], "nice clutch 😄 same team next game?"), 3200);
    }, 300);
  };

  seedLog();

  const done = (): Promise<void> => Promise.resolve();

  return {
    snapshot: () =>
      Promise.resolve(
        new M.Snapshot({
          mode: modeInfo(),
          log: [...log],
          toasts: new M.Toasts({ entries: [...toasts] }),
          composer: composer(),
          status,
          config: configView(),
          login,
          local: localLogins(),
          layout: new M.Layout({ width: 640, composerHeight: 360, toastHeight: 360 }),
        }),
      ),

    overlayReady: done,

    saveConfig(c) {
      if (c.hotkey.code === "F12") {
        return Promise.reject(new Error("that key is already in use by another app"));
      }
      const next = new Cfg.Config({ ...c, hotkey: new Cfg.Hotkey({ ...c.hotkey }), machineSeed: cfg.machineSeed });
      if (!next.hotkey.code) next.hotkey = new Cfg.Hotkey({ code: "KeyI" });
      if (next.toastSeconds < 2 || next.toastSeconds > 30) next.toastSeconds = 6;
      if (next.bottomOffset < 0 || next.bottomOffset > 1000) next.bottomOffset = 150;
      if (next.scale < 0.75 || next.scale > 1.5) next.scale = 1;
      if (next.toastLayout !== "stack") next.toastLayout = "group";
      if (!(next.toastMax >= 1 && next.toastMax <= 5)) next.toastMax = 3;
      if (next.overlayStyle !== "minimal") next.overlayStyle = "legacy";
      if (!STYLES.includes(next.animationStyle)) next.animationStyle = "slide-up";
      if (!(next.animationSpeed in SPEEDS)) next.animationSpeed = "normal";
      cfg = next;
      const saved = configView();
      emit("config", saved);
      emit("mode", modeInfo());
      return new Promise((resolve) => window.setTimeout(() => resolve(saved), 60));
    },

    startLogin() {
      stopLogin();
      stopWatch();
      emitLocal();
      setLogin({ phase: "waiting", method: "qr" });
      loginTimers = [
        window.setTimeout(
          () => setLogin({ phase: "waiting", method: "qr", url: "https://s.team/q/1/1234567890", qr: fakeQr(Date.now()) }),
          700,
        ),
      ];
      return done();
    },

    localLogins: () => Promise.resolve(localLogins()),

    useSteamAccount(account) {
      stopLogin();
      stopWatch();
      const name = account.trim();
      return new Promise((resolve, reject) => {
        window.setTimeout(() => {
          const found = name ? accounts.find((a) => a.accountName === name) : accounts.find((a) => a.hasLogin);
          emitLocal();
          if (!found || !found.hasLogin || found.accountName === "expired") {
            const message = noLogin(name || "this PC");
            setLogin({ phase: "error", method: "steam", account: name, error: message });
            reject(new Error(message));
            return;
          }
          finishLogin("steam", found.accountName);
          resolve();
        }, 400);
      });
    },

    openSteamLogin() {
      console.info("mock: OpenSteamLogin (the Steam app would open now)");
      stopLogin();
      watch(false);
      return done();
    },

    loginWithPassword(account, password) {
      stopLogin();
      stopWatch();
      emitLocal();
      const name = account.trim();
      if (!name || !password) {
        setLogin({ phase: "error", method: "password", account: name, error: "Enter your account name and password." });
        return done();
      }
      const base = { method: "password", account: name };
      const outcome: Partial<M.LoginState> | null =
        password === "wrong"
          ? { ...base, phase: "error", error: "Wrong account name or password." }
          : password === "guard"
            ? { ...base, phase: "guard", guards: ["device_code", "device_confirmation"] }
            : password === "email"
              ? { ...base, phase: "guard", guards: ["email_code"], emailHint: "gmail.com" }
              : null;
      setLogin({ ...base, phase: "submitting" });
      loginTimers = [
        window.setTimeout(() => (outcome ? setLogin(outcome) : finishLogin("password", name)), 900),
      ];
      return done();
    },

    submitGuardCode(code) {
      if (login.phase !== "guard") return done();
      const step = { ...login };
      setLogin({ ...step, checking: true, codeError: "" });
      loginTimers = [
        window.setTimeout(() => {
          if (code.trim().toUpperCase() === "ABCDE") finishLogin("password", step.account);
          else setLogin({ ...step, checking: false, codeError: "Wrong Steam Guard code." });
        }, 700),
      ];
      return done();
    },

    cancelLogin() {
      stopLogin();
      stopWatch();
      setLogin({ phase: "idle" });
      emitLocal();
      return done();
    },

    logout() {
      stopLogin();
      stopWatch();
      status = new M.Status({ state: "needs_login" });
      emit("status", status);
      setLogin({ phase: "idle" });
      emitLocal();
      return done();
    },

    previewToast() {
      toast(new C.Entry({ id: -Date.now(), time: Date.now(), kind: "in", peerName: "Preview", text: "This is how new messages look in game." }));
      return done();
    },

    setSpots(list) {
      spots = new Map(list.map((s) => [s.key, s]));
      return done();
    },
    clickSpot: (key) => (clickSpot(key), done()),
    closePicture: () => (closePicture(), done()),
    overlayResized: () => done(),

    hideSettings: () => (console.info("mock: HideSettings"), done()),
    minimiseSettings: () => (console.info("mock: MinimiseSettings"), done()),
    quit: () => (console.info("mock: Quit"), done()),

    subscribe(h) {
      handlers = h;
      scenario();
      return () => {
        handlers = null;
        window.removeEventListener("keydown", onKey);
      };
    },
  };
}
