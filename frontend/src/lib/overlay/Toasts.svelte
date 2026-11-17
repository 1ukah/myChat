<script lang="ts">
  // The pop-ups (toast mode, 640×360): the recent incoming messages, newest
  // at the bottom, stacked upwards from the bottom edge of the window.
  //
  //   stack  one compact box per message
  //   group  messages in a row from one person share a box (Discord style):
  //          avatar and name once, the messages as lines below
  //
  // Each box is a small CS2-style translucent dark panel: square avatar, the
  // name as a small coloured header, the message as the readable part.
  // Only a newly added message animates in; what is already on screen slides
  // up to make room (transform only, one Web Animation per box, no loops).
  //
  // A message can carry a picture, drawn under its text. The window is only
  // 360 px high, so: the newest picture is shown at full size (300 × 150 at
  // most), pictures in older messages shrink to thumbnails (40 px high), and
  // if the list still does not fit, the oldest messages are left out (`fit`).
  // A click on a picture opens it at full size (`zoom`); the pop-ups then
  // stay until it is closed.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
  import { app, clickSpot } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import { pictureOf, TOAST_FULL, TOAST_THUMB } from "./media";
  import type { Motion } from "./motion";
  import Picture from "./Picture.svelte";
  import { nameSpot } from "./spots";
  import Typed from "./Typed.svelte";

  let {
    entries,
    layout,
    hotkey,
    live,
    avatars = false,
    images = true,
    minimal = false,
    zoom = false,
    motion,
  }: {
    /** Oldest first; the whole list is replaced on every change. */
    entries: C.Entry[];
    /** `toastLayout`: "stack" or "group" (anything else counts as group). */
    layout: string;
    /** Label of the open-chat hotkey, e.g. "I" or "Ctrl + F6". */
    hotkey: string;
    /** The pop-ups are on screen (not fading out, not hidden). */
    live: boolean;
    /** Show the senders' avatars (setting `showAvatars`). */
    avatars?: boolean;
    /** Draw the pictures in messages (setting `showImages`); off shows a label instead. */
    images?: boolean;
    /** The Minimal look (`overlayStyle`): dark blue-grey glass, light text, fading at the edges. */
    minimal?: boolean;
    /** Pictures open at full size and names start a whisper when clicked (the in-game overlay, not previews). */
    zoom?: boolean;
    motion: Motion;
  } = $props();

  interface Line {
    entry: C.Entry;
    /** The message's picture, if it has one. */
    media: C.Media | null;
    /** Joined a box that was already on screen: the line itself moves in. */
    fresh: boolean;
    /** Typewriter style: the text types itself out. */
    typed: boolean;
  }

  interface Box {
    key: number;
    name: string;
    avatar: string;
    time: number;
    tint: string;
    who: string;
    /** Arrived while the pop-ups were on screen: the box moves in. */
    fresh: boolean;
    lines: Line[];
  }

  // Plain memory of what was shown last time, so a new list can tell which
  // messages are new. Decided once per message and then kept, so re-running
  // this never restarts an animation.
  let marks = new Map<number, { enter: "box" | "line" | ""; typed: boolean }>();
  /** Entry id → key of its box. A box keeps its key when its oldest line drops out. */
  let keys = new Map<number, number>();
  let wasLive = false;
  let lastGrouped: boolean | null = null;

  const boxes = $derived.by((): Box[] => {
    const grouped = layout !== "stack";
    const relaid = lastGrouped !== null && lastGrouped !== grouped;
    // Coming back from hidden the whole panel animates in; nothing extra then.
    const animate = live && wasLive && !relaid;
    const typewriter = untrack(() => motion.typewriter);

    const nextMarks: typeof marks = new Map();
    const nextKeys: typeof keys = new Map();
    const out: Box[] = [];

    for (const entry of entries) {
      const who = entry.peerId || entry.peerName;
      const last = out[out.length - 1];
      const join = grouped && last !== undefined && last.who === who;

      const mark = relaid
        ? { enter: "" as const, typed: false }
        : (marks.get(entry.id) ?? {
            enter: animate ? (join ? ("line" as const) : ("box" as const)) : ("" as const),
            typed: typewriter,
          });
      nextMarks.set(entry.id, mark);

      const line: Line = {
        entry,
        media: pictureOf(entry),
        fresh: mark.enter === "line",
        typed: mark.typed,
      };
      if (join) {
        last.lines.push(line);
        nextKeys.set(entry.id, last.key);
        continue;
      }
      const key = relaid ? entry.id : (keys.get(entry.id) ?? entry.id);
      nextKeys.set(entry.id, key);
      out.push({
        key,
        name: entry.peerName,
        avatar: entry.peerAvatar,
        time: entry.time,
        tint: playerColor(who),
        who,
        fresh: mark.enter === "box",
        lines: [line],
      });
    }

    marks = nextMarks;
    keys = nextKeys;
    wasLive = live;
    lastGrouped = grouped;
    return out;
  });

  /** The newest message with a picture: the only one shown at full size. */
  const fullId = $derived.by(() => {
    for (let i = entries.length - 1; i >= 0; i--) {
      if (pictureOf(entries[i])) return entries[i].id;
    }
    return null;
  });

  /** "Ctrl + F6" → "CTRL+F6", as CS2 writes key hints. */
  const key = $derived((hotkey || "I").replace(/\s*\+\s*/g, "+").toUpperCase());
  const typeDelay = $derived(Math.round(motion.enterMs * 0.6));

  // ---- making room: older boxes slide up when a message is added ----
  //
  // Positions are read just before the DOM changes and compared just after.
  // A box that only moved slides as a whole. A box that grew (a new line in a
  // group) stays put and its old content slides inside it.

  let list = $state<HTMLElement>();
  /** Top and bottom (layout px inside the list) before the change. */
  const before = new Map<HTMLElement, { top: number; bottom: number }>();
  const sliding = new WeakMap<HTMLElement, Animation>();

  const PARTS = ".face, .title, .msg";

  // ---- the height budget: what does not fit is left out, oldest first ----
  //
  // The list is built up from the bottom edge, so too much content sticks
  // out above the top of the window. This walks from the oldest message on
  // and hides whole messages until the rest fits: a line of a grouped box
  // first, and the box itself with its last line. The newest message is
  // never hidden, so it and the key hint are always fully visible (a
  // full-size picture with two lines of text is about 220 px). Run after
  // every change, starting from everything shown.

  function fit(root: HTMLElement): void {
    for (const el of root.querySelectorAll<HTMLElement>("[hidden]")) el.hidden = false;
    const all = Array.from(root.querySelectorAll<HTMLElement>(".box"));
    for (let b = 0; b < all.length; b++) {
      const box = all[b];
      // Everything older is hidden by now: this box is the top of the list.
      const over = (): boolean => box.offsetTop < 0;
      const msgs = Array.from(box.querySelectorAll<HTMLElement>(".msg"));
      for (let m = 0; m < msgs.length - 1; m++) {
        if (!over()) return;
        msgs[m].hidden = true;
      }
      if (b === all.length - 1 || !over()) return;
      box.hidden = true;
    }
  }

  function measure(root: HTMLElement): void {
    for (const box of root.querySelectorAll<HTMLElement>(".box")) {
      const top = box.offsetTop;
      before.set(box, { top, bottom: top + box.offsetHeight });
      for (const part of box.querySelectorAll<HTMLElement>(PARTS)) {
        before.set(part, { top: top + part.offsetTop, bottom: 0 });
      }
    }
  }

  function slide(el: HTMLElement, from: number): void {
    sliding.get(el)?.cancel();
    if (Math.abs(from) < 0.5) return;
    sliding.set(
      el,
      el.animate([{ transform: `translateY(${from}px)` }, { transform: "none" }], {
        duration: motion.enterMs,
        easing: "cubic-bezier(0.2, 0.8, 0.2, 1)",
      }),
    );
  }

  function play(root: HTMLElement): void {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    for (const box of root.querySelectorAll<HTMLElement>(".box")) {
      const old = before.get(box);
      // Left out now, or was left out before: nothing to slide.
      if (!old || box.hidden || old.bottom === old.top) continue;
      const top = box.offsetTop;
      const dTop = old.top - top;
      const dBottom = old.bottom - (top + box.offsetHeight);
      if (Math.abs(dTop - dBottom) < 0.5) {
        slide(box, dTop);
        continue;
      }
      for (const part of box.querySelectorAll<HTMLElement>(PARTS)) {
        const was = before.get(part);
        if (was) slide(part, was.top - (top + part.offsetTop));
      }
    }
  }

  $effect.pre(() => {
    void boxes;
    before.clear();
    if (list && live && motion.on) measure(list);
  });

  $effect(() => {
    void boxes;
    void fullId;
    void images;
    void avatars;
    void minimal;
    if (list) fit(list);
    if (list && before.size > 0) play(list);
    before.clear();
  });

  /** A picture's box changed after it was drawn (its size came with the image). */
  function refit(): void {
    if (list) fit(list);
  }
