// The single reactive store. It wraps Snapshot() and every Go → UI event, and
// exposes the few actions the UI can take. Components read `app` and never
// talk to the backend directly.

import { hasRuntime, wailsBackend, type Backend } from "./backend";
import * as M from "../../bindings/mychat/models";
import * as C from "../../bindings/mychat/internal/chat/models";
import * as Cfg from "../../bindings/mychat/internal/config/models";

export type View = "overlay" | "settings";

/** Same cap as the Go engine keeps. */
const MAX_LOG = 300;

interface AppState {
  /** The snapshot has been applied; views may render. */
  ready: boolean;
  /** Running without the Wails runtime, on sample data. */
  mock: boolean;
  mode: M.ModeInfo;
  log: C.Entry[];
  composer: C.ComposerState;
  status: M.Status;
  config: M.ConfigView;
  login: M.LoginState;
  /** Accounts the Steam app saved on this PC, and the sign-in watch. */
  local: M.LocalLogins;
  layout: M.Layout;
  /** The pop-up messages shown in toast mode, oldest first. */
  toasts: C.Entry[];
  /** The clickable spot under the mouse pointer (`Spot.key`), or "". */
  hover: string;
}

export const app: AppState = $state({
  ready: false,
  mock: false,
  mode: new M.ModeInfo({ mode: "hidden", animationStyle: "slide-up", animationSpeed: "normal" }),
  log: [],
  composer: new C.ComposerState({ channel: new C.Channel({ kind: "none" }) }),
  status: new M.Status({ state: "offline" }),
  config: new M.ConfigView({
    scale: 1,
    animations: true,
    animationStyle: "slide-up",
    animationSpeed: "normal",
    showAvatars: true,
    toastLayout: "group",
    toastMax: 3,
  }),
  login: new M.LoginState({ phase: "idle" }),
  local: new M.LocalLogins({ steamInstalled: true }),
  layout: new M.Layout({ width: 640, composerHeight: 360, toastHeight: 360 }),
  toasts: [],
  hover: "",
});

let backend: Backend = wailsBackend;

// ---- log ----

let lastAdded = { id: 0, at: 0 };

/** True for a line that arrived live a moment ago (it gets the slide-in). */
export function isFresh(id: number): boolean {
  return id === lastAdded.id && performance.now() - lastAdded.at < 400;
}

function replaceEntry(e: C.Entry): boolean {
  for (let i = app.log.length - 1; i >= 0; i--) {
    if (app.log[i].id === e.id) {
      app.log[i] = e;
      return true;
    }
  }
  return false;
}

function addEntry(e: C.Entry): void {
  const last = app.log[app.log.length - 1];
  // An event that raced with the snapshot may already be in the log.
  if (last && e.id <= last.id && replaceEntry(e)) return;
  app.log.push(e);
  if (app.log.length > MAX_LOG) app.log.splice(0, app.log.length - MAX_LOG);
  lastAdded = { id: e.id, at: performance.now() };
}

// ---- config ----

/** Last config confirmed by Go; `app.config` may be ahead of it (optimistic). */
let confirmed: M.ConfigView = app.config;
let savesInFlight = 0;

/** The config Go expects: every field of the view except the derived label. */
function toConfig(v: M.ConfigView): Cfg.Config {
  const { hotkeyLabel: _derived, ...config } = v;
  return new Cfg.Config({ ...config, hotkey: new Cfg.Hotkey({ ...v.hotkey }) });
}

export function errorText(e: unknown): string {
  const s = (e instanceof Error ? e.message : String(e ?? "")).trim();
  return s ? s[0].toUpperCase() + s.slice(1) : "Something went wrong";
}

/**
 * Saves a settings change at once. The UI updates optimistically and falls
 * back to the last confirmed values if Go rejects the change.
 * Returns "" on success, or the message to show next to the control.
 */
export async function saveConfig(patch: Partial<Cfg.Config>): Promise<string> {
  const next = new M.ConfigView({ ...app.config, ...patch });
  // The hotkey label is derived by Go, so a new hotkey shows only once saved.
  if (!patch.hotkey) app.config = next;
  savesInFlight++;
  try {
    confirmed = await backend.saveConfig(toConfig(next));
    return "";
  } catch (e) {
    return errorText(e);
  } finally {
    if (--savesInFlight === 0) app.config = confirmed;
  }
}

