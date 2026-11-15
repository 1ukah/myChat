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
  toast(e: C.Entry): void;
  status(s: M.Status): void;
  config(c: M.ConfigView): void;
  login(l: M.LoginState): void;
}

export interface Backend {
  snapshot(): Promise<M.Snapshot>;
  overlayReady(): Promise<void>;
  saveConfig(c: Cfg.Config): Promise<M.ConfigView>;
  startLogin(): Promise<void>;
  loginWithPassword(account: string, password: string): Promise<void>;
  submitGuardCode(code: string): Promise<void>;
  cancelLogin(): Promise<void>;
  logout(): Promise<void>;
  previewToast(): Promise<void>;
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
  startLogin: () => ChatService.StartLogin(),
  loginWithPassword: (account, password) => ChatService.LoginWithPassword(account, password),
  submitGuardCode: (code) => ChatService.SubmitGuardCode(code),
  cancelLogin: () => ChatService.CancelLogin(),
  logout: () => ChatService.Logout(),
  previewToast: () => ChatService.PreviewToast(),
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
      Events.On(Ev.MychatToast, (ev) => h.toast(ev.data)),
      Events.On(Ev.MychatStatus, (ev) => h.status(ev.data)),
      Events.On(Ev.MychatConfig, (ev) => h.config(ev.data)),
      Events.On(Ev.MychatLogin, (ev) => h.login(ev.data)),
    ];
    return () => off.forEach((f) => f());
  },
};
