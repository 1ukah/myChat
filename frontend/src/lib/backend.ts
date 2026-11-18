import { Events } from "@wailsio/runtime";
import { ChatService } from "../../bindings/mychat";
import * as Ev from "../../bindings/events";
import type * as M from "../../bindings/mychat/models";
import type * as C from "../../bindings/mychat/internal/chat/models";
import type * as Cfg from "../../bindings/mychat/internal/config/models";

export interface Handlers {
  mode(m: M.ModeInfo): void;
  composer(c: C.ComposerState): void;
  entry(e: C.Entry): void;
  entryUpdated(e: C.Entry): void;
  logCleared(): void;
  toasts(t: M.Toasts): void;
  status(s: M.Status): void;
  config(c: M.ConfigView): void;
  login(l: M.LoginState): void;
  localLogins(l: M.LocalLogins): void;
  hover(key: string): void;
}

export interface Backend {
  snapshot(): Promise<M.Snapshot>;
  overlayReady(): Promise<void>;
  saveConfig(c: Cfg.Config): Promise<M.ConfigView>;
  localLogins(): Promise<M.LocalLogins>;
  useSteamAccount(account: string): Promise<void>;
  openSteamLogin(): Promise<void>;
  startLogin(): Promise<void>;
  loginWithPassword(account: string, password: string): Promise<void>;
  submitGuardCode(code: string): Promise<void>;
  cancelLogin(): Promise<void>;
  logout(): Promise<void>;
  previewToast(): Promise<void>;
  setSpots(spots: M.Spot[]): Promise<void>;
  clickSpot(key: string): Promise<void>;
  closePicture(): Promise<void>;
  overlayResized(): Promise<void>;
  hideSettings(): Promise<void>;
  minimiseSettings(): Promise<void>;
  quit(): Promise<void>;
  subscribe(h: Handlers): () => void;
}

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
  overlayResized: () => ChatService.OverlayResized(),
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
