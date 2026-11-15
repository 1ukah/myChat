# Pass 4: redesign to the CS2 look (not CS:GO)

The user tested the app over CS2 and said the UI looks like **CS:GO's
Panorama**, not the new **CS2** UI. Redesign every screen (overlay and settings)
to match CS2. The user captured reference screenshots from CS2:

**Open every image in `reference/cs2/` with the Read tool before you start and
again while you compare.** They are the source of truth for the look. Do not
rely on memory of CS:GO.

| File | Use it for |
|---|---|
| `09-chat-box.jpg` | **The overlay chat box.** CS2's own chat: translucent black panel, square corners, bold sans lines `[CT] ● name@place: text` with coloured names, separate input row with placeholder ("Say to all") and grey `SEND` at the right, thin scrollbar. Our composer should feel like this. |
| `06-console-font.jpg` | Font feel and the console panel: grey title bar with letter-spaced `CONSOLE`, dense text, input line at the bottom. |
| `04-round-won-mvp.jpg` | Banner style for the toast: grey translucent panel with fine dot texture, light-blue letter-spaced caps, `»` `«` marks, blue pixel-mosaic MVP strip. |
| `07-map-overview-spectator.jpg`, `08-spectator-hud.jpg` | HUD panels over the game, key hints `[CTRL] Close`, circle avatar of the spectated player. |
| `01-scoreboard.jpg` | Translucent dark panel over the game, row bands, highlighted row, condensed headers, `[MOUSE2] Enable Cursor` hint. |
| `10-menu-tabs-subtabs-search.jpg`, `11-settings-video-search.jpg`, `02-settings-crosshair.jpg` | **The settings window.** Top nav in big condensed caps with thin vertical separators; sub-tab row with a **magnifying-glass search button** on the left; tabs in small bold condensed caps; the **active tab is a filled steel-blue box with light-blue text**; a second sub-tab row in the same style; italic bold grey section title ("Video"); rows with a regular-weight label left and a **condensed bold caps value + chevron** right, separated by thin light lines; sliders with a thin track, small square thumb and the value in a bordered box; `RESET` with a history icon bottom right; cool blue-grey smoky blurred background with topographic contour lines in the header and the right rail. |
| `05-friends-panel.jpg` | Friends list: **square** avatars, bold names, small grey sub-lines, country pills, mail icon, contour-line background, `?` placeholder avatar. Use it for the account cards and suggestion rows. |
| `03-detail-crop.png` | A tiny accidental crop; ignore it. |

## Rules

- Re-read `docs/UI_CONTRACT.md`. It changed: the overlay window is now truly
  per-pixel transparent, click-through and frameless; **blur is off by
  default** and the panel must look right without it (translucent dark fill
  like `09-chat-box.jpg`). Panels may now have transparent margins. Label the
  blur toggle "Blur behind (experimental, may look solid)".
- Keep all behaviour, bindings, store logic, mock modes, avatars, animation
  styles and the login flow from passes 1–3. This pass is visual. Avatar shape
  follows CS2: square in lists and cards (`05`), circle only where CS2 uses a
  circle (spectated player, `07`/`08`); keep the coloured ring or border idea.
- Fonts: CS2 uses Valve's proprietary "Stratum2". Do not ship it. Pick the
  closest free fonts you can bundle locally (fontsource), for the condensed caps
  and for the bold chat text, and match weight, width and letter spacing to the
  screenshots.
- Colours: sample them from the screenshots (active tab blue, light-blue text,
  CT blue / T gold, row lines, panel fills). Put them in `theme.css`.
- Work only in `frontend/` (plus reading `reference/` and `docs/`). Do not edit
  Go files or `frontend/bindings/`.

## Compare, don't guess

You may run `npm run dev` (port 9245) in the background and use
`npx playwright screenshot` to capture the mock URLs (for example
`http://127.0.0.1:9245/?view=overlay&mock=composer&bg=bright`,
`?view=overlay&mock=toast`, `?view=settings&mock=online&tab=overlay`,
`?view=settings&mock=login`) into the system temp folder, then Read your
screenshots next to the matching reference image and iterate until they read as
the same game. Stop the dev server when done.

## Finish

`npm run check` and `npm run build` must pass with zero errors and zero
warnings. Reply with: what you changed per screen, which fonts you chose and
why, which references you matched each screen to, the screenshots you took (if
any), and anything that still differs from CS2.
