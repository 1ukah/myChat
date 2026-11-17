// The one seam between the UI and the outside world. The real implementation
// talks to Go through the generated Wails bindings; the mock implementation
// (src/lib/mock, loaded only when the runtime is missing) fakes the same API.

import { Events } from "@wailsio/runtime";
import { ChatService } from "../../bindings/mychat";
import * as Ev from "../../bindings/events";
import type * as M from "../../bindings/mychat/models";
import type * as C from "../../bindings/mychat/internal/chat/models";
import type * as Cfg from "../../bindings/mychat/internal/config/models";

/** Callbacks for every event Go sends to the UI. */
export interface Handlers {
  mode(m: M.ModeInfo): void;
  composer(c: C.ComposerState): void;
  entry(e: C.Entry): void;
  entryUpdated(e: C.Entry): void;
  logCleared(): void;
  /** The pop-up messages, oldest first; replaces the whole list. */
  toasts(t: M.Toasts): void;
  status(s: M.Status): void;
  config(c: M.ConfigView): void;
  login(l: M.LoginState): void;
  localLogins(l: M.LocalLogins): void;
  /** The spot (`Spot.key`) under the mouse pointer, or "". */
  hover(key: string): void;
}

export interface Backend {
  snapshot(): Promise<M.Snapshot>;
  overlayReady(): Promise<void>;
  saveConfig(c: Cfg.Config): Promise<M.ConfigView>;
  /** Accounts the Steam app saved on this PC (also pushed as an event). */
  localLogins(): Promise<M.LocalLogins>;
  /** Rejects with a message when Steam has no usable login for the account. */
  useSteamAccount(account: string): Promise<void>;
  /** Opens the Steam app; Go then watches for a new saved login. */
  openSteamLogin(): Promise<void>;
  /** Starts a QR login. */
  startLogin(): Promise<void>;
  /** Starts an account name + password login; progress arrives as login events. */
  loginWithPassword(account: string, password: string): Promise<void>;
  /** Sends the Steam Guard code during the "guard" phase. */
  submitGuardCode(code: string): Promise<void>;
  /** Stops any login in progress and the Steam sign-in watch. */
  cancelLogin(): Promise<void>;
  logout(): Promise<void>;
  previewToast(): Promise<void>;
  /** Overlay only: the clickable pictures and names on screen (CSS px of the 640×360 area). */
  setSpots(spots: M.Spot[]): Promise<void>;
  /** Acts on a click on a spot (by `Spot.key`): opens the picture, or whispers to the person. */
  clickSpot(key: string): Promise<void>;
  /** Closes the picture open at full size. */
  closePicture(): Promise<void>;
  hideSettings(): Promise<void>;
  minimiseSettings(): Promise<void>;
  quit(): Promise<void>;
  /** Starts delivering events. Returns a function that stops them. */
  subscribe(h: Handlers): () => void;
}

/** True inside a Wails webview, false in a plain browser (`npm run dev`). */
export function hasRuntime(): boolean {
  const w = window as unknown as {
    chrome?: { webview?: { postMessage?: unknown } };
    webkit?: { messageHandlers?: { external?: unknown } };
  };
  return Boolean(w.chrome?.webview?.postMessage || w.webkit?.messageHandlers?.external);
}

export const wailsBackend: Backend = {
  snapshot: () => ChatService.Snapshot(),
  overlayReady: () => ChatService.OverlayReady(),
  saveConfig: (c) => ChatService.SaveConfig(c),
  localLogins: () => ChatService.LocalLogins(),
  useSteamAccount: (account) => ChatService.UseSteamAccount(account),
  openSteamLogin: () => ChatService.OpenSteamLogin(),
  startLogin: () => ChatService.StartLogin(),
  loginWithPassword: (account, password) => ChatService.LoginWithPassword(account, password),
  submitGuardCode: (code) => ChatService.SubmitGuardCode(code),
  cancelLogin: () => ChatService.CancelLogin(),
  logout: () => ChatService.Logout(),
  previewToast: () => ChatService.PreviewToast(),
  setSpots: (spots) => ChatService.SetSpots(spots),
  clickSpot: (key) => ChatService.ClickSpot(key),
  closePicture: () => ChatService.ClosePicture(),
  hideSettings: () => ChatService.HideSettings(),
  minimiseSettings: () => ChatService.MinimiseSettings(),
  quit: () => ChatService.Quit(),

  subscribe(h) {
    const off = [
      Events.On(Ev.MychatMode, (ev) => h.mode(ev.data)),
      Events.On(Ev.MychatComposer, (ev) => h.composer(ev.data)),
      Events.On(Ev.MychatEntry, (ev) => h.entry(ev.data)),
      Events.On(Ev.MychatEntryUpdated, (ev) => h.entryUpdated(ev.data)),
      Events.On(Ev.MychatLogCleared, () => h.logCleared()),
      Events.On(Ev.MychatToasts, (ev) => h.toasts(ev.data)),
      Events.On(Ev.MychatStatus, (ev) => h.status(ev.data)),
      Events.On(Ev.MychatConfig, (ev) => h.config(ev.data)),
      Events.On(Ev.MychatLogin, (ev) => h.login(ev.data)),
      Events.On(Ev.MychatLocalLogins, (ev) => h.localLogins(ev.data)),
      Events.On(Ev.MychatHover, (ev) => h.hover(ev.data)),
    ];
    return () => off.forEach((f) => f());
  },
};
