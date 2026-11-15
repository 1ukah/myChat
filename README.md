# myChat

Read and answer Steam private messages from inside Counter-Strike 2, without
Shift+Tab.

- Press the hotkey (default **I**). A small chat box opens at the bottom middle.
- Type and press **Enter**. The message goes to Steam.
- New messages pop up at the same place.
- You keep moving while you type: keys you held before you opened the box
  (for example W) still go to the game.

## Commands (WoW style)

| Input | Action |
|---|---|
| `/r [msg]` | Reply to the last person who messaged you |
| `/w name [msg]` | Message a friend (part of the name is enough) |
| `Tab` / `Shift+Tab` | Cycle recent people |
| `↑` `↓` | Sent history |
| `PgUp` `PgDn` | Scroll |
| `Enter` on empty line, `Esc` | Close |
| `Ctrl+V` | Paste |

Typing `/r ` or `/w name ` switches the channel at once.

## Requirements

- Windows 10 or 11. Blur needs Windows 11.
- CS2 in **Fullscreen Windowed**.
- Use a hotkey that CS2 does not use.

## Safety

myChat never opens, reads or writes the CS2 process and never changes game
files. It uses:

- a normal Windows hotkey (`RegisterHotKey`),
- a keyboard hook (`WH_KEYBOARD_LL`) only while the chat box is open,
- a window that never takes focus, so CS2 stays the active window.

It logs in to Steam as a quiet second session (like the Steam mobile app), in
one of three ways: the login the Steam app already saved on this PC, account
name + password + Steam Guard, or a QR code scanned with the Steam mobile app.
The password is encrypted with Steam's key and sent only to Steam; it is never
stored. myChat stays a silent, offline session: it never sets a status and
never asks the Steam connection for anyone's status, so your Steam app stays
signed in to Friends & Chat with your status and a correct friends list.
Friend names and avatars come from Steam's web API instead. The login token is encrypted
with Windows DPAPI in `%APPDATA%\myChat`.

## Build

```
wails3 build
```

Output: `bin\myChat.exe`.

Development:

```
cd frontend
npm run dev        # UI preview in a browser with mock data (?view=overlay&mock=composer)
```

## Layout

| Path | What |
|---|---|
| `internal/steam` | Minimal Steam client (WebSocket CM, QR and password login, friends, chat) |
| `internal/input` | Hotkey and keyboard hook thread |
| `internal/chat` | Chat logic (channels, commands, history) |
| `internal/config` | Settings and encrypted session |
| `internal/win` | Win32 calls |
| `app.go`, `overlay.go`, `main.go` | Wails app, overlay window control, tray |
| `frontend/` | Svelte UI (overlay + settings) |
| `docs/UI_CONTRACT.md` | API between Go and the UI |
