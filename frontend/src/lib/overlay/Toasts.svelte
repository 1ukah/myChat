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
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
  import Avatar from "../ui/Avatar.svelte";
  import type { Motion } from "./motion";
  import Typed from "./Typed.svelte";

  let {
    entries,
    layout,
    hotkey,
    live,
    avatars = false,
    minimal = false,
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
    /** The Minimal look (`overlayStyle`): light, compact, rounded boxes. */
    minimal?: boolean;
    motion: Motion;
  } = $props();

  interface Line {
    entry: C.Entry;
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

      const line: Line = { entry, fresh: mark.enter === "line", typed: mark.typed };
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

  const PARTS = ".face, .title, .text";

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
      if (!old) continue;
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
    if (list && before.size > 0) play(list);
    before.clear();
  });
</script>

<div class="toasts" class:minimal bind:this={list}>
  {#each boxes as box (box.key)}
    <div class="box" class:fresh={box.fresh} style:--who={box.tint}>
      {#if avatars}
        <span class="face">
          <Avatar size={32} tint={box.tint} src={box.avatar} name={box.name} />
        </span>
      {/if}
      <div class="body">
        <div class="title">
          <bdi class="name">{box.name}</bdi>
          <span class="time">{clock(box.time)}</span>
        </div>
        {#each box.lines as line (line.entry.id)}
          <div class="text" class:fresh={line.fresh}>
            <bdi>
              {#if line.typed}
                <Typed text={line.entry.text} delay={typeDelay} />
              {:else}
                {line.entry.text}
              {/if}
            </bdi>
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
     Small frosted-white cards with dark text, a hairline edge, a soft shadow
     and round corners; narrower than Legacy. The person's colour, darkened so
     it reads on white, is kept for the name only. */

  .minimal {
    --width: 400px;
    --avatar-radius: 4px;
    --muted: #6f7b8a;
    --tile: #66758a;
    font-family: var(--font-body);
  }

  .minimal .box {
    gap: 8px;
    padding: 5px 12px 6px 8px;
    background: linear-gradient(rgba(255, 255, 255, 0.93), rgba(243, 247, 252, 0.83));
    border-radius: 11px;
    box-shadow:
      inset 0 0 0 1px rgba(255, 255, 255, 0.8),
      0 0 0 1px rgba(24, 38, 62, 0.1),
      0 3px 10px rgba(14, 24, 44, 0.18);
  }

  .minimal .face {
    margin: 2px 0 0;
  }

  .minimal .name {
    letter-spacing: 0;
    color: color-mix(in srgb, var(--who) 46%, #0d1b33);
    text-shadow: none;
  }

  .minimal .time {
    font-weight: 400;
  }

  .minimal .text {
    font-weight: 400;
    color: #151c26;
    text-shadow: none;
  }

  .minimal .hint {
    height: 17px;
  }

  .minimal .hint span {
    padding: 0 8px;
    font-weight: 400;
    line-height: 17px;
    letter-spacing: 0;
    color: #4b5767;
    background: rgba(255, 255, 255, 0.84);
    border-radius: 9px;
    box-shadow: 0 0 0 1px rgba(24, 38, 62, 0.08);
  }

  .minimal .hint b {
    font-weight: 700;
    color: #151c26;
  }

  /* A new message moves in the way the chosen style says (--dx/--dy). */
  :global(.anim) .box.fresh {
    animation: box-in var(--dur) var(--ease-out) both;
  }

  :global(.anim) .text.fresh {
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
