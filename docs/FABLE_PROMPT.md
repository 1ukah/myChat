# Task: build the myChat UI (overlay + settings) in the Counter-Strike 2 Panorama style

You are building the complete front end of **myChat**, a Windows desktop app
(Go + Wails v3 + Svelte 5). The user plays Counter-Strike 2 and wants to read and
answer **Steam private messages without leaving the game**. They press a hotkey
(default `I`), a minimal chat box appears at the **bottom middle** of the
screen over the game, they type, press Enter, and it goes to Steam. New messages
pop up as a small toast at the same place.

The Go backend is finished and tested. Your job is **only the UI** in
`frontend/`. Read `docs/UI_CONTRACT.md` first: it is the binding API contract
(methods, events, types, window rules). Follow it exactly.

## What exists

- `frontend/` – Vite + Svelte 5 + TS project. `src/App.svelte` is a throwaway
  test page: replace it completely. `src/main.ts` mounts `App`.
- `frontend/bindings/` – generated, typed Wails bindings. Do not edit.
- Go code at the repo root and in `internal/` – do not edit.

## Deliverables

1. `?view=overlay` – the in-game overlay with two modes:
   - **composer** (640×360 CSS px): the chat box.
   - **toast** (640×92 CSS px): one incoming message.
2. `?view=settings` – the settings window (980×680, frameless, resizable).
3. A **mock mode** for development without Go: when the Wails runtime is not
   available (plain `npm run dev` in a browser), feed the views with realistic
   sample data and simulate events. Support `?view=overlay&mock=composer`,
   `?view=overlay&mock=toast`, `?view=settings&mock=login`,
   `?view=settings&mock=online` so every state can be previewed. In mock mode
   draw a game-like background behind the overlay (for example a dark gradient
   with some bright shapes) so readability can be judged, but in the real app the
   page background stays fully transparent. Keep mock code out of the hot path
   (a separate module, imported dynamically only when the runtime is missing).
4. `npm run check` (svelte-check) and `npm run build` must pass with zero errors
   and zero warnings that you introduced.

Organise the code cleanly (for example `src/lib/` for the store, components for
`Overlay`, `Composer`, `Toast`, `ChatLine`, `Settings`, `QrCode`, `Toggle`,
`Slider`, `HotkeyInput`, and a `theme.css`). One reactive store module wraps
`Snapshot()` and all events.

## Visual direction: CS2 Panorama UI

It must look like it shipped inside Counter-Strike 2. The Panorama menus use:

