<script lang="ts">
  // One incoming message (640×92), laid out like a spectator player card:
  // ringed circle avatar, condensed name, the message. A newer toast replaces
  // the one on screen with a quick cross-fade.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock } from "../format";
  import Avatar from "../ui/Avatar.svelte";
  import Typed from "./Typed.svelte";

  let {
    entry,
    hotkey,
    live,
    avatars = false,
    typewriter = false,
    typeDelay = 0,
  }: {
    entry: C.Entry;
    /** Label of the open-chat hotkey, e.g. "I" or "Ctrl + F6". */
    hotkey: string;
    /** The toast is on screen (not fading out, not hidden). */
    live: boolean;
    /** Show the sender's avatar (setting `showAvatars`). */
    avatars?: boolean;
    /** Animation style is "typewriter": the text types itself out. */
    typewriter?: boolean;
    /** Typing starts this many ms in, once the panel has mostly faded in. */
    typeDelay?: number;
  } = $props();

  interface Card {
    entry: C.Entry;
    /** Fading out because a newer message replaced it. */
    out: boolean;
    /** Came in over a previous card, so it moves in. */
    swap: boolean;
  }

  // Plain memory of what was shown, so a change of `entry` can tell whether
  // there is a card on screen to cross-fade from. The old card stays in the
  // DOM (invisible) until the next change: no timers needed.
  let current: C.Entry | null = null;
  let previous: C.Entry | null = null;

  const cards = $derived.by((): Card[] => {
    const next = entry;
    if (current && current.id !== next.id) {
      previous = untrack(() => live) ? current : null;
    }
    current = next;
    return previous
      ? [
          { entry: previous, out: true, swap: false },
          { entry: next, out: false, swap: true },
        ]
      : [{ entry: next, out: false, swap: false }];
  });
</script>

<div class="toast">
  {#each cards as card (card.entry.id)}
    <div class="card" class:out={card.out} class:swap={card.swap}>
      {#if avatars}
        <Avatar size={44} ring="whisper" src={card.entry.peerAvatar} name={card.entry.peerName} />
      {/if}
      <div class="body">
        <div class="top">
          <bdi class="name">{card.entry.peerName}</bdi>
          <svg class="mail" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            <path
              d="M1.5 3.5h13v9h-13z M1.5 4l6.5 5 6.5-5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            />
          </svg>
          <span class="reply">Press <kbd>{hotkey || "I"}</kbd> to reply</span>
          <span class="time">{clock(card.entry.time)}</span>
        </div>
        <div class="text">
          <bdi>
            {#if typewriter && !card.out}
              <Typed text={card.entry.text} delay={typeDelay} />
            {:else}
              {card.entry.text}
            {/if}
          </bdi>
        </div>
      </div>
    </div>
  {/each}
</div>

<style>
  /* Pinned to the bottom edge: the window grows upwards when Go switches
     to the composer, and the card must not jump while that happens. */
  .toast {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    height: 92px;
    box-shadow: inset 3px 0 0 var(--whisper);
  }

  .card {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 0 16px 0 21px;
  }

  .body {
    flex: 1;
    min-width: 0;
  }

  .top {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 20px;
    white-space: nowrap;
  }

  .name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--font-head);
    font-size: 19px;
    font-weight: 700;
    line-height: 20px;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--whisper);
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  }

  .mail {
    flex: none;
    color: var(--whisper);
    opacity: 0.8;
  }

  .reply {
    flex: none;
    display: flex;
    align-items: center;
    gap: 5px;
    margin-left: auto;
    padding-left: 8px;
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 500;
    line-height: 20px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--muted);
  }

  /* Boxed key, as in the scoreboard footer. */
  .reply kbd {
    min-width: 16px;
    height: 16px;
    padding: 0 5px;
    font: inherit;
    font-size: 12px;
    font-weight: 700;
    line-height: 16px;
    text-align: center;
    color: var(--text);
    background: rgba(255, 255, 255, 0.08);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.34);
  }

  .time {
    flex: none;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    color: var(--muted);
  }

  .text {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    overflow: hidden;
    margin-top: 4px;
    font-size: 14px;
    line-height: 19px;
    overflow-wrap: anywhere;
    color: var(--text);
    text-shadow:
      0 1px 2px rgba(0, 0, 0, 0.85),
      0 0 1px rgba(0, 0, 0, 0.6);
  }

  /* Cross-fade between two messages, in the direction of the chosen style. */
  .card.out {
    opacity: 0;
  }

  :global(.anim) .card.out {
    animation: card-out calc(var(--dur) * 0.7) ease-in both;
  }

  :global(.anim) .card.swap {
    animation: card-in var(--dur) var(--ease-out) both;
  }

  @keyframes card-in {
    from {
      opacity: 0;
      transform: translate(calc(var(--dx) * 0.6), calc(var(--dy) * 0.6));
    }
    to {
      opacity: 1;
      transform: none;
    }
  }

  @keyframes card-out {
    from {
      opacity: 1;
      transform: none;
    }
    to {
      opacity: 0;
      transform: translate(calc(var(--dx) * -0.6), calc(var(--dy) * -0.6));
    }
  }
</style>
