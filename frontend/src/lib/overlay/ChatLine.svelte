<script lang="ts">
  // One chat log entry, World of Warcraft style:
  //   [12:41] (o) [SANT1] whispers: text
  //   [12:41] To (o) [SANT1]: text
  //   [12:41] [Party] You: text
  // (o) is the friend's avatar, when avatars are on.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock } from "../format";
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
</script>

{#snippet face()}{#if avatars}<Avatar inline size={18} ring="whisper" src={entry.peerAvatar} name={entry.peerName} />{/if}{/snippet}

<!-- String literals carry the spaces and nothing inside .msg is split over
     lines, so the layout never depends on how template whitespace collapses.
     <bdi> keeps right-to-left names and messages from reordering the brackets. -->
<div class="line {kind}" class:pending class:fresh>
  <span class="time">[{clock(entry.time)}]</span>
  <span class="msg">{#if kind === "in"}{@render face()}<span class="who">[<bdi>{entry.peerName}</bdi>]</span>{" whispers: "}{:else if kind === "out"}{"To "}{@render face()}<span class="who">[<bdi>{entry.peerName}</bdi>]</span>{": "}{:else if kind === "party"}{#if avatars}<svg class="group" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true"><path d="M5.5 7a2.25 2.25 0 1 0 0-4.5 2.25 2.25 0 0 0 0 4.5z M1.5 13c.3-2.4 1.9-3.6 4-3.6s3.7 1.2 4 3.6 M10.4 6.9a2 2 0 1 0-.5-3.8 M11.2 9.5c1.8.3 3 1.5 3.3 3.5" /></svg>{/if}<span class="who">[Party]</span>{" You: "}{/if}<bdi>{#if typed}<Typed text={entry.text} />{:else}{entry.text}{/if}</bdi></span>
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
    font-size: 15px;
    line-height: 21px;
    overflow-wrap: anywhere;
    color: var(--c);
    /* Keeps thin coloured text legible over bright scenes. */
    text-shadow:
      0 1px 2px rgba(0, 0, 0, 0.85),
      0 0 1px rgba(0, 0, 0, 0.6);
  }

  .in {
    --c: var(--whisper);
  }

  .out {
    --c: var(--whisper-out);
  }

  .party {
    --c: var(--party);
  }

  .system {
    --c: #a7aeb5;
    font-style: italic;
  }

  .error {
    --c: var(--red-text);
  }

  .time {
    font-size: 12px;
    font-style: normal;
    font-variant-numeric: tabular-nums;
    color: var(--muted);
  }

  .who {
    font-weight: 600;
  }

  .group {
    margin-right: 4px;
    vertical-align: -2px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.4;
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
    font-size: 12px;
    font-style: normal;
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
