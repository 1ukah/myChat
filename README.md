# myChat

myChat is a Steam chat overlay for Windows. With myChat, you can read and send Steam messages in games and in other applications. Counter-Strike 2 was the first target, but you can use myChat in all games.

## Requirements

- Windows 10 or Windows 11.
- A Steam account.
- For games: windowed mode or fullscreen windowed mode.

## Install

1. Download `mychat.exe` from the latest release.
2. Start `mychat.exe`.
3. Log in to Steam.

## Use

1. Push the hotkey to open the chat box. The default hotkey is `I`.
2. Type the message.
3. Push `Enter` to send the message.

## Build

You must have Go, Node.js and the Wails 3 CLI.

```
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0
wails3 build
```

The build writes the file `bin\myChat.exe`.

## Release

1. Write the new version in `VERSION`.
2. Write the changes for the new version in `CHANGELOG`.
3. Push a tag with the same version, for example `v0.1.0`.

GitHub Actions builds `mychat.exe` and publishes the release.
