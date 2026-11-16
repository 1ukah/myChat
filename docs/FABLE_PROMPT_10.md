# Pass 10: a new app icon (one glossy MySpace person, black and white)

Only the icon changes in this pass. No Go changes, no other UI changes.

## The icon

The user rejected the current icon (three figures with `?` heads in Steam
colours). The new one:

- **One person, centred**, in the style of the classic **MySpace logo**
  figure: a round head above a body whose top is a wide rounded dome
  (shoulders) and whose bottom is cut flat. Bold, simple, flat-silhouette
  proportions like MySpace's people. No `?`, no face, no arms, no text.
- **Monochrome:** the figure is **white**, the tile is **black**. No colour
  anywhere (greys are fine).
- **Glossy**, like a polished app icon, but still clean:
  - the figure has a white → light grey vertical gradient;
  - a curved glossy highlight (a soft white band whose lower edge is an arc)
    across the upper part of the head and across the shoulders;
  - the black tile has a subtle sheen: a very faint lighter top half or a
    soft radial glow behind the figure, and maybe a hairline inner edge
    highlight. Keep it subtle: it must still read as black.
- **Tile:** a square with a modest corner radius (about 18 to 22 % of the
  side, app-icon style), filling the whole canvas.
- **Readable at 16 and 32 px:** the figure fills roughly 60 to 70 % of the
  tile, the gap between the head and the body is clear but not thin, and
  there are no fine details that turn to mush when small.

## Deliver

1. `frontend/src/assets/mychat-icon.svg`: `viewBox="0 0 1024 1024"`, plain
   shapes and gradients only (no `<text>`, no external fonts or images, no
   `filter` elements, since it is rasterised outside a browser too). Keep
   the `<title>myChat</title>`. Short comments, like the current file.
2. Render it to PNG:
   - `build/appicon.png`: 1024×1024 (the source of the Windows `.ico` and
     the macOS `.icns`).
   - `build/trayicon.png`: **exactly 64×64** (`icon_test.go` checks this;
     it is embedded by `icon.go`). Make sure it still looks good at that
     size; if the gloss gets muddy at 64 px you may simplify the tray render
     slightly, as long as it is clearly the same icon.
   - No SVG rasteriser is installed. Put any tool you need **outside the
     repo**, in `C:\Users\Admin\AppData\Local\Temp\opencode\icon-render`
     (for example `npm i @resvg/resvg-js` there and a small Node script, or
     headless Edge). Do **not** touch `frontend/package.json` or the lock
     file.
3. From `build/`, regenerate the Windows and macOS icons:
   `wails3 generate icons -input appicon.png -macfilename darwin/icons.icns -windowsfilename windows/icon.ico`
4. Also save previews for review at
   `C:\Users\Admin\AppData\Local\Temp\opencode\icon-render\preview-256.png`
   and `preview-32.png`.

## Rules

- Do not change Go files, `frontend/bindings/**`, `docs/`, or any Svelte or
  CSS file. The settings title-bar change is part of a later pass.
- Done means: the SVG and the four build icons are written, and
  `go test ./...` passes from the repo root.
- Do not commit. At the end, list the files you changed and describe the icon
  in two or three sentences.
