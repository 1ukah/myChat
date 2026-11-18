import { hasRuntime, wailsBackend, type Backend } from "./backend";
import * as M from "../../bindings/mychat/models";
import * as C from "../../bindings/mychat/internal/chat/models";
import * as Cfg from "../../bindings/mychat/internal/config/models";

export type View = "overlay" | "settings";

const MAX_LOG = 300;

interface AppState {
  ready: boolean;
  mock: boolean;
  mode: M.ModeInfo;
  log: C.Entry[];
  composer: C.ComposerState;
  status: M.Status;
  config: M.ConfigView;
  login: M.LoginState;
  local: M.LocalLogins;
  layout: M.Layout;
  toasts: C.Entry[];
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

let lastAdded = { id: 0, at: 0 };

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
  if (last && e.id <= last.id && replaceEntry(e)) return;
  app.log.push(e);
  if (app.log.length > MAX_LOG) app.log.splice(0, app.log.length - MAX_LOG);
  lastAdded = { id: e.id, at: performance.now() };
}

let confirmed: M.ConfigView = app.config;
let savesInFlight = 0;

function toConfig(v: M.ConfigView): Cfg.Config {
  const { hotkeyLabel: _derived, ...config } = v;
  return new Cfg.Config({ ...config, hotkey: new Cfg.Hotkey({ ...v.hotkey }) });
}

export function errorText(e: unknown): string {
  const s = (e instanceof Error ? e.message : String(e ?? "")).trim();
  return s ? s[0].toUpperCase() + s.slice(1) : "Something went wrong";
}

export async function saveConfig(patch: Partial<Cfg.Config>): Promise<string> {
  const next = new M.ConfigView({ ...app.config, ...patch });
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

function fire(f: () => Promise<void>): void {
  f().catch((e: unknown) => console.error("myChat:", e));
}

export const overlayReady = (): void => fire(() => backend.overlayReady());
export const startLogin = (): void => fire(() => backend.startLogin());
export const cancelLogin = (): void => fire(() => backend.cancelLogin());
export const loginWithPassword = (account: string, password: string): void =>
  fire(() => backend.loginWithPassword(account, password));
export const submitGuardCode = (code: string): void => fire(() => backend.submitGuardCode(code));

export const refreshLocalLogins = (): void =>
  fire(async () => {
    app.local = await backend.localLogins();
  });

export async function useSteamAccount(account: string): Promise<string> {
  try {
    await backend.useSteamAccount(account);
    return "";
  } catch (e) {
    return errorText(e);
  }
}

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
export const overlayResized = (): void => fire(() => backend.overlayResized());
export const hideSettings = (): void => fire(() => backend.hideSettings());
export const minimiseSettings = (): void => fire(() => backend.minimiseSettings());
export const quit = (): void => fire(() => backend.quit());

let started = false;

export async function start(view: View): Promise<void> {
  if (started) return;
  started = true;

  if (!hasRuntime()) {
    const mock = await import("./mock/backend");
    backend = mock.createMockBackend(view, new URLSearchParams(location.search));
    app.mock = true;
  }

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
