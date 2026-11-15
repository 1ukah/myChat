# Pass 5: stacked and grouped pop-ups, smaller names

The user tested pass 4 over CS2 and liked it. Three changes to the pop-ups
(toast mode) and their settings. The Go side is done; bindings are regenerated.

**Open every image in `reference/toasts/` with the Read tool first:**

| File | What it shows |
|---|---|
| `01-discord-grouped-messages.png` | Discord: several messages in a row from one person share one block, name once at the top. This is the `group` layout. |
| `02-multiple-people.png` | Messages from different people, each with its own header. |
| `03-small-name-size.png` | The name size the user wants: small, bold, readable (`november` in the CS2 HUD). Ignore the lightning icon. |

Also keep looking at `reference/cs2/` for the CS2 look of pass 4 (keep it).

## What changed in Go (see `docs/UI_CONTRACT.md`, re-read it)

- Event `mychat:toast` (one Entry) is **gone**. New `mychat:toasts` sends
  `Toasts { entries: Entry[] }`: the recent pop-up messages, oldest first, at
  most `config.toastMax`. Initial value in `Snapshot.toasts`. Replace the
  whole list on each event.
- Config: `toastLayout: "stack" | "group"` (default `group`) and
  `toastMax: 1..5` (default 3).
- The toast-mode window is now **640×360** (was 640×92), fully transparent and
  click-through. Draw the pop-ups from the bottom up.
- `npm run check` currently fails on the old `MychatToast` event; fix that.

## Build

1. **Pop-up list** per the contract's "Pop-ups" rules:
   - `stack`: one compact box per message.
   - `group`: consecutive entries with the same `peerId` share one box, name
     and avatar once, messages as lines below. A, A, B, A → three boxes.
   - Newest at the bottom; only new entries animate in (using the user's
     `animationStyle`/`animationSpeed`), older boxes slide up (transform).
   - Exit animation on mode `hidden` as before (`hideDelayMs`).
2. **Smaller name.** The big `» NAME «` banner of pass 4 clutters the screen.
   Make the name a small bold header (~13–14 px, coloured), like `03`. The
   message text stays the readable part. Keep the CS2 feel (translucent dark
   panel, fine lines, square avatar) but compact: 5 messages must fit in
   360 px in both layouts. Long text: max 2 lines, then ellipsis. You may keep a
   subtle hint like `[I] Reply` once, under the newest box, not on every box.
3. **Settings → Overlay → Pop-ups section:** add
   - "Multiple messages": segmented `STACK` / `GROUP` with a one-line hint each.
   - "Messages shown": 1–5 (slider or segmented).
   Add both to the settings search index.
4. **Animation preview / PreviewToast** in settings: show 3 sample messages
   from 2 people so the layout choice is visible.
5. **Mock mode:** `?view=overlay&mock=toast` shows 4 entries (A, A, B, A);
   add `&layout=stack|group` and `&max=1..5` params; a mock button or timer
   that adds a new message every ~2 s is welcome (`mock=toaststream`).

## Rules

- Work only in `frontend/` (read `reference/` and `docs/`). Do not edit Go
  files or `frontend/bindings/`.
- Overlay rules still apply: no focus/inputs, transparent background, animate
  only transform/opacity, no timers or rAF loops while hidden.

## Compare, don't guess

Run `npm run dev` (port 9245) in the background and `npx playwright
screenshot --channel msedge` the mock URLs (both layouts, max 1 and 5, a long
message, a bright background `&bg=bright`) into the temp folder, Read them next
to the references, iterate. Stop the dev server when done.

## Finish

`npm run check` and `npm run build` pass with zero errors and warnings. Reply
with what changed, the mock URLs, the screenshots you took, and anything left.
