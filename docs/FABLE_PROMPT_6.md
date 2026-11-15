# Pass 6: remove the party channel and the friend/party counters

The Go side is done and the bindings are regenerated. myChat is now a silent,
offline Steam session (it must never change the user's Steam status), so it
no longer knows about CS2 parties and does not count friends. The user asked
for the party command and the counters to be removed. Read
`docs/UI_CONTRACT.md` again first; it is the source of truth.

## What changed in Go

- `Status` lost `friendsOnline` and `partySize`. It is now
  `{ state, detail, accountName, source, personaName }`.
- `Entry.kind` no longer has `"party"`: only `"in" | "out" | "system" | "error"`.
- `Channel.kind` no longer has `"party"`: only `"none" | "whisper"`.
- The `/p` and `/party` commands are gone. Composer hints from Go are now
  `"/r to reply · /w name"` and `"/w name to whisper · /help"`. The empty-channel
  error is `"Pick who to message: /w name or /r"`.
- Friend `online`/`inGame` on suggestions and the whisper channel still work
  (they now come from Steam's web API, refreshed about once a minute). Keep
  those dots as they are.

`npm run check` currently fails with 14 errors in 3 files, all from the removed
fields. Fix them by removing the features, not by adding fields back.

## Build

1. **Settings, right rail** (`Settings.svelte`, "right rail" with the two
   counters): remove the friends-online counter and the party counter. Keep
   the rail itself with its contour lines, so the window layout and the CS2
   look do not change. Also drop the now unused status/counter code around
   "Steam summary (top right, and the counters in the rail)", but keep the top
   right summary (name and connection state).
2. **Account tab** (`AccountTab.svelte`): remove the "Friends online" and "In
   your party" stats from the status card. The card keeps the round avatar,
   name, connection state and account name; make it look balanced without the
   stats (no empty gap on the right).
3. **Overlay** (`ChatLine.svelte`, `Composer.svelte` and anything else): remove
   the `[PARTY]` line rendering and the party channel label/styling. Keep
   whisper and system lines exactly as they look now.
4. **Help tab** (`HelpTab.svelte`): remove `/p`. Settings search words in
   `Settings.svelte` must not mention `/p` or party.
5. **Mock backend** (`src/lib/mock/backend.ts`, `MockStage.svelte`): remove
   party everywhere (the imitation engine's `/p`, `sendParty`, party log lines,
   hints, HELP lines, `friendsOnline`/`partySize` in statuses). Keep every
   other mock scenario working, including all login mocks.
6. Remove `--party` from `theme.css` if nothing uses it any more.

## Rules

- Do **not** edit Go files, `frontend/bindings/**` (generated), or anything in
  `docs/`. Do not touch the login screen (`Login.svelte`, `AccountCards.svelte`,
  `GuardStep.svelte`) except where the compiler forces you to.
- Keep the existing CS2 style, sizes and animations. No new features.
- Done means: `npm run check` reports **0 errors and 0 warnings**, and
  `npm run build` succeeds. Run both from `frontend/`.
- Do not commit. At the end, list the files you changed and what you removed.
