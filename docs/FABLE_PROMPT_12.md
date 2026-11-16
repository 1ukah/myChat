# Pass 12: crisp icon files at every size

The user approved the icon you drew in pass 10
(`frontend/src/assets/mychat-icon.svg`: one glossy white MySpace person on
a black rounded tile). **Keep that design.** The problem is quality: the
Windows icons look soft.

Why: `build/windows/icon.ico` was made by `wails3 generate icons`, which
shrinks the 1024 px PNG to only 16, 32, 48, 64, 128 and 256 px. Windows also
needs 20, 24, 40 and 96 px (display scaling 125 % / 150 %) and resizes the
nearest one, and the tray got a single 64 px PNG that Windows shrinks to
16–32 px. Fine gloss and the hairline turn to mush when shrunk.

## 1. A small-size master

Add `build/icon/mychat-icon-small.svg`, the same icon redrawn for 16–32 px
(like a hand-hinted small icon):

- same black rounded tile and white person, recognisably the same icon;
- the figure may be slightly larger in the tile so it holds at 16 px;
- edges of the head, the gap and the body's flat bottom and sides land on
  whole pixels at 16, 20, 24 and 32 px where possible (pick coordinates in
  the 1024 viewBox that are multiples that work for those sizes, or design
  separate size-specific variants if needed, e.g. `-16.svg`, `-24.svg`);
- gloss reduced to what reads at that size (a simple highlight on the top
  of the head and shoulders); no hairline, no faint glows.

Same SVG rules as before: `viewBox="0 0 1024 1024"` (or the pixel size for
size-specific variants), plain shapes and gradients, no text, no filters.

## 2. A committed render script

Add `build/icon/render.mjs` (Node, ES module) that renders **every size
directly from the SVGs** (never by shrinking a bigger PNG), using
`@resvg/resvg-js`. It is **not** a project dependency: the header comment
says to install it outside the repo and how to run the script (you already
have it in `C:\Users\Admin\AppData\Local\Temp\opencode\icon-render`; resolve
the module from there, e.g. via `createRequire` on a path given by an env
var or a CLI argument, defaulting to that folder). It writes:

- `build/appicon.png`: 1024×1024 from the main SVG.
- `build/windows/icon.ico`: 16, 20, 24, 32, 40, 48, 64, 96, 128, 256.
- `build/trayicon.ico`: 16, 20, 24, 32, 40, 48, 64.
- Sizes ≤ 32 come from the small master (or its size variants), the rest
  from `frontend/src/assets/mychat-icon.svg`. Use 40 px from whichever
  looks crisper.
- ICO format: write it yourself (ICONDIR + ICONDIRENTRY + PNG payloads,
  32-bit RGBA, width/height byte 0 for 256). Entries sorted by size.
  Validate after writing: parse the file back and check the count and every
  size.
- `build/darwin/icons.icns`: from `build/appicon.png` with
  `wails3 generate icons -input appicon.png -macfilename darwin/icons.icns -windowsfilename <a path in the temp folder>`
  (run from `build/`), so the hand-made `windows/icon.ico` is not
  overwritten. The script may run this command itself.

Delete `build/trayicon.png` (the tray will embed `build/trayicon.ico`; the
Go side is being changed separately).

## 3. Review sheet

Write `C:\Users\Admin\AppData\Local\Temp\opencode\icon-render\sheet.png`:
every ICO size on a mid-grey background, each shown at 1× and enlarged 8×
with nearest-neighbour (so pixel edges are visible), labelled with its size
if easy. Look at it yourself and fix anything blurry or lopsided before you
finish.

## Rules

- Do not change Go files, `Taskfile.yml` files, `frontend/` (except nothing:
  leave `mychat-icon.svg` as is unless a tiny fix is needed for crispness),
  `docs/`, or anything Svelte/CSS. Another agent is editing `frontend/src`
  right now.
- Do not touch `frontend/package.json` or the lock file.
- Do not commit. At the end, list the files you changed or deleted and say
  how you hinted the small sizes.