</script>

<div class="toasts" class:minimal bind:this={list}>
  {#each boxes as box (box.key)}
    <!-- With `zoom` the name is a link: a click opens the chat box
         whispering to this person. -->
    {@const spot = zoom ? nameSpot(box.who, box.key) : ""}
    <div class="box" class:fresh={box.fresh} style:--who={box.tint}>
      {#if avatars}
        <span class="face">
          <Avatar size={32} tint={box.tint} src={box.avatar} name={box.name} />
        </span>
      {/if}
      <div class="body">
        <div class="title">
          <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
          <bdi
            class="name"
            class:link={spot !== ""}
            class:hover={spot !== "" && app.hover === spot}
            data-spot={spot || undefined}
            onclick={spot ? () => clickSpot(spot) : undefined}>{box.name}</bdi
          >
          <span class="time">{clock(box.time)}</span>
        </div>
        {#each box.lines as line (line.entry.id)}
          {@const full = line.entry.id === fullId}
          <div class="msg" class:fresh={line.fresh}>
            <!-- A message that is only a picture has no text. -->
            {#if line.entry.text}
              <div class="text">
                <bdi>
                  {#if line.typed}
                    <Typed text={line.entry.text} delay={typeDelay} />
                  {:else}
                    {line.entry.text}
                  {/if}
                </bdi>
              </div>
            {/if}
            {#if line.media}
              <div class="media" class:plain={!images}>
                <Picture
                  media={line.media}
                  show={images}
                  maxW={full ? TOAST_FULL.w : TOAST_THUMB.w}
                  maxH={full ? TOAST_FULL.h : TOAST_THUMB.h}
                  {minimal}
                  {zoom}
                  onsize={refit}
                />
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/each}
  {#if boxes.length > 0}
    <div class="hint">
      {#if minimal}
        <span><b>{hotkey || "I"}</b> to reply</span>
      {:else}
        <span>[{key}] Reply</span>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* Built from the bottom edge up; everything above stays see-through. */
  .toasts {
    --width: 460px;
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    font-family: var(--font-head);
  }

  /* About 0.65 black, like CS2's own chat box, with a fine line in the
     person's colour on the left and a hairline on top. */
  .box {
    position: relative;
    flex: none;
    display: flex;
    gap: 9px;
    width: var(--width);
    padding: 5px 12px 5px 9px;
    background:
      linear-gradient(var(--box-fill), var(--box-fill)),
      var(--frame-fill);
    box-shadow:
      inset 2px 0 0 var(--who),
      inset 0 1px 0 rgba(255, 255, 255, 0.07);
  }

  .face {
    flex: none;
    display: flex;
    height: 32px;
    margin: 1px 0 0 3px;
  }

  .body {
    flex: 1;
    min-width: 0;
  }

  /* The name: a small bold header, not a banner. (Not `.head`: that is the
     global caps style.) */
  .title {
    display: flex;
    align-items: baseline;
    gap: 7px;
    height: 15px;
    line-height: 15px;
    white-space: nowrap;
  }

  .name {
    min-width: 0;
    overflow: hidden;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.01em;
    text-overflow: ellipsis;
    color: var(--who);
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.9);
  }

  /* A name under the mouse pointer looks like a hovered link. `.hover`
     comes from Go (the window never sees the mouse); `:hover` is for the
     browser mock. */
  .name.link {
    cursor: pointer;
  }

  .name.link.hover,
  .name.link:hover {
    color: color-mix(in srgb, var(--who) 65%, #fff);
    text-decoration: underline;
    text-decoration-thickness: 1px;
    text-underline-offset: 2px;
  }

  .time {
    flex: none;
    font-size: 11px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--muted);
  }

  /* The message: the readable part. Two lines at most, then an ellipsis. */
  .text {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    overflow: hidden;
    font-size: 15px;
    font-weight: 600;
    line-height: 19px;
    overflow-wrap: anywhere;
    color: #fff;
    text-shadow:
      1px 1px 0 rgba(0, 0, 0, 0.9),
      0 0 2px rgba(0, 0, 0, 0.7);
  }

  /* The picture, under the text and in line with it. */
  .media {
    display: flex;
    margin: 3px 0 2px;
  }

  /* Pictures are off: only the small label. */
  .media.plain {
    margin: 1px 0;
  }

  /* Left out by the height budget (see `fit`). */
  .box:global([hidden]),
  .msg:global([hidden]) {
    display: none;
  }

  /* One quiet key hint under the newest box. */
  .hint {
    flex: none;
    display: flex;
    justify-content: flex-end;
    width: var(--width);
    height: 16px;
  }

  .hint span {
    padding: 0 6px;
    font-size: 11px;
    font-weight: 600;
    line-height: 16px;
    letter-spacing: 0.02em;
    color: #c5ccd3;
    background: rgba(0, 0, 0, 0.45);
  }

  /* ---------- Minimal look ----------
     Each box is a band of lit glass, after the CS2 Panorama panels: a dim,
     see-through blue-grey tint with near-white text, square, with no frame.
     The tint lives on a pseudo-element that reaches a little past the text
     and is masked away to nothing at both ends (--ends, --fade), with a
     feathered bottom. On its top edge is a bright cyan-white hairline with
     a few px of light under it, fading out with the band. A tight, faint
     cyan bloom behind it and a small pale glow on the text (over a 1 px
     dark edge, for bright scenes) stand in for blur. The person's colour,
     lightened to read on the dark tint, is kept for the name. Narrower than
     Legacy. */

  .minimal {
    --width: 380px;
    --fade: 26px;
    --avatar-radius: 0px;
    --muted: #a9b9c8;
    --tile: #4a5f75;
    --halo: 0 1px 1px rgba(4, 10, 18, 0.55), 0 0 5px rgba(178, 226, 255, 0.5);
    --ends: linear-gradient(
      to right,
      transparent 0,
      rgba(0, 0, 0, 0.26) calc(var(--fade) * 0.4),
      rgba(0, 0, 0, 0.8) calc(var(--fade) * 0.8),
      #000 var(--fade),
      #000 calc(100% - var(--fade)),
      rgba(0, 0, 0, 0.8) calc(100% - var(--fade) * 0.8),
      rgba(0, 0, 0, 0.26) calc(100% - var(--fade) * 0.4),
      transparent 100%
    );
    font-family: var(--font-ui);
  }

  .minimal .box {
    z-index: 0;
    gap: 8px;
    padding: 5px 12px 6px 9px;
    background: none;
    box-shadow: none;
  }

  .minimal .box::before {
    content: "";
    position: absolute;
    z-index: -1;
    inset: 0 calc(var(--fade) * -1);
    background:
      linear-gradient(
          to right,
          rgba(150, 214, 255, 0.6),
          rgba(236, 248, 255, 0.9) 50%,
          rgba(150, 214, 255, 0.6)
        )
        top / 100% 1px no-repeat,
      linear-gradient(to bottom, rgba(168, 222, 255, 0.2), rgba(168, 222, 255, 0) 8px),
      linear-gradient(
        to bottom,
        rgba(34, 48, 64, 0.73) 0,
        rgba(20, 30, 42, 0.75) calc(100% - 4px),
        rgba(20, 30, 42, 0) 100%
      );
    -webkit-mask-image: var(--ends);
    mask-image: var(--ends);
  }

  /* The bloom: hugs the band and is gone a few px past its ends. */
  .minimal .box::after {
    content: "";
    position: absolute;
    z-index: -2;
    inset: -6px -34px;
    background: radial-gradient(
      closest-side,
      rgba(120, 196, 255, 0.16) 0,
      rgba(120, 196, 255, 0.12) 66%,
      rgba(120, 196, 255, 0) 100%
    );
  }

  .minimal .face {
    margin: 2px 0 0;
  }

  .minimal .name {
    letter-spacing: 0;
    color: color-mix(in srgb, var(--who) 58%, #fff);
    text-shadow: var(--halo);
  }

  .minimal .name.link.hover,
  .minimal .name.link:hover {
    color: color-mix(in srgb, var(--who) 34%, #fff);
  }

  .minimal .time {
    font-weight: 400;
  }

  .minimal .text {
    font-weight: 400;
    color: #f4f8fb;
    text-shadow: var(--halo);
  }

  .minimal .hint {
    height: 17px;
  }

  /* A short strip of the same dark tint, faded at both ends. */
  .minimal .hint span {
    padding: 0 14px;
    font-weight: 400;
    line-height: 17px;
    letter-spacing: 0;
    color: #c9d6e2;
    text-shadow: var(--halo);
    background: linear-gradient(
      to right,
      rgba(24, 36, 50, 0),
      rgba(24, 36, 50, 0.68) 14px,
      rgba(24, 36, 50, 0.68) calc(100% - 14px),
      rgba(24, 36, 50, 0)
    );
  }

  .minimal .hint b {
    font-weight: 700;
    color: #f4f8fb;
  }

  /* A new message moves in the way the chosen style says (--dx/--dy). */
  :global(.anim) .box.fresh {
    animation: box-in var(--dur) var(--ease-out) both;
  }

  :global(.anim) .msg.fresh {
    animation: line-in var(--dur) var(--ease-out) both;
  }

  @keyframes box-in {
    from {
      opacity: 0;
      transform: translate(var(--dx), var(--dy));
    }
    to {
      opacity: 1;
      transform: none;
    }
  }

  @keyframes line-in {
    from {
      opacity: 0;
      transform: translate(calc(var(--dx) * 0.5), calc(var(--dy) * 0.5));
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
</style>
