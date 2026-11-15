<script lang="ts">
  // One chat log entry, World of Warcraft style:
  //   [12:41] [SANT1] whispers: text
  //   [12:41] To [SANT1]: text
  //   [12:41] [Party] You: text
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock } from "../format";
  import { isFresh } from "../store.svelte";

  let { entry }: { entry: C.Entry } = $props();

  // Decided once, when the line is created: only a message that just arrived
  // slides in. Lines that come back into view by scrolling do not.
  const fresh = untrack(() => isFresh(entry.id));

  const kind = $derived(entry.kind);
  const pending = $derived(entry.status === "pending");
  const failed = $derived(entry.status === "failed");
</script>

<!-- String literals carry the spaces, so the layout never depends on how
     template whitespace is collapsed. <bdi> keeps right-to-left names and
     messages from reordering the brackets around them. -->
<div class="line {kind}" class:pending class:fresh>
  <span class="time">[{clock(entry.time)}]</span>
  <span class="msg"
    >{#if kind === "in"}<span class="who">[<bdi>{entry.peerName}</bdi>]</span>{" whispers: "}{:else if kind === "out"}{"To "}<span
        class="who">[<bdi>{entry.peerName}</bdi>]</span
      >{": "}{:else if kind === "party"}<span class="who">[Party]</span>{" You: "}{/if}<bdi>{entry.text}</bdi></span
  >
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

  /* Motion only when enabled, and only while the composer is open. */
  :global(.anim) .fresh {
    animation: line-in 160ms var(--ease-out) both;
  }

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
      transform: translateY(5px);
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