// ---- actions ----

function fire(f: () => Promise<void>): void {
  f().catch((e: unknown) => console.error("myChat:", e));
}

export const overlayReady = (): void => fire(() => backend.overlayReady());
export const startLogin = (): void => fire(() => backend.startLogin());
export const cancelLogin = (): void => fire(() => backend.cancelLogin());
/** The password is handed to Go and not kept here. */
export const loginWithPassword = (account: string, password: string): void =>
  fire(() => backend.loginWithPassword(account, password));
export const submitGuardCode = (code: string): void => fire(() => backend.submitGuardCode(code));

/** Re-reads Steam's saved accounts (they change when the user signs in to Steam). */
export const refreshLocalLogins = (): void =>
  fire(async () => {
    app.local = await backend.localLogins();
  });

/**
 * Logs in with the login the Steam app saved for this account.
 * Returns "" on success, or the message to show on the account card.
 */
export async function useSteamAccount(account: string): Promise<string> {
  try {
    await backend.useSteamAccount(account);
    return "";
  } catch (e) {
    return errorText(e);
  }
}

/**
 * Opens the Steam app so the user can sign in there; Go then watches for
 * the new login and uses it. Returns "" or the message to show.
 */
export async function openSteamLogin(): Promise<string> {
  try {
    await backend.openSteamLogin();
    return "";
  } catch (e) {
    return errorText(e);
  }
}
export const logout = (): void => fire(() => backend.logout());
export const previewToast = (): void => fire(() => backend.previewToast());
export const setSpots = (spots: M.Spot[]): void => fire(() => backend.setSpots(spots));
export const clickSpot = (key: string): void => fire(() => backend.clickSpot(key));
export const closePicture = (): void => fire(() => backend.closePicture());
export const hideSettings = (): void => fire(() => backend.hideSettings());
export const minimiseSettings = (): void => fire(() => backend.minimiseSettings());
export const quit = (): void => fire(() => backend.quit());

// ---- start ----

let started = false;

/** Connects to the backend (or the mock), loads the snapshot, starts events. */
export async function start(view: View): Promise<void> {
  if (started) return;
  started = true;

  if (!hasRuntime()) {
    // Plain browser: sample data instead of Go. Kept out of the real bundle path.
    const mock = await import("./mock/backend");
    backend = mock.createMockBackend(view, new URLSearchParams(location.search));
    app.mock = true;
  }

  // Events can arrive before the snapshot does; hold them and replay after.
  let held: (() => void)[] | null = [];
  const on =
    <A extends unknown[]>(fn: (...a: A) => void) =>
    (...a: A): void => {
      if (held) held.push(() => fn(...a));
      else fn(...a);
    };

  backend.subscribe({
    mode: on((m) => {
      app.mode = m;
    }),
    composer: on((c) => {
      app.composer = c;
    }),
    entry: on(addEntry),
    entryUpdated: on((e) => {
      replaceEntry(e);
    }),
    logCleared: on(() => {
      app.log = [];
    }),
    toasts: on((t) => {
      app.toasts = t.entries ?? [];
    }),
    status: on((s) => {
      app.status = s;
    }),
    config: on((c) => {
      confirmed = c;
      if (savesInFlight === 0) app.config = c;
    }),
    login: on((l) => {
      app.login = l;
    }),
    localLogins: on((l) => {
      app.local = l;
    }),
    hover: on((key) => {
      app.hover = key;
    }),
  });

  try {
    const snap = await backend.snapshot();
    app.mode = snap.mode;
    app.toasts = snap.toasts?.entries ?? [];
    app.log = (snap.log ?? []).slice(-MAX_LOG);
    app.composer = snap.composer;
    app.status = snap.status;
    app.config = confirmed = snap.config;
    app.login = snap.login;
    app.local = snap.local;
    if (snap.layout.width > 0) app.layout = snap.layout;
  } catch (e) {
    console.error("myChat: snapshot failed:", e);
  }

  const queued = held;
  held = null;
  for (const f of queued) f();
  app.ready = true;
}
