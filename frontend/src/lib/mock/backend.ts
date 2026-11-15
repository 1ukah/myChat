// Development-only stand-in for the Go backend. It is imported dynamically,
// and only when the Wails runtime is missing (plain `npm run dev`), so none
// of this is on the real app's path.
//
//   ?view=overlay&mock=composer   chat box with a sample log (you can type)
//   ?view=overlay&mock=toast      incoming-message pop-ups, one after another
//   ?view=overlay&mock=suggest    chat box while typing "/w s", with an error
//   ?view=overlay&mock=offline    chat box while Steam is reconnecting
//   ?view=settings&mock=login     logged out; the QR flow runs when started
//   ?view=settings&mock=qr        waiting for the QR scan
//   ?view=settings&mock=error     QR login failed
//   ?view=settings&mock=online    logged in and online
//   ?view=settings&mock=offline   logged in, Steam unreachable
//
// Extra switches: &blur=0  &anim=0  &scale=0.75  &bg=dark|bright|mixed
//
// The chat engine below is a small imitation of internal/chat, good enough to
// try the keys: text, arrows, Tab, /w name, /r, /p, /help, /clear, PgUp/PgDn,
// Enter, Esc. A message containing "fail" fails to send.

import type { Backend, Handlers } from "../backend";
import type { View } from "../store.svelte";
import * as M from "../../../bindings/mychat/models";
import * as C from "../../../bindings/mychat/internal/chat/models";
import * as Cfg from "../../../bindings/mychat/internal/config/models";

interface Friend {
  id: string;
  name: string;
  online: boolean;
  inGame: string;
}

const CS2 = "Counter-Strike 2";

const FRIENDS: Friend[] = [
  { id: "76561198000000001", name: "sant1", online: true, inGame: CS2 },
  { id: "76561198000000002", name: "Mira", online: true, inGame: CS2 },
  { id: "76561198000000003", name: "Дмитрий", online: true, inGame: "" },
  { id: "76561198000000004", name: "samurai_jack_the_unbelievably_long_nickname", online: false, inGame: "" },
  { id: "76561198000000005", name: "サクラ", online: true, inGame: "Dota 2" },
  { id: "76561198000000006", name: "Sandro", online: false, inGame: "" },
  { id: "76561198000000007", name: "أحمد", online: true, inGame: CS2 },
  { id: "76561198000000008", name: "s1mple_fan", online: true, inGame: "" },
];

const INCOMING: [number, string][] = [
  [1, "we need one more for premier, you in?"],
  [0, "gg, that last round was insane"],
  [2, "ты где? заходи в лобби"],
  [3, "this is a deliberately long message to show how a pop-up clamps its text to two lines and ends with an ellipsis instead of growing or pushing anything around"],
  [4, "おつかれ！また明日"],
];

const HELP = [
  "/r [msg] – reply to the last whisper",
  "/w name [msg] – whisper a friend",
  "/p [msg] – message your CS2 party",
  "Tab – cycle recent people · ↑↓ – sent history",
  "PgUp/PgDn – scroll · Esc – close",
];

const WHISPER = /^\/(w|whisper|t|tell|msg|m)$/i;

function hotkeyLabel(h: Cfg.Hotkey): string {
  const key = h.code.replace(/^(Key|Digit)/, "").replace(/^Numpad/, "Num ");
  return (h.ctrl ? "Ctrl + " : "") + (h.alt ? "Alt + " : "") + (h.shift ? "Shift + " : "") + key;
}

