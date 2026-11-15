# Pass 2: Panorama feel, avatars, animation styles, normal Steam login

You built the first UI pass in this session. The Go backend has changed since
then. Re-read `docs/UI_CONTRACT.md` (updated) and the regenerated
`frontend/bindings/` (new fields and methods). Work only in `frontend/`. Do not
edit Go files or `frontend/bindings/`.

## 1. Design direction: "Panorama-alike", not a copy

The user sent screenshots of the CS2 menu (play tab, settings) and of the
in-game HUD (scoreboard, spectator player cards). Do not copy either one
exactly. Aim for the overall CS2 Panorama feeling everywhere:

- Smoky, heavily blurred dark glass with a subtle vertical gradient (darker
  edges, slightly lighter centre). Thin 1 px light borders at low opacity.
  Square corners.
- Condensed, letter-spaced uppercase for labels, headers and numbers; clean
  sans for message text. Near-white text, muted grey secondary text.
- Restrained accents: CS2 green for party/online/positive, a gold highlight
  for the selected/active row (as in the scoreboard's highlighted player), light
  blue for an active tab, red for errors.
- Key hints styled like the CS2 scoreboard footer: small boxed keys, for example
  `[ENTER] Send  [ESC] Close  [TAB] Next  [/HELP]`.
- Row backgrounds: faint translucent bands. The selected suggestion uses the gold
  highlight treatment.
- Spectator-card style for people: a **circle avatar with a thin coloured ring**
  (ring colour: whisper violet for whispers, green for party, grey offline).

Review every screen you built and push it closer to this feeling. Keep the
overlay minimal: it sits over a live match.

## 2. Avatars (setting `showAvatars`)

- Circle avatar to the left of: the toast (larger, about 44 px), each chat line
  (small, about 18 px, inline), suggestions (about 22 px), and the channel label
  of the input bar (about 18 px).
- Use `peerAvatar` / `avatar` URLs. Until the image loads, or if it fails, show a
  circle with the first letter. Use `loading="eager"` and `decoding="async"`.
  Remember failed URLs so they are not retried every render.
- Party lines and system lines get no avatar (a small icon for party is fine).
- When `ModeInfo.showAvatars` is false, render no avatars and no empty space.
- Settings → OVERLAY: a "Show avatars" toggle.
- Mock mode: realistic avatar URLs are not reachable offline, so generate
  data-URI SVG avatars (coloured circles with initials or simple shapes). One
  mock friend should have a broken URL to show the fallback.

## 3. Animation styles

Read `animationStyle`, `animationSpeed` and `hideDelayMs` from `ModeInfo`.

| Style | Enter | Exit |
|---|---|---|
| `fade` | opacity 0→1 | 1→0 |
| `slide-up` | moves up from about 14 px below + fade | moves down + fade |
| `slide-down` | moves down from about 14 px above + fade | moves up + fade |
| `slide-left` | moves left from about 24 px right + fade | reverse |
| `slide-right` | moves right from about 24 px left + fade | reverse |
| `typewriter` | panel fades in, then message text types out letter by letter | fade |

- It applies to the composer panel, the toast, and each new log line (log
  lines use a shorter, smaller version of the same motion).
- Speed: fast about 120 ms, normal about 200 ms, slow about 320 ms. Exit must end
  20 ms before `hideDelayMs`.
- Typewriter: reveal the toast text and new incoming log lines character by
  character (whole grapheme clusters, never split emoji). The total reveal is
  ≤ 600 ms, whatever the message length (more characters per step for long
  text). Never apply it to the user's own input line. Implement it with a CSS
  `steps()` animation, or with a short-lived timer that stops when finished. No
  timers may run while the overlay is hidden.
- `animations: false` or `prefers-reduced-motion`: no motion at all.
- Only `transform` and `opacity` are animated.
- Settings → OVERLAY: an "Animation style" picker that looks like a CS2
  dropdown or segmented control, an "Animation speed" picker
  (FAST / NORMAL / SLOW), and a small live preview tile in the settings window
  that replays the chosen animation on a sample toast. Disable both pickers when
  "Animations" is off.

## 4. Normal Steam login (not QR only)

The user cannot use QR (they use a Windows authenticator app). Rebuild the
ACCOUNT login screen as described in the contract (section "Settings window
rules → Login screen"):

- Make it look like the native Steam sign-in, in our Panorama style: account
  name + password on the left, QR on the right ("OR SIGN IN WITH QR").
- Steam Guard step (phase `guard`): a code input (5 large boxes or one wide box,
  uppercase, auto-submit at 5 characters, paste works), the right text for
  `device_code` and `email_code` (with `emailHint`), and a hint for
  `device_confirmation`. Show `checking` and `codeError`. Back → `CancelLogin()`.
- `submitting`: a spinner, and the form is disabled.
- `error`: show the message under the form and keep the account name.
- A trust line: "Your password goes only to Steam, encrypted with Steam's own
  key. myChat never stores it."
- Never store, log or echo the password. Clear it after submit.
  `autocomplete="current-password"` is fine.
- Mock mode: `?view=settings&mock=login` shows the form. Password `guard`
  leads to the guard step with `device_code`; code `ABCDE` succeeds, other codes
  give "Wrong Steam Guard code." Password `wrong` gives "Wrong account name or
  password." Add `mock=guard` for a direct preview.

## Finish

Fix all type errors caused by the new binding fields (mock data included).
`npm run check` and `npm run build` must pass with zero errors and zero new
warnings. Reply with a short summary and the list of mock URLs.