- **Typography**: a condensed, slightly squared sans for headings, navigation,
  labels and values, usually in UPPERCASE (Valve's Stratum2). Use
  **Barlow Condensed** (500/600/700, plus italic 600/700) as the stand-in, and
  **Noto Sans** for body/chat text. Install them with `@fontsource/barlow-condensed`
  and `@fontsource/noto-sans` and bundle them locally (no network at runtime).
  Fall back to `"Segoe UI", sans-serif` for scripts these fonts lack (players
  have Arabic, Cyrillic, CJK names).
- **Colours**: near-white text (`#e8eaed`), muted grey labels (`#8f979e`), dark
  smoky panels (`rgba(16,18,21,0.6–0.9)`), thin 1 px separators
  (`rgba(255,255,255,0.12)`), a light-blue active accent (`#86c5ff`-ish, as the
  active "PLAY" tab and the selected settings tab), CS2 party/lobby **green**
  (`#3fae49`, toggles are this green with a white knob), and a strong **red**
  (`#c8352f`) for errors/destructive actions. Gold (`#e0b252`) is fine for small
  highlights.
- **Shapes**: square or 2 px corners, thin outlines. The active sub-tab is
  either an outlined box or a filled muted-blue box. Buttons are flat boxes
  with letter-spaced uppercase labels (like "CANCEL SEARCH": red fill, red
  border, wide tracking).
- **Texture**: the menu panels have a faint **topographic contour line**
  pattern (thin wavy height lines, very low opacity). Recreate it as an inline
  SVG pattern for the settings window background and side panels.
- **Settings page look** (match this closely): a top bar with large condensed
  uppercase tabs separated by thin vertical lines; below, a row of small
  uppercase sub-tabs; then a centred column of rows. Each row: label on the
  left (regular weight, light grey), control on the right; rows divided by thin
  lines; a section title in grey bold italic (like "Video"). Dropdown/value
  text is uppercase condensed bold with a chevron. Sliders are a thin track with
  a small white handle and a boxed numeric value to the right ("93%"). Bottom
  right: uppercase action links with icons ("RESET DEFAULTS").
- **Friends/party panel look** for status and lists: square avatars (we have no
  avatars: use a square with the first letter), white semibold names, small
  grey status text, envelope icon for messages.

## Overlay design

Minimal and functional, like a Linux app launcher (rofi/wofi) crossed with the
CS2 in-game chat. The user is mid-match: it must be **instant to read, never
cluttered, and must not hide much of the screen**.

**Composer (640×360)**
- Bottom: the **input bar** is the hero. It is a single line with the channel
  label on the left, in WoW style: `TO SANT1:` (whisper), `PARTY:`, or a muted
  `CHAT:` when no channel is set (show `hint` as placeholder text). Use
  condensed uppercase for the label and colour it by channel (whisper pink-violet
  `#e58cff`-ish, party CS2 green, none grey). For a whisper, show a small
  online dot for `channel.online`. Then the text with a drawn caret (thin
  vertical bar, blinks via CSS only when animations are on). Long text scrolls
  horizontally so the caret stays visible.
- `error` shows as a short red line just above the input bar (it is cleared by
  the next key).
- `suggestions` (while typing `/w name`): a compact list just above the input
  bar, highlight `selected`, each with an online dot and `inGame` in small grey
  text. Max 6 items.
- Above that: the **chat log**, newest at the bottom, WoW style, one entry per
  line (wrapping allowed):
  - `in`: `[12:41] [SANT1] whispers: text` – pink-violet name and text.
  - `out`: `[12:41] To [SANT1]: text` – slightly lighter/less saturated.
  - `party`: `[12:41] [Party] You: text` – CS2 green.
  - `system`: grey italic. `error`: red.
  - `status` `pending`: dimmed with a tiny spinner/dots. `failed`: red mark +
    `detail` in small red text.
  - Respect `scroll` (lines up from the bottom); when scrolled show a small
    "↓ newer" marker. Fade the top of the log with a mask so old lines melt
    away.
- A tiny footer/header strip with key hints in muted text (`Enter send · Esc
  close · Tab next · /help`) and the Steam status when it is not `online`
  (e.g. red dot "Steam offline – reconnecting", or "Log in from the tray icon").
- Entry animation: panel slides up 8–12 px and fades in; new log lines fade/slide
  in subtly; exit: fade + small drop within 160 ms.

**Toast (640×92)**
- One card filling the window: sender name (condensed uppercase, pink-violet),
  message text (max 2 lines, ellipsis), time, and a muted hint
  `PRESS I TO REPLY` using `config.hotkeyLabel`. Small envelope icon. A thin
  accent line at the left or top edge. A newer toast replaces the old one with
  a quick cross-fade/slide.

## Settings window design

Tabs (top bar): **ACCOUNT · CHAT · OVERLAY · HELP**. Title bar with an app
mark ("myChat" in condensed italic bold, like "PRIME ENABLED"), drag region,
minimise and close buttons.

- **ACCOUNT**: Steam status card (persona name, account, state dot,
  friends online, party size). If `needs_login`/no account: explain "Log in
  with the Steam mobile app. Your password never touches myChat." and a big
  `LOG IN WITH QR` button → `StartLogin()`. While `waiting`, show the QR
  (`LoginState.qr`) large and crisp with steps: open Steam app → Steam Guard
  icon → scan. Cancel button. On `error` show the message and a retry button.
  When online, show `LOG OUT`.
- **CHAT**: open-chat hotkey (capture control showing `hotkeyLabel`), "Close
  after sending" toggle, "Show new message pop-ups" toggle, "Pop-up duration"
  slider (2–30 s).
- **OVERLAY**: "Animations" toggle, "Background blur" toggle (subtext: "Turn off
  on low-end PCs"), "Size" slider (75–150 %), "Distance from bottom" slider
  (0–600 px), `PREVIEW POP-UP` button → `PreviewToast()`.
- **HELP**: the commands (`/r`, `/w name`, `/p`, Tab, ↑↓, PgUp/PgDn, Esc,
  Ctrl+V) in a clean two-column table; a short note: "myChat never opens,
  reads or writes the game process. It uses a normal Windows hotkey and a
  keyboard hook only while the chat box is open."; a `QUIT MYCHAT` red button.

## Quality bar

- Pixel-careful, consistent spacing (4 px grid), crisp 1 px lines, no blurry
  text (avoid fractional transforms at rest).
- Everything must be readable at 75 % scale.
- No layout jumps when text changes; long names and long messages must not break
  the layout.
- All animations respect `animations` (overlay: from `ModeInfo`; settings:
  from config) and `prefers-reduced-motion`.
- No external network requests. No new heavy dependencies beyond the two
  fontsource packages (no UI kits, no icon fonts; inline SVG icons).
- TypeScript strict, no `any` where a binding type exists.

When finished, run `npm run check` and `npm run build` in `frontend/` and fix
everything. Then reply with a short summary of the files you created and how to
preview each mock state.
