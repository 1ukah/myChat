# Pass 9: own typography, the new icon, Bliss preview, plain toggles

No Go changes. `npm run check` passes now; keep it so.

## 1. Toggles without "Yes" / "No"

The toggle switch already says yes or no by its position and colour. Remove
the `Yes`/`No` text next to every toggle (keep an accessible label for
screen readers, e.g. `aria-label` / `aria-checked`).

## 2. Typography: stop cloning the CS2 menu

The settings window copies CS2's Panorama menu 1:1, including its fonts. The
user likes the layout (keep where everything is, the colours and the overall
dark look) but finds the fonts ugly: the italic section titles ("*Steam*",
"*Chat box*", "*Appearance*"), the condensed caps fonts (Saira family) and the
all-caps everywhere.

- Use **Inter** for the whole settings window. It is installed:
  `import "@fontsource-variable/inter";` (family name `"Inter Variable"`),
  with the existing Noto Sans / Segoe UI fallbacks for other scripts.
- No italics anywhere. Section titles become clean, slightly larger semibold
  headings. Tabs and sub-tabs, buttons, labels: sentence or title case, not
  all caps; use weight and colour for emphasis instead.
- Keep it tidy and modern: consistent sizes (a small type scale), comfortable
  line heights, numbers with tabular figures where values line up.
- This applies to the settings window (all tabs, the login screens, Help).
  The **overlay Legacy look keeps its CS2 fonts** (it is meant to look like
  CS2's chat). The overlay **Minimal** look should use Inter too.
- If a Saira package ends up unused, remove its import (do not touch
  `package.json`).

## 3. The new app icon

Design the myChat icon as SVG, inspired by two references the user sent:

- **The MySpace logo:** three people side by side, growing from left to
  right (the right one is the biggest). Each has a round head above a body
  with a rounded top; the bodies overlap, and their bottoms form a staircase
  (left body ends highest, the right body reaches lowest). Flat, bold, no
  outlines.
- **Steam's default avatar:** a dark grey square with a light grey `?` in the
  middle.

The icon: the same three MySpace figures in the same positions, but each
**head is a `?`** (Steam's default-avatar question mark) instead of a circle.
Use **Steam's colours**: a dark navy tile (`#171a21` → `#1b2838`, a subtle
`#2a475e` glow is fine) with the figures in Steam light blue (`#66c0f4`, may
fade towards `#c7d5e0`). Square tile with small or no corner radius. It must
stay readable at 16 and 32 px (bold shapes, no thin details).

Deliver:
- `frontend/src/assets/mychat-icon.svg`: the icon on its tile,
  `viewBox="0 0 1024 1024"`, no external fonts (draw the `?` as paths), no
  filters that need a browser (it will be rasterised for the Windows icon).
- In the settings title bar, top left, replace the current "myChat" wordmark
  with this icon (about 22–26 px) followed by the text `myChat`.

## 4. Pop-up preview background

The preview in the Overlay tab ("Animation speed" section, with the Replay
button) draws a split sky/sand scene. Use the Windows XP "Bliss" wallpaper
instead: `frontend/src/assets/preview-bg.jpg` (1280×720: bright blue sky with
white clouds over a green hill). `background-size: cover`, centred so the hill
line is visible behind the pop-ups. Keep the Replay button readable on it.

## Rules

- Do **not** edit Go files, `frontend/bindings/**`, `docs/` or `build/`.
- Done means: `npm run check` reports **0 errors and 0 warnings** and
  `npm run build` succeeds (both from `frontend/`).
- Do not commit. At the end, list the files you changed and describe the icon.
