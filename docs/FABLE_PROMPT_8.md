# Pass 8: a real "hidden" account name, and Minimal redone in Panorama colours

No Go changes. Re-read `docs/UI_CONTRACT.md` (the `minimal` look and
`showAccountName` lines changed). `npm run check` passes now; keep it so.

## 1. Account name: a hidden effect, not just gone

Today, with `config.showAccountName` false the login name simply disappears,
and with it true it shows as `· ruajosedaux117` with an eye-off button. The
user wants a visible *hidden* state:

- Hidden (`showAccountName` false, the default): in the same place show a
  masked placeholder: a short run of dots or a pill of the same size that is
  blurred / frosted (CSS `filter: blur()` is fine here, this is the settings
  window, not the overlay). It must **never** contain the real name (do not
  blur the real text; it can be read back). Next to it an **eye** button
  ("Show account name").
- Shown: the real name, and an **eye-off** button ("Hide account name").
- Toggle with `saveConfig({ showAccountName })` as now. A short fade between
  the two states is welcome when animations are on.
- Same rule anywhere else the logged-in login name appears.

## 2. Minimal look, redone

The user tried Minimal and it is not what they want:

- **Square**, not rounded: 0–2 px corners on panels, input, pills, avatars.
- **Not white.** Use the CS:GO **Panorama** palette, a brighter version of
  it: pale steel blue / blue-grey tints (think `#c9d8e4`, `#a9bfd0`,
  `#8fa9bd` with soft cyan accents like `#9fd1f2`), translucent. No pure or
  near-pure white fills.
- **"Angelic", colour fading away to nothing.** Panels should not be boxes
  with hard edges: the tint should fade into transparency towards the edges
  (for example a gradient from the tinted side to transparent, and/or
  `mask-image` gradients on the panel edges), with a soft glow/haze around
  text and panels that suggests blur. No real blur exists (no
  `backdrop-filter`, see the contract), so fake it with layered gradients,
  glow (`box-shadow`/`text-shadow` with large soft radii) and translucency.
- No heavy frames, no thick borders; at most a hairline highlight that also
  fades out.
- Compact and small on screen, as now.
- **Readable** over bright and dark scenes: check the mock with
  `&look=minimal&bg=bright` and `&look=minimal&bg=dark`, composer and toast
  (`&mock=toast`) views. Pick the text colour (dark slate on the pale tint, or
  light text with a soft dark halo) that stays readable on both.
- Applies to the composer (log, input row, suggestions, key hints) and the
  pop-ups (stack and group layouts), with avatars on and off.
- Legacy must not change at all.

All overlay rules still apply: no focus, no hover, transparent outside the
panels, `transform`/`opacity` animations only, the animation styles and
speeds, `hideDelayMs` exit timing.

## Rules

- Do **not** edit Go files, `frontend/bindings/**` or `docs/`.
- Done means: `npm run check` reports **0 errors and 0 warnings** and
  `npm run build` succeeds (both from `frontend/`).
- Do not commit. At the end, list the files you changed and describe the
  Minimal colours and fade technique you used.
