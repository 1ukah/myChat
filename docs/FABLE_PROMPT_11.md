# Pass 11: "?" tiles, hide every login name, a darker Minimal look, dropdowns

No Go changes. `npm run check` passes now; keep it so.

## 1. Account rows: no letter tiles

`frontend/src/lib/settings/AccountCards.svelte` (the Steam accounts list on
the login screen and under "Switch account") gives `Avatar` no picture, so
every usable account shows the first letter of its name ("N", "C", "N"). The
user does not want letters there.

- Every row shows CS2's / Steam's **"?" tile** instead (`Avatar` with
  `name=""` draws it). The account myChat is using right now may show its
  real picture (`app.status.avatar`) with the "?" tile as fallback.
- Rows of accounts the Steam app is not signed in to keep the dimmed "?"
  tile they have now.

## 2. The eye hides or shows every login name

`app.config.showAccountName` (false by default) masks the Steam login name.
In `AccountCards.svelte` it only masks the account in use
(`veiled = inUse && !app.config.showAccountName`); the other rows show their
login names in clear. That is wrong: the setting is about all login names.

- When `showAccountName` is false, **every** row shows `<NameMask />` in
  place of its login name (also on the "· Not signed in to Steam" rows).
  When it is true, every row shows its name. A row with no persona name
  shows "Steam account" as its title while masked, as now.
- The login screen (`Login.svelte`, "Use Steam on this PC", logged out) uses
  the same list but has no eye. Give it the **same eye button** as the
  Account tab (same icon, size, labels, `aria-pressed`, hover colour; save
  with `saveConfig({ showAccountName: … })` and show a failure the same way).
  Put it where it clearly belongs to the list, e.g. right-aligned in the
  "Use Steam on this PC" heading row. If the eye's markup/styles would be
  duplicated, factor it into a small `ui/` component used by both screens.
- Anywhere else a login name is shown (title bar, Account tab) already
  follows the setting; keep it that way.

## 3. Minimal look: darker, more see-through, Panorama-like

The overlay's Minimal look (`overlay/Composer.svelte`, `overlay/Toasts.svelte`,
`overlay/ChatLine.svelte`, the `.minimal` blocks) is a pale steel-blue band
with dark slate text. The user finds it **too bright**. New direction:

- **Darker band, light text.** A dim, translucent slate / steel-blue band
  (think CS2 Panorama HUD and main menu panels: deep blue-grey glass) with
  **white / near-white text**. Names keep the person's colour, now lightened
  to read on the dark band instead of darkened. "To" lines, system lines,
  errors, times, hints, the input, the caret, the send button, the name
  suggestions list and the "↓ newer" pill all move to the dark band with
  light text, consistent with each other.
- **About 15 % more transparent** than the alpha the bands have today
  (multiply the band alphas by ~0.85, after the colour change).
- **A shorter colour fade at the ends.** The colour should die out faster
  so there is no long smoky trail: `--fade` about 34 px in the chat box
  (was 60) and about 26 px in the pop-ups (was 46); shrink the haze behind
  the boxes (`.frame::before`, `.box::after`) to match: it should hug the
  box, not spread far around it.
- **Glowy, angelic, but compact.** A thin bright hairline on the top edge of
  each band (pale cyan-white, fading with the band), a tight soft bloom
  around the band, and a small light glow on the text (`--halo`, now a pale
  white/cyan glow of a few px instead of the current pale backing). It
  should feel like Panorama's lit glass, not identical to it, and stay
  tight to the content.
- It must stay readable over both a bright and a dark game scene.
- Legacy is unchanged. Keep Inter for Minimal.
- Update the Minimal hint in `OverlayTab.svelte` (currently "Pale
  steel-blue, square and compact; the colour fades out at the edges.") to
  describe the new look in the same short style, and the comments above the
  `.minimal` blocks so they describe what the code now does.
- The pop-up preview in the Overlay tab and the mock stage
  (`lib/mock/MockStage.svelte`) render the same components; check the look
  there.

## 4. Dropdowns instead of side-by-side choices

Every setting with more than one choice uses the existing dropdown
`ui/Select.svelte` instead of `ui/Segmented.svelte`:

- Overlay tab: **Chat look**, **Animation speed**.
- Chat tab: **Multiple messages**, **Messages shown**.

Keep each row's hint (Chat look and Multiple messages show the hint of the
chosen option, as now) and the disabled/dim behaviour. When nothing uses
`Segmented.svelte` any more, delete it.

## 5. No icon in the title bar

In `Settings.svelte`, the top-left mark is the app icon followed by
`myChat`. Remove the icon (and its import); keep the `myChat` text,
aligned as before.

## Rules

- Do **not** edit Go files, `frontend/bindings/**`, `docs/` or `build/`, and
  do not touch `frontend/src/assets/mychat-icon.svg` (it was just redrawn).
- Done means: `npm run check` reports **0 errors and 0 warnings** and
  `npm run build` succeeds (both from `frontend/`).
- Do not commit. At the end, list the files you changed and summarise the
  new Minimal colours (band, text, hairline, glow, alphas, fade widths).
