# Pass 7: own avatar, Minimal look, calmer settings

The Go side is done and the bindings are regenerated. Re-read
`docs/UI_CONTRACT.md` first (it changed: `Status.avatar`, `Suggestion.altName`,
`ModeInfo.overlayStyle`, config `doubleEnterClose`, `overlayStyle`,
`showAccountName`, the blur finding, the two looks, square avatars).
`npm run check` passes right now; keep it that way.

## Changes the user asked for

1. **Square avatar in the input row.** When replying, the whisper target in
   the composer's input row is a circle with a ring, while chat lines use
   squares. Make it a square too (same look as the chat-line avatars, a bit
   larger is fine). Online state can stay as a small colour cue.

2. **The user's own avatar.** The Account tab status card and the top-right
   summary in the title bar show a letter tile ("N"). Use `status.avatar` (an
   image URL) when it is not empty, with the letter tile as fallback while it
   loads or if it fails (the `Avatar` component already does fallbacks).

3. **Remove the right column.** The settings window still has the empty
   right rail (contour lines, ~50 px wide, where the friend and party counters
   used to be). Remove the rail completely and let the content use the space.

4. **Hide the account name by default, with an eye button.** The status card
   shows `· ruajosedaux117` (the Steam login name, `status.accountName`) next
   to the connection state. Hide it unless `config.showAccountName` is true,
   and add a small eye / eye-off icon button there that toggles it and saves
   with `saveConfig({ showAccountName })`. Apply this to every place that
   shows the logged-in account's login name (status card, title-bar summary
   if it shows it, Account tab hints). The login and switch-account cards keep
   showing account names (the user needs them to pick an account).

5. **"Logged in" instead of "Online".** myChat cannot see the user's Steam
   status any more, so `status.state === "online"` means "logged in to Steam".
   Show **Logged in** with the app's blue accent (`--accent` / steel blue)
   instead of green "ONLINE": the dot, the label, and the avatar ring of the
   status card and the title-bar summary. Connecting and offline keep their
   current colours.

6. **Normal case on setting controls.** The controls on the right of the
   settings rows are all caps (`CHANGE`, `PAGEDOWN`, `NO`, `YES`, `6 S`,
   `STACK`, `GROUP`, ...). Use sentence case instead: `Change`, `PageDown`,
   `No`, `Yes`, `Stack`, `Group`, `Fade`, ... and write seconds as `6s` (no
   space, lower-case s). This covers toggles, segmented controls, selects, the
   hotkey control, slider values and number pickers, and the buttons inside
   settings rows (`Change`, `Preview`, `Switch account`, `Log out`, ...). Keep
   the tab names, section titles and the overlay's CS2 labels as they are.

7. **Double-tap Enter to close.** In the Chat tab, directly below "Close after
   send", add a toggle row "Double-tap Enter to close" bound to
   `config.doubleEnterClose`. Show it only while `closeAfterSend` is off (it
   has no effect otherwise). Hint, short: Enter on an empty line closes the
   chat only when pressed twice quickly.

8. **Minimal look.** Add an option (Overlay tab, near the animation options)
   to choose the chat look: **Legacy** (the current CS2-style chat, unchanged)
   or **Minimal**, bound to `config.overlayStyle`. The overlay reads it from
   `mode.overlayStyle`. Build Minimal for the composer and the pop-ups:
   light and "angelic", frosted translucent white panels, dark text, thin
   light borders, soft shadow, rounded corners, compact spacing, smaller
   footprint than Legacy, no heavy frames or thick edges. There is **no real
   blur** (see the contract); fake the frosted feel with translucency, a
   subtle white gradient and shadow. It must stay readable over a bright and
   a dark game scene (check with the mock's `&bg=bright` and `&bg=dark`).
   All overlay rules still apply (no focus, no hover, transparent outside the
   panels, transform/opacity animations only, animation styles and speeds,
   avatars on/off, stack/group pop-up layouts).

9. **Nicknames in /w suggestions.** A suggestion with a non-empty `altName`
   is a friend the user gave a nickname: show the nickname (`name`) as now and
   the Steam name (`altName`) small and muted next to it.

10. **Text cleanup.** There is too much text, and too many lines talk about
    "myChat" itself ("myChat uses the login…", "Ends myChat's session…",
    "myChat runs as a quiet second session…", "Your Steam app's status is not
    touched", "It never leaves this PC", ...). Go through the settings window
    and the login screens and remove text the user does not need. Keep short
    hints that explain what an option does or help choose between options;
    keep error messages. Do not mention "myChat" in hints at all. Keep the
    one line on the password form that says the password goes only to Steam
    and is not stored (it builds trust), but make it short.

11. **Mock backend** (`src/lib/mock/backend.ts`, `MockStage.svelte`): add
    `avatar` to the mock statuses (a generated data-URI avatar like the friend
    ones), the new config fields with their defaults (`doubleEnterClose:
    true`, `overlayStyle: "legacy"`, `showAccountName: false`), a URL switch
    `&look=minimal|legacy`, a link for it in the mock stage, and one mock
    friend with a nickname (`name` = nickname, `altName` = Steam name).

## Rules

- Do **not** edit Go files, `frontend/bindings/**` (generated) or `docs/`.
- Keep Legacy looking exactly as it does now, except the square avatar (1).
- Done means: `npm run check` reports **0 errors and 0 warnings** and
  `npm run build` succeeds (both from `frontend/`).
- Do not commit. At the end, list the files you changed and, per numbered
  item above, what you did.
