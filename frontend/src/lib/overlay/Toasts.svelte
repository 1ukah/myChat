<script lang="ts">
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
  import { app, clickSpot } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Embed from "./Embed.svelte";
  import { pictureOf, TOAST_FULL, TOAST_THUMB } from "./media";
  import type { Motion } from "./motion";
  import Picture from "./Picture.svelte";
  import RichText from "./RichText.svelte";
  import { nameSpot } from "./spots";

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
    entries: C.Entry[];
    layout: string;
    hotkey: string;
    live: boolean;
    avatars?: boolean;
    images?: boolean;
    minimal?: boolean;
    zoom?: boolean;
    motion: Motion;
  } = $props();

  interface Line {
    entry: C.Entry;
    media: C.Media | null;
    fresh: boolean;
    typed: boolean;
  }

  interface Box {
    key: number;
    name: string;
    avatar: string;
    time: number;
    tint: string;
    who: string;
    fresh: boolean;
    lines: Line[];
  }

  let marks = new Map<number, { enter: "box" | "line" | ""; typed: boolean }>();
  let keys = new Map<number, number>();
  let wasLive = false;
  let lastGrouped: boolean | null = null;

  const boxes = $derived.by((): Box[] => {
    const grouped = layout !== "stack";
    const relaid = lastGrouped !== null && lastGrouped !== grouped;
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

  const fullId = $derived.by(() => {
    for (let i = entries.length - 1; i >= 0; i--) {
      if (pictureOf(entries[i])) return entries[i].id;
    }
    return null;
  });

  const key = $derived((hotkey || "I").replace(/\s*\+\s*/g, "+").toUpperCase());
  const typeDelay = $derived(Math.round(motion.enterMs * 0.6));

  let list = $state<HTMLElement>();
  const before = new Map<HTMLElement, { top: number; bottom: number }>();
  const sliding = new WeakMap<HTMLElement, Animation>();

  const PARTS = ".face, .title, .msg";

  function fit(root: HTMLElement): void {
    for (const el of root.querySelectorAll<HTMLElement>("[hidden]")) el.hidden = false;
    const all = Array.from(root.querySelectorAll<HTMLElement>(".box"));
    for (let b = 0; b < all.length; b++) {
      const box = all[b];
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

  function refit(): void {
    if (list) fit(list);
  }
</script>

<div class="toasts" class:minimal bind:this={list}>
  {#each boxes as box (box.key)}
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
            {#if line.entry.text}
              <div class="text" data-clip>
                <bdi
                  ><RichText
                    text={line.entry.text}
                    at={line.entry.id}
                    links={zoom}
                    typed={line.typed}
                    delay={typeDelay}
                  /></bdi
                >
              </div>
            {/if}
            {#if line.entry.embed}
              <div class="embed">
                <Embed embed={line.entry.embed} at={line.entry.id} {images} {minimal} links={zoom} compact />
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

  .media {
    display: flex;
    margin: 3px 0 2px;
  }

  .media.plain {
    margin: 1px 0;
  }

  .embed {
    display: flex;
    margin: 3px 0 2px;
  }

  .box:global([hidden]),
  .msg:global([hidden]) {
    display: none;
  }

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

  .minimal {
    --link: #c4e6ff;
    --link-hover: #eef8ff;
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
