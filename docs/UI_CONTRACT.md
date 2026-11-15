# myChat UI contract

The Go backend is finished. The UI only renders state and calls the service.
Do **not** edit Go files or anything in `frontend/bindings/` (generated).

## Stack

- Svelte 5 (runes) + TypeScript + Vite, in `frontend/`.
- Wails v3 runtime: `import { Events } from "@wailsio/runtime"`.
- Service: `import { ChatService } from "../bindings/mychat"` (path relative to `frontend/src/`).
- Types: `import * as M from "../bindings/mychat/models"` and
  `import * as C from "../bindings/mychat/internal/chat/models"` and
  `import * as Cfg from "../bindings/mychat/internal/config/models"`.
- Event name constants: `import * as Ev from "../bindings/events"`.
- One bundle, two views, chosen by the query string:
  - `/?view=overlay` – the in-game overlay window.
  - `/?view=settings` – the settings window.

## Service methods (`ChatService`)

| Method | Notes |
|---|---|
| `Snapshot(): Snapshot` | Full state. Call once on mount, in both views. |
| `OverlayReady(): void` | **Overlay view only.** Call once after the first render. Required, or the overlay never opens. |
| `SaveConfig(c: Cfg.Config): ConfigView` | Save settings. Rejects (throws) with a message, for example when a hotkey is taken. Send the whole config object, keeping `machineSeed` as received. |
| `LocalLogins(): LocalLogins` | Accounts saved by the Steam client on this PC. Also pushed as `mychat:local-logins`. |
| `UseSteamAccount(account): void` | Log in with the login the Steam client saved for this account (`""` = Steam's most recent). Rejects (throws) with a message. The choice is remembered for next start. |
| `OpenSteamLogin(): void` | Opens the Steam app. If nobody is signed in there, Steam shows its own sign-in window (password, Steam Guard, QR). myChat watches up to 10 min (`LocalLogins.watching`) and uses the new login on its own. |
| `StartLogin(): void` | Begin a QR login (myChat's own session). Progress arrives as `mychat:login` events. |
| `CancelLogin(): void` | Stops a QR login and the Steam sign-in watch. Phase becomes `idle`. |
| `Logout(): void` | Ends myChat's session and forgets the choice. Does **not** sign the Steam app out. |
| `PreviewToast(): void` | Shows a sample toast in game position. |
| `HideSettings(): void` | Close button of the settings window (it hides; the app keeps running in the tray). |
| `MinimiseSettings(): void` | |
| `Quit(): void` | Exit the app. |

## Events (Go → UI)

Listen with `Events.On(name, (ev) => ev.data)`. `ev.data` is the typed object.

| Event | Data | Meaning |
|---|---|---|
| `mychat:mode` | `ModeInfo {mode, animations, blur}` | `mode` is `"hidden" \| "toast" \| "composer"`. |
| `mychat:composer` | `ComposerState` | Input line state. Replace the whole object. |
| `mychat:entry` | `Entry` | Append to the log. |
| `mychat:entry-updated` | `Entry` | Replace the entry with the same `id` (status changes). |
| `mychat:log-cleared` | – | Empty the log. |
| `mychat:toast` | `Entry` | Show this entry as the toast. |
| `mychat:status` | `Status` | Steam connection summary. |
| `mychat:config` | `ConfigView` | Settings changed. |
| `mychat:login` | `LoginState` | Login progress. |
| `mychat:local-logins` | `LocalLogins` | Steam's saved accounts changed (or the watch started/ended). |

## Types (summary)

```ts
ModeInfo   { mode: string; animations: boolean; blur: boolean; showAvatars: boolean;
             animationStyle: "fade"|"slide-up"|"slide-down"|"slide-left"|"slide-right"|"typewriter";
             animationSpeed: "fast"|"normal"|"slow"; hideDelayMs: number /*0 when animations off*/ }
Layout     { width: 640; composerHeight: 360; toastHeight: 92 }   // CSS px
Entry      { id: number; time: number /*unix ms*/; kind: "in"|"out"|"party"|"system"|"error";
             peerId: string; peerName: string; peerAvatar: string /*image URL or ""*/; text: string;
             status: ""|"pending"|"sent"|"failed"; detail: string /*failure reason*/ }
Channel    { kind: "none"|"whisper"|"party"; targetId: string; targetName: string; online: boolean; avatar: string }
Suggestion { id: string; name: string; online: boolean; inGame: string; avatar: string }
ComposerState { text: string; caret: number /*index in characters (code points)*/;
             channel: Channel; hint: string; suggestions: Suggestion[]; selected: number;
             error: string; scroll: number /*log lines scrolled up from the bottom*/ }
Status     { state: "offline"|"connecting"|"online"|"needs_login"; detail: string;
             accountName: string; source: "steam"|"qr"|""; personaName: string;
             friendsOnline: number; partySize: number }
Config     { hotkey: {code: string /*KeyboardEvent.code*/; ctrl; shift; alt}; closeAfterSend: boolean;
             animations: boolean; animationStyle: string; animationSpeed: string; showAvatars: boolean;
             blur: boolean; showToasts: boolean; toastSeconds: number /*2..30*/;
             bottomOffset: number /*0..1000 px*/; scale: number /*0.75..1.5*/; machineSeed: string }
ConfigView = Config & { hotkeyLabel: string /*e.g. "I", "Ctrl + F6"*/ }
LoginState { phase: "idle"|"waiting" /*QR shown*/|"done"|"error"; method: "steam"|"qr"|"";
             url: string; qr: string[] /*rows of "1" dark / "0" light, no border*/;
             error: string /*also the reason on "idle", e.g. expired*/; account: string }
LocalLogins { steamInstalled: boolean; watching: boolean; error: string;
             accounts: { accountName: string; personaName: string; steamId: string;
                         mostRecent: boolean; hasLogin: boolean /*usable saved login*/ }[] }
Snapshot   { mode; log; composer; status; config; login: LoginState; local: LocalLogins; layout }
// Config also has loginSource / steamAccount: send them back unchanged (Go ignores them).
```

## Overlay window rules (critical)

- The window is **never focused** and **ignores the mouse**. Keys are caught by a
  Windows keyboard hook in Go. So: **no `<input>`, no focus, no hover states, no
  clicks.** Draw the text and the caret from `ComposerState` (`caret` counts code
  points: use `Array.from(text)`).
- The window is exactly the layout size in CSS px: **640×360** in `composer`
  mode and **640×92** in `toast` mode. Go resizes the window when the mode
  changes. Draw the panel to fill the whole viewport (`position: fixed; inset: 0`).
  The window can be on any monitor and DPI; never depend on screen size.
- `html, body` background must be **fully transparent**. Everything outside the
  panel must be transparent.
- Blur: when `blur` is true, Windows blurs the game behind the **whole window**,
  so the panel must cover the whole window edge to edge with square corners. Use
  a semi-transparent dark fill over the blur (about `rgba(12,14,17,0.55)`). When
  `blur` is false use a more opaque fill (about `rgba(12,14,17,0.86)`) so text
  stays readable over the game. Do not use CSS `backdrop-filter` (it has nothing
  to blur in this window and only costs GPU).
- Mode `hidden`: play the exit animation. Go hides the window `hideDelayMs`
  after `hidden` (fast 120, normal 200, slow 320; 0 when animations are off).
  Exit animations must end **20 ms before** `hideDelayMs`. Entry animations use
  the same speed (fast ≈ 120 ms, normal ≈ 200 ms, slow ≈ 320 ms).
- `animationStyle` applies to the composer panel, the toast and new log lines:
  `fade`; `slide-up` (enters moving up from below); `slide-down`; `slide-left`;
  `slide-right` (enters moving left to right); `typewriter` (the panel fades in,
  then the new message text appears character by character; exit is a fade).
  The typewriter reveal takes at most ~600 ms and never delays what the user
  types in the input line.
- Avatars (`showAvatars`): a circle with the friend's picture (`peerAvatar`,
  `avatar`) left of the toast, chat lines, suggestions and the channel label.
  Images come from `https://avatars.steamstatic.com/`. Show a letter circle while
  the image loads or if it fails. When `showAvatars` is false, render none.
- When `animations` is false: no transitions, no caret blink (steady caret).
- Performance: animate only `transform` and `opacity`. No `setInterval`, no
  `requestAnimationFrame` loops, no work while mode is `hidden`. The caret blink
  is a CSS animation that runs only in `composer` mode. Keep the DOM small (render
  at most the last ~60 log lines).
- The overlay sits at the bottom middle of the screen over the game, so it must
  read well over bright and dark scenes.

## Settings window rules

- Frameless, 980×680 (min 820×560), normal focus and mouse.
- Draw a custom title bar. Drag area: CSS `--wails-draggable: drag;`. Buttons
  inside it need `--wails-draggable: no-drag;`. Minimise → `MinimiseSettings()`,
  close → `HideSettings()`.
- Save on change (no Save button). Show `SaveConfig` errors inline.
- Hotkey capture: a button that, when clicked, listens to the next `keydown`
  (use `e.code`, `e.ctrlKey`, `e.shiftKey`, `e.altKey`; ignore pure modifier
  presses; `Escape` cancels) and saves it.
- Login screen (status `needs_login`). There is **no password form**; myChat
  never asks for a password. Three ways, in this order:
  1. **Use Steam on this PC (default, top, largest).** List `local.accounts`
     as account cards (persona name big, account name small, "Most recent"
     tag). Cards with `hasLogin` get a primary `USE THIS ACCOUNT` →
     `UseSteamAccount(accountName)`; the most recent one is preselected. Cards
     without `hasLogin`: muted, "Not signed in to Steam" + `SIGN IN IN STEAM`.
     Line under the list: "Uses the login the Steam app saved on this PC. It
     never leaves this PC." Show `UseSteamAccount` errors inline on the card.
  2. **Sign in with the Steam app.** `OPEN STEAM SIGN-IN` → `OpenSteamLogin()`.
     Text: "Sign in to Steam with \"Remember me\" on. myChat continues on its
     own." While `local.watching`: spinner + "Waiting for you to sign in to
     Steam…" + Cancel → `CancelLogin()`.
  3. **QR code** (secondary, collapsed by default: `SIGN IN WITH QR CODE`).
     When expanded call `StartLogin()`; collapsing calls `CancelLogin()`.
  If `steamInstalled` is false, hide 1 and 2 and show the QR directly. Show
  `login.error` (for example "saved login expired") as a banner at the top.
  When logged in, the ACCOUNT tab shows the account, `source` ("Using Steam on
  this PC" / "QR login"), and `Log out`; with source `steam` also a
  `Switch account` button that shows the account list again.
- QR login: render `LoginState.qr` as an SVG grid of squares (crisp, with a
  quiet zone of 4 modules, dark modules on a white background so phones can
  scan it).

## Chat logic you only display

Go owns all chat logic. For reference, so the UI can explain it:

- `/r [msg]` reply to last whisper · `/w name [msg]` whisper · `/p [msg]` CS2
  party · typing `/r `, `/p ` or `/w name ` switches the channel at once (WoW
  style) and clears the line.
- Tab cycles recent people (Shift+Tab back). ↑/↓ sent history (or move in the
  suggestion list). PgUp/PgDn scroll the log. Enter sends (Enter on an empty line
  closes). Esc closes. Ctrl+V pastes.
