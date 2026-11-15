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
| `LocalLogins(): LocalLogins` | Accounts saved by the Steam app on this PC. Also pushed as `mychat:local-logins`. |
| `UseSteamAccount(account): void` | Log in with the login the Steam app saved for this account (`""` = Steam's most recent). Rejects (throws) with a message. The choice is remembered for next start. |
| `OpenSteamLogin(): void` | Opens the Steam app. If nobody is signed in there, Steam shows its own sign-in window. myChat watches up to 10 min (`LocalLogins.watching`) and uses the new login on its own. |
| `LoginWithPassword(account, password): void` | Account name + password login, the same flow as the Steam app's login screen. Phase `submitting`, then `guard` if Steam Guard is needed, then `done` or `error`. Clear the password field as soon as it is sent. |
| `SubmitGuardCode(code): void` | Steam Guard code (authenticator app or email) during phase `guard`. A wrong code keeps phase `guard` with `codeError` set. |
| `StartLogin(): void` | Begin a QR login. Progress arrives as `mychat:login` events. |
| `CancelLogin(): void` | Stops any login in progress and the Steam sign-in watch. Phase becomes `idle`. |
| `Logout(): void` | Ends myChat's session, deletes its own saved login (QR/password) and forgets the choice. Does **not** sign the Steam app out. |

None of the three logins changes the Steam app's status: myChat never sets
or asks for its own status.
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
| `mychat:toasts` | `Toasts { entries: Entry[] }` | The pop-up messages, oldest first (at most `config.toastMax`). Replace the whole list. Also in `Snapshot.toasts`. |
| `mychat:status` | `Status` | Steam connection summary. |
| `mychat:config` | `ConfigView` | Settings changed. |
| `mychat:login` | `LoginState` | Login progress. |
| `mychat:local-logins` | `LocalLogins` | Steam's saved accounts changed (or the watch started/ended). |

## Types (summary)

```ts
ModeInfo   { mode: string; animations: boolean; blur: boolean; showAvatars: boolean;
             animationStyle: "fade"|"slide-up"|"slide-down"|"slide-left"|"slide-right"|"typewriter";
             animationSpeed: "fast"|"normal"|"slow"; hideDelayMs: number /*0 when animations off*/;
             overlayStyle: "legacy"|"minimal" }
Layout     { width: 640; composerHeight: 360; toastHeight: 360 }   // CSS px
Entry      { id: number; time: number /*unix ms*/; kind: "in"|"out"|"system"|"error";
             peerId: string; peerName: string; peerAvatar: string /*image URL or ""*/; text: string;
             status: ""|"pending"|"sent"|"failed"; detail: string /*failure reason*/ }
Channel    { kind: "none"|"whisper"; targetId: string; targetName: string; online: boolean; avatar: string }
Suggestion { id: string; name: string /*nickname the user gave, else Steam name*/;
             altName: string /*Steam name when name is a nickname, else ""*/;
             online: boolean; inGame: string; avatar: string }
ComposerState { text: string; caret: number /*index in characters (code points)*/;
             channel: Channel; hint: string; suggestions: Suggestion[]; selected: number;
             error: string; scroll: number /*log lines scrolled up from the bottom*/ }
Status     { state: "offline"|"connecting"|"online"|"needs_login"; detail: string;
             accountName: string; source: "steam"|"qr"|"password"|""; personaName: string;
             avatar: string /*the user's own avatar URL, "" until known*/ }
// "online" means myChat is logged in to Steam. It says nothing about the
// user's Steam status (myChat cannot see it), so show it as "Logged in".
Config     { hotkey: {code: string /*KeyboardEvent.code*/; ctrl; shift; alt}; closeAfterSend: boolean;
             doubleEnterClose: boolean /*only meaningful when closeAfterSend is false*/;
             overlayStyle: "legacy"|"minimal"; showAccountName: boolean /*default false*/;
             animations: boolean; animationStyle: string; animationSpeed: string; showAvatars: boolean;
             blur: boolean; showToasts: boolean; toastSeconds: number /*2..30*/;
             toastLayout: "stack"|"group"; toastMax: number /*1..5*/;
             bottomOffset: number /*0..1000 px*/; scale: number /*0.75..1.5*/; machineSeed: string }
ConfigView = Config & { hotkeyLabel: string /*e.g. "I", "Ctrl + F6"*/ }
LoginState { phase: "idle"|"waiting" /*QR shown*/|"submitting" /*password sent*/|"guard"|"done"|"error";
             method: "steam"|"qr"|"password"|"";
             url: string; qr: string[] /*rows of "1" dark / "0" light, no border*/;
             error: string /*also the reason on "idle", e.g. expired*/; account: string;
             guards: ("device_code"|"email_code"|"device_confirmation"|"email_confirmation")[];
             emailHint: string /*email domain*/; codeError: string; checking: boolean }
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
  mode and also **640×360** in `toast` mode (pop-ups are drawn from the bottom up; the rest stays transparent). Go resizes the window when the mode
  changes. Draw everything inside that viewport; it is the maximum area.
  The window can be on any monitor and DPI; never depend on screen size.
- `html, body` background must be **fully transparent**. Everything outside the
  panel must be transparent.
- Transparency: the window is truly per-pixel transparent (DirectComposition),
  click-through, with no Windows frame. Anything the page does not paint shows
  the game. Panels may leave transparent margins and have their own shape.
- No real blur: Windows blur-behind (classic and acrylic, also clipped to the
  panel with a window region) turns this window solid; tested again in pass 7.
  `blur` is always false and has no setting. Do not use CSS `backdrop-filter`
  (it has nothing to blur in this window and only costs GPU). Fake depth with
  translucent fills, fine borders and soft shadows instead.
- Two looks, chosen by `overlayStyle` (applies to the composer and pop-ups):
  - `legacy` (default): the CS2 chat look of passes 4–6, a translucent dark
    fill (about `rgba(0,0,0,0.55)`–`0.65`).
  - `minimal`: light and airy ("angelic"): frosted translucent white panels,
    dark text, thin light borders, soft shadow, rounded corners, compact
    padding, small footprint. No heavy frames or thick edges. Must stay
    readable over bright and dark scenes.
- Mode `hidden`: play the exit animation. Go hides the window `hideDelayMs`
  after `hidden` (fast 120, normal 200, slow 320; 0 when animations are off).
  Exit animations must end **20 ms before** `hideDelayMs`. Entry animations use
  the same speed (fast ≈ 120 ms, normal ≈ 200 ms, slow ≈ 320 ms).
- Pop-ups (`toast` mode) show `toasts.entries`, newest at the bottom, stacked
  upwards from the bottom of the window, with a small gap. Only a newly added
  entry animates in; the ones above slide up smoothly (transform only).
  - `toastLayout: "stack"`: one compact box per message.
  - `toastLayout: "group"` (Discord style): consecutive entries with the same
    `peerId` share one box; the name and avatar appear once at the top, the
    messages are lines below it. A different person starts a new box, so
    A, A, B, A gives three boxes: A (2 lines), B (1), A (1).
  - The sender's name is a small header (about 13–14 px, bold, the person's
    colour), never a big banner. The message text is the most readable part
    (about 15–16 px). Keep each box compact so 5 messages fit in 360 px.
  - Long messages wrap to at most 2 lines with an ellipsis.
- `animationStyle` applies to the composer panel, the toast and new log lines:
  `fade`; `slide-up` (enters moving up from below); `slide-down`; `slide-left`;
  `slide-right` (enters moving left to right); `typewriter` (the panel fades in,
  then the new message text appears character by character; exit is a fade).
  The typewriter reveal takes at most ~600 ms and never delays what the user
  types in the input line.
- Avatars (`showAvatars`): the friend's picture (`peerAvatar`, `avatar`) left of
  the toast, chat lines, suggestions and the channel label. Always square,
  like the CS2 friends list, including the whisper target in the input row.
  Images come from `https://avatars.steamstatic.com/`.
  Show a letter tile while the image loads or if it fails. When `showAvatars` is
  false, render none.
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
- Login screen (status `needs_login`). Three ways, in this order:
  0. **Use Steam on this PC** (top). `local.accounts` as account cards;
     cards with `hasLogin` → `USE THIS ACCOUNT` → `UseSteamAccount()`, the
     most recent preselected; cards without it offer `SIGN IN IN STEAM` →
     `OpenSteamLogin()`. No accounts: a line plus `OPEN STEAM SIGN-IN`. While
     `local.watching`: spinner + "Waiting for you to sign in to Steam…" +
     Cancel. If `steamInstalled` is false, only a short note.
  1. **Account name + password**. `SIGN IN` →
     `LoginWithPassword()`; Enter in either field submits. While `submitting`:
     spinner + "Contacting Steam…". Errors (`method: "password"`, phase
     `error`) show under the button. A line says the password goes only to
     Steam and is never stored.
  2. **QR code** (collapsed by default: `SIGN IN WITH QR CODE`). Expanding
     calls `StartLogin()`; collapsing calls `CancelLogin()`. Starting one way
     cancels the other.
  Phase `guard` replaces both with the Steam Guard step: 5 code boxes when
  `guards` has `device_code` or `email_code` (sent on the 5th character),
  "approve in the Steam mobile app" / "use the email link" lines for the
  confirmation guards, and `BACK` → `CancelLogin()`.
  Show `login.error` on `idle` (for example "login expired") as a banner at
  the top. When logged in, the ACCOUNT tab shows the account, `source`
  ("Using Steam on this PC" / "Password login" / "QR login") and `Log out`;
  with source `steam` also `Switch account`, which shows the cards again.
- QR login: render `LoginState.qr` as an SVG grid of squares (crisp, with a
  quiet zone of 4 modules, dark modules on a white background so phones can
  scan it).

## Chat logic you only display

Go owns all chat logic. For reference, so the UI can explain it:

- `/r [msg]` reply to last whisper · `/w name [msg]` whisper · typing `/r ` or
  `/w name ` switches the channel at once (WoW style) and clears the line.
  There is no party command.
- Friend names, avatars and online/in-game state come from Steam's web API,
  refreshed about once a minute. myChat shows no friend or party counts.
- Tab cycles recent people (Shift+Tab back). ↑/↓ sent history (or move in the
  suggestion list). PgUp/PgDn scroll the log. Enter sends (Enter on an empty line
  closes). Esc closes. Ctrl+V pastes.
