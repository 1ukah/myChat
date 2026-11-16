# Pass 13: pictures in pop-ups and the chat box

The Go side is done (do not change it): messages can now carry pictures.
Build the UI for them. `npm run check` passes now; keep it so.

## What Go sends

`chat.Entry` (`frontend/bindings/mychat/internal/chat/models.ts`) has a new
field `media: Media[]` (empty or `null` for most messages, at most one
picture for now):

```ts
Media { id: string; url: string; width: number; height: number;
        kind: "upload" | "gif" | "link"; title: string }
```

- `url` is a **local** address served by myChat (`/media/<id>`): load it
  with a plain `<img src={m.url}>`. Never load anything else from a
  message. The request waits while Go is still downloading, so the image
  may take a moment; it 404s if the download failed.
- `width`/`height` are the picture's real size (Steam sends it for uploads
  and GIFs, so the box can be sized before the picture arrives). They can
  be 0 when unknown (rare: a WebP link); then size from the loaded image.
- `kind`: `upload` = a picture sent through Steam chat, `gif` = `/giphy`,
  `link` = a picture behind an address typed in the message. `title` is a
  GIF's title (use it as `alt`), often empty.
- When a message is only a picture, its `text` is now **empty**. Text and
  picture can also come together ("look at this" + picture).
- A linked picture arrives later: the entry first comes with its text (the
  address) and no media; once Go has checked the picture it sends the same
  entry again with `media` set and the address removed from `text`
  (`mychat:entry-updated` for the chat log, and a new `mychat:toasts` list
  if it is still on screen). The UI must handle an entry gaining a picture.

New settings in `Config` (`frontend/bindings/mychat/internal/config/models.ts`):
`showImages` (default true) and `imageLinks` (default false). Go already
decides per message whether to fetch, but the UI must also respect
`showImages` live: when it is off, do not draw pictures from entries that
already carry them.

## 1. Pictures in the pop-ups (`overlay/Toasts.svelte`)

- The picture sits **under** the message text in the same box, left-aligned
  with the text (not under the avatar), in both looks (Legacy and Minimal),
  matching each look's style (Legacy: square, thin dark frame like CS2;
  Minimal: no frame, a soft edge that fits the glass band).
- Size: keep the aspect ratio, at most about **300 × 150 CSS px** (the
  overlay is scaled by the Size setting already), never wider than the box.
  Reserve the space from `width`/`height` so nothing jumps; show a quiet
  placeholder (the band colour, maybe a faint shimmer, honouring the
  Animations setting) until the image loads, then fade it in.
- GIFs play (they are GIFs); nothing else is needed.
- If the image fails to load or `showImages` is off, show a small muted
  label instead of the picture: "Image" or "GIF" (by `kind`), so a message
  that was only a picture is never an empty box.
- **Height budget.** The toast area is 360 CSS px high (`layout.toastHeight`)
  and today 5 text messages fit. With pictures that no longer holds. Rule:
  the **newest** message with a picture shows it at full size; **older**
  pop-ups with pictures shrink theirs to a small thumbnail (about 40 px
  high). If the list still does not fit, drop the oldest boxes first (the
  newest must always be fully visible together with the "[I] Reply" hint).
  Both toast layouts (stack and group) must follow this.
- The text line-clamp (2 lines) stays.

## 2. Pictures in the chat box (`overlay/ChatLine.svelte`, `Composer.svelte`)

- A message with a picture shows a **small thumbnail** under its line,
  aligned with the message text: at most about 160 × 90 px, same placeholder
  / fallback rules as the pop-ups, both looks.
- The log scrolls correctly with the taller lines (PgUp/PgDn, "↓ newer",
  the fade of the oldest lines, the `fit` estimate in `Composer.svelte`).
- An entry that gains its picture later (`mychat:entry-updated`) updates in
  place; check the store (`lib/store.svelte.ts`) replaces the whole entry,
  not only status/detail.

## 3. Settings (`settings/ChatTab.svelte`, "Pop-ups" section)

Add after "Messages shown":

- **Show images**: a toggle bound to `showImages`. Hint: "Pictures and GIFs
  sent in Steam chat, in pop-ups and the chat box."
- **Also load image links**: a toggle bound to `imageLinks`, dimmed and
  disabled while Show images is off. Hint: "Pictures from links to other
  sites. Those sites can see your IP address."

Both save like the other rows (`saveConfig`, errors per row). Add both to
the settings search index in `Settings.svelte` (words such as "picture
photo gif giphy embed link ip").

## 4. Mock / preview

- `lib/mock/backend.ts`: add `showImages: true, imageLinks: false` to the
  mock config and a few sample messages with pictures (an upload, a GIF,
  one with text + picture, and one whose picture fails), so the mock stage
  shows them. In the mock there is no `/media/` server: point sample
  `url`s at bundled files (add small sample images under
  `frontend/src/assets/` if needed; one animated GIF is enough for the GIF
  case) or at `src/assets/preview-bg.jpg`; a failing one can point at a
  missing path. Keep the mock's switches working.
- Check both looks, both toast layouts and the chat box in the mock stage,
  including the height budget with 3–5 pop-ups where several have
  pictures.

## Rules

- Do **not** edit Go files, `frontend/bindings/**`, `docs/` or `build/`.
- Never put an address from a message into `src` other than `media[].url`.
- Done means: `npm run check` reports **0 errors and 0 warnings** and
  `npm run build` succeeds (both from `frontend/`).
- Do not commit. At the end, list the files you changed and how the height
  budget works.
