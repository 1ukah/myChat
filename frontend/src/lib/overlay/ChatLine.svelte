<script lang="ts">
  // One chat log entry, written like CS2's own chat ("[CT] ● name: text"):
  //   12:41 [FROM] (o) sant1: text
  //   12:41 [TO] (o) sant1: text
  // (o) is the friend's square avatar with a player-colour border, or a
  // player-colour dot when avatars are off.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
  import { isFresh } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Typed from "./Typed.svelte";

  let {
    entry,
    avatars = false,
    typewriter = false,
  }: {
    entry: C.Entry;
    avatars?: boolean;
    /** Animation style is "typewriter": a new incoming message types itself out. */
    typewriter?: boolean;
  } = $props();

  // Decided once, when the line is created: only a message that just arrived
  // animates. Lines that come back into view by scrolling do not.
  const fresh = untrack(() => isFresh(entry.id));
  const typed = untrack(() => fresh && typewriter && entry.kind === "in");

  const kind = $derived(entry.kind);
  const pending = $derived(entry.status === "pending");
  const failed = $derived(entry.status === "failed");
  const tint = $derived(playerColor(entry.peerId || entry.peerName));
</script>

{#snippet face()}{#if avatars}<Avatar inline size={16} {tint} src={entry.peerAvatar} name={entry.peerName} />{:else}<span class="dot" style:color={tint}>●</span>{" "}{/if}{/snippet}

<!-- String literals carry the spaces and nothing inside .msg is split over
     lines, so the layout never depends on how template whitespace collapses.
     <bdi> keeps right-to-left names and messages from reordering the brackets. -->
<div class="line {kind}" class:pending class:fresh>
  <span class="time">{clock(entry.time)}</span>
  <span class="msg">{#if kind === "in"}<span class="who">{"[FROM] "}{@render face()}<bdi>{entry.peerName}</bdi></span>{": "}{:else if kind === "out"}<span class="who">{"[TO] "}{@render face()}<bdi>{entry.peerName}</bdi></span>{": "}{/if}<bdi class="text">{#if typed}<Typed text={entry.text} />{:else}{entry.text}{/if}</bdi></span>
  {#if pending}
    <span class="dots" role="img" aria-label="sending"><i></i><i></i><i></i></span>
  {:else if failed}
    <span class="fail">
      <svg viewBox="0 0 10 10" width="10" height="10" aria-hidden="true">
        <path d="M1.5 1.5l7 7M8.5 1.5l-7 7" fill="none" stroke="currentColor" stroke-width="2" />
      </svg>
      <span>{entry.detail || "Not sent"}</span>
    </span>
  {/if}
</div>

<style>
  .line {
    --c: var(--text);
    flex: none;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 700;
    line-height: 21px;
    overflow-wrap: anywhere;
    color: #fff;
    /* CS2 chat: a hard 1 px shadow under bold text. */
    text-shadow:
      1px 1px 0 rgba(0, 0, 0, 0.9),
      0 0 2px rgba(0, 0, 0, 0.7);
  }

  .in {
    --c: var(--whisper);
  }

  .out {
    --c: var(--whisper-out);
  }

  .out .text {
    color: #dfe4e9;
  }

  .system {
    color: #c5ccd3;
    font-weight: 600;
  }

  .error {
    color: var(--red-text);
  }

  .time {
    margin-right: 2px;
    font-size: 12px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--muted);
  }

  .who {
    color: var(--c);
  }

  .dot {
    font-size: 11px;
    vertical-align: 1px;
  }

  /* Dim only the message, so the "sending" dots stay clear. */
  .pending .msg {
    opacity: 0.6;
  }

  .dots {
    display: inline-flex;
    gap: 3px;
    margin-left: 2px;
    vertical-align: 2px;
  }

  .dots i {
    width: 3px;
    height: 3px;
    background: var(--c);
    opacity: 0.7;
  }

  .fail {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    margin-left: 4px;
    font-size: 13px;
    color: var(--red-text);
  }

  .fail svg {
    flex: none;
  }

  /* A new line uses a shorter, smaller version of the panel's motion
     (--dx/--dy come from the animation style). Motion only when enabled. */
  :global(.anim) .fresh {
    animation: line-in calc(var(--dur) * 0.75) var(--ease-out) both;
  }

  /* The dots pulse only while the composer is open. */
  :global(.anim .live) .dots i {
    animation: dot 1.2s ease-in-out infinite;
  }

  :global(.anim .live) .dots i:nth-child(2) {
    animation-delay: 0.15s;
  }

  :global(.anim .live) .dots i:nth-child(3) {
    animation-delay: 0.3s;
  }

  @keyframes line-in {
    from {
      opacity: 0;
      transform: translate(calc(var(--dx) * 0.4), calc(var(--dy) * 0.4));
    }
    to {
      opacity: 1;
      transform: none;
    }
  }

  @keyframes dot {
    0%,
    60%,
    100% {
      opacity: 0.25;
    }
    30% {
      opacity: 1;
    }
  }
</style>