/** A matrix that looks like a version-4 QR code (it does not scan). */
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

  let handlers: Handlers | null = null;
  const emit = <K extends keyof Handlers>(name: K, ...args: Parameters<Handlers[K]>): void => {
    const h = handlers?.[name] as ((...a: Parameters<Handlers[K]>) => void) | undefined;
    h?.(...args);
  };

  // ---------------- config ----------------

  let cfg = new Cfg.Config({
    hotkey: new Cfg.Hotkey({ code: "KeyI" }),
    closeAfterSend: false,
    animations: flag("anim", true),
    blur: flag("blur", true),
    showToasts: true,
    toastSeconds: 6,
    bottomOffset: 150,
    scale: scale >= 0.75 && scale <= 1.5 ? scale : 1,
    machineSeed: "mock",
  });
  const configView = (): M.ConfigView =>
    new M.ConfigView({ ...cfg, hotkeyLabel: hotkeyLabel(cfg.hotkey) });

  // ---------------- Steam status and login ----------------

  const loggedIn = view === "overlay" || mock === "online" || mock === "offline";
  let status = new M.Status(
    !loggedIn
      ? { state: "needs_login" }
      : mock === "offline"
        ? { state: "offline", detail: "Connection lost. Trying again in 5 s.", accountName: "sant1_cs", personaName: "sant1" }
        : { state: "online", accountName: "sant1_cs", personaName: "sant1", friendsOnline: 7, partySize: 2 },
  );
  let login = new M.LoginState(
    mock === "qr"
      ? { phase: "waiting", url: "https://s.team/q/1/1234567890", qr: fakeQr(7) }
      : mock === "error"
        ? { phase: "error", error: "Steam did not answer in time. Check your connection and try again." }
        : { phase: "idle" },
  );
  let loginTimers: number[] = [];

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

  // ---------------- chat engine (imitation) ----------------

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

  const friend = (id: string): Friend | undefined => FRIENDS.find((f) => f.id === id);
  const touch = (id: string): void => {
    mru = [id, ...mru.filter((v) => v !== id)].slice(0, 20);
  };

  const composer = (): C.ComposerState => {
    const ch = new C.Channel({ ...channel });
    if (ch.kind === "whisper") ch.online = friend(ch.targetId)?.online ?? false;
    let hint = "";
    if (sugg.length > 0) hint = "Tab or Enter to pick";
    else if (channel.kind === "none" && lastFrom) hint = "/r to reply · /w name · /p party";
    else if (channel.kind === "none") hint = "/w name to whisper · /p party · /help";
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
    new M.ModeInfo({ mode, animations: cfg.animations, blur: cfg.blur });
  const setMode = (m: string): void => {
    mode = m;
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
    mode = "toast";
    toastTimer = window.setTimeout(() => {
      if (mode === "toast") setMode("hidden");
    }, cfg.toastSeconds * 1000);
    emit("toast", entry);
    emit("mode", modeInfo());
  };

  const incoming = (from: Friend, text: string): void => {
    const entry = add({ kind: "in", peerId: from.id, peerName: from.name, text });
    lastFrom = from.id;
    touch(from.id);
    changed();
    toast(entry);
  };

  const setWhisper = (f: Friend): void => {
    channel = new C.Channel({ kind: "whisper", targetId: f.id, targetName: f.name });
  };

  const match = (q: string): { all: Friend[]; exact: Friend[] } => {
    const query = q.toLowerCase();
    const groups: Friend[][] = [[], [], []];
    for (const f of FRIENDS) {
      const n = f.name.toLowerCase();
      if (n === query) groups[0].push(f);
      else if (n.startsWith(query)) groups[1].push(f);
      else if (n.includes(query)) groups[2].push(f);
    }
    const onlineFirst = (g: Friend[]): Friend[] => [...g.filter((f) => f.online), ...g.filter((f) => !f.online)];
    return { all: groups.flatMap(onlineFirst), exact: groups[0] };
  };

  /** Live channel switching: "/r ", "/p " and "/w name " act at once. */
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
    } else if ((cmd === "/p" || cmd === "/party") && arg === "") {
      channel = new C.Channel({ kind: "party", targetName: "Party" });
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
    const entry = add({ kind: "out", peerId: to.id, peerName: to.name, text, status: "pending" });
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

  const sendParty = (text: string): boolean => {
    if (status.partySize === 0) {
      error = "You are not in a CS2 party";
      return false;
    }
    const entry = add({ kind: "party", peerName: `Party (${status.partySize})`, text, status: "pending" });
    chars = [];
    caret = 0;
    window.setTimeout(() => update(entry, "sent"), 600);
    return true;
  };

  /** Returns true when a message was accepted for sending. */
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
    if (cmd === "/p" || cmd === "/party") {
      channel = new C.Channel({ kind: "party", targetName: "Party" });
      if (rest === "") return clear(), false;
      return sendParty(rest);
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

  /** Returns "close", "sent" or "". */
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
    if (channel.kind === "party") return sendParty(t) ? "sent" : "";
    error = "Pick who to message: /w name, /r or /p";
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
    if (action === "close" || (action === "sent" && cfg.closeAfterSend)) closeComposer();
  };

  // ---------------- starting state ----------------

  const seedLog = (): void => {
    const now = Date.now();
    const at = (minutesAgo: number): number => now - minutesAgo * 60_000;
    const [sant1, mira, dmitri, long, sakura, , ahmed] = FRIENDS;
    const lines: Partial<C.Entry>[] = [
      { time: at(34), kind: "system", text: "Connected to Steam as sant1." },
      { time: at(31), kind: "in", peerId: dmitri.id, peerName: dmitri.name, text: "привет, го катку?" },
      { time: at(30), kind: "out", peerId: dmitri.id, peerName: dmitri.name, text: "через 10 минут", status: "sent" },
      { time: at(22), kind: "in", peerId: sakura.id, peerName: sakura.name, text: "こんにちは！今夜プレイする？" },
      { time: at(18), kind: "in", peerId: ahmed.id, peerName: ahmed.name, text: "مرحبا، هل تلعب الليلة؟" },
      { time: at(12), kind: "in", peerId: long.id, peerName: long.name, text: "did you see the new operation trailer? https://store.steampowered.com/news/app/730/view/000000000000000000 the maps look great and I think we should try the co-op missions this weekend" },
      { time: at(9), kind: "party", peerName: "Party (2)", text: "buy armor this round, save next", status: "sent" },
      { time: at(7), kind: "in", peerId: mira.id, peerName: mira.name, text: "rotate B, they stacked A again" },
      { time: at(6), kind: "out", peerId: mira.id, peerName: mira.name, text: "on my way, smoke window", status: "sent" },
      { time: at(4), kind: "out", peerId: dmitri.id, peerName: dmitri.name, text: "ready when you are", status: "failed", detail: "Steam did not answer" },
      { time: at(3), kind: "error", text: "Steam connection lost. Reconnecting…" },
      { time: at(2), kind: "system", text: "Back online." },
      { time: at(1), kind: "in", peerId: sant1.id, peerName: sant1.name, text: "one more after this?" },
    ];
    log = lines.map((l) => new C.Entry({ id: nextId++, ...l }));
    mru = [sant1.id, mira.id, dmitri.id, long.id, sakura.id, ahmed.id];
    lastFrom = sant1.id;
  };

  const scenario = (): void => {
    if (view !== "overlay") return;
    window.addEventListener("keydown", onKey);

    if (mock === "toast") {
      // Alternates a short and a long gap: one pop-up replaces the one on
      // screen (cross-fade), the next arrives after it has gone (slide in).
      let n = 0;
      const next = (): void => {
        const [who, text] = INCOMING[n % INCOMING.length];
        incoming(FRIENDS[who], text);
        window.setTimeout(next, n++ % 2 === 0 ? 3500 : cfg.toastSeconds * 1000 + 3000);
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
      const pending = add({ kind: "out", peerId: FRIENDS[0].id, peerName: FRIENDS[0].name, text: "sure, queue it", status: "pending" });
      window.setTimeout(() => update(pending, "sent"), 1800);
      window.setTimeout(() => incoming(FRIENDS[1], "nice clutch 😄 same team next game?"), 3200);
    }, 300);
  };

  seedLog();

  // ---------------- the Backend ----------------

  const done = (): Promise<void> => Promise.resolve();

  return {
    snapshot: () =>
      Promise.resolve(
        new M.Snapshot({
          mode: modeInfo(),
          log: [...log],
          composer: composer(),
          status,
          config: configView(),
          login,
          layout: new M.Layout({ width: 640, composerHeight: 360, toastHeight: 92 }),
        }),
      ),

    overlayReady: done,

    saveConfig(c) {
      // "F12" stands in for a key another app has already registered.
      if (c.hotkey.code === "F12") {
        return Promise.reject(new Error("that key is already in use by another app"));
      }
      const next = new Cfg.Config({ ...c, hotkey: new Cfg.Hotkey({ ...c.hotkey }), machineSeed: cfg.machineSeed });
      if (!next.hotkey.code) next.hotkey = new Cfg.Hotkey({ code: "KeyI" });
      if (next.toastSeconds < 2 || next.toastSeconds > 30) next.toastSeconds = 6;
      if (next.bottomOffset < 0 || next.bottomOffset > 1000) next.bottomOffset = 150;
      if (next.scale < 0.75 || next.scale > 1.5) next.scale = 1;
      cfg = next;
      const saved = configView();
      emit("config", saved);
      return new Promise((resolve) => window.setTimeout(() => resolve(saved), 60));
    },

    startLogin() {
      stopLogin();
      setLogin({ phase: "waiting" });
      loginTimers = [
        window.setTimeout(() => setLogin({ phase: "waiting", url: "https://s.team/q/1/1234567890", qr: fakeQr(Date.now()) }), 700),
        window.setTimeout(() => {
          setStatus({ state: "connecting", accountName: "sant1_cs" });
          setLogin({ phase: "done", account: "sant1_cs" });
        }, 9000),
        window.setTimeout(() => setStatus({ state: "online", personaName: "sant1", friendsOnline: 7, partySize: 2 }), 10500),
      ];
      return done();
    },

    cancelLogin() {
      stopLogin();
      setLogin({ phase: "idle" });
      return done();
    },

    logout() {
      stopLogin();
      status = new M.Status({ state: "needs_login" });
      emit("status", status);
      setLogin({ phase: "idle" });
      return done();
    },

    previewToast() {
      toast(new C.Entry({ id: -Date.now(), time: Date.now(), kind: "in", peerName: "myChat", text: "This is how new messages look in game." }));
      return done();
    },

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
