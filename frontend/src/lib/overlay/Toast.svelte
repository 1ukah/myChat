<script lang="ts">
  // One incoming message (640×92), in the style of CS2's round banner:
  // a grey dotted banner with light-blue edges and » « marks that carries the
  // sender's name, and under it the blue pixel-mosaic "MVP" strip with the
  // square avatar, a tag and the message. A newer toast replaces the one on
  // screen with a quick cross-fade.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
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

  /** "Ctrl + F6" → "CTRL+F6", as CS2 writes key hints. */
  const key = $derived((hotkey || "I").replace(/\s*\+\s*/g, "+").toUpperCase());
</script>

<div class="toast">
  {#each cards as card (card.entry.id)}
    <div class="card" class:out={card.out} class:swap={card.swap}>
      <div class="banner">
        <span class="mark">»</span>
        <bdi class="name">{card.entry.peerName}</bdi>
        <span class="mark">«</span>
      </div>
      <div class="strip">
        {#if avatars}
          <Avatar
            size={40}
            tint={playerColor(card.entry.peerId || card.entry.peerName)}
            src={card.entry.peerAvatar}
            name={card.entry.peerName}
          />
        {/if}
        <div class="text">
          <bdi>
            {#if typewriter && !card.out}
              <Typed text={card.entry.text} delay={typeDelay} />
            {:else}
              {card.entry.text}
            {/if}
          </bdi>
        </div>
        <div class="side">
          <span class="tag">Whisper</span>
          <span class="reply">[{key}] Reply · {clock(card.entry.time)}</span>
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
    font-family: var(--font-head);
  }

  .card {
    position: absolute;
    inset: 0;
  }

  /* ---------- the banner: "»  R O U N D   W O N  «" ---------- */

  .banner {
    position: absolute;
    top: 0;
    right: 88px;
    left: 88px;
    display: flex;
    align-items: center;
    gap: 12px;
    height: 34px;
    padding: 0 14px;
    color: var(--banner);
    background:
      var(--dots),
      rgba(62, 64, 68, 0.84);
    /* Light-blue bars close the banner on both sides. */
    box-shadow:
      inset 2px 0 0 var(--banner),
      inset -2px 0 0 var(--banner);
  }

  .mark {
    flex: none;
    font-size: 20px;
    font-weight: 700;
    line-height: 34px;
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    /* The tracking adds space after the last letter; this re-centres. */
    padding-left: 0.2em;
    font-size: 22px;
    font-weight: 600;
    line-height: 34px;
    letter-spacing: 0.2em;
    text-align: center;
    text-overflow: ellipsis;
    text-transform: uppercase;
    white-space: nowrap;
  }

  /* ---------- the strip: blue pixel mosaic, soft at both ends ---------- */

  .strip {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    align-items: center;
    gap: 12px;
    height: 54px;
    padding: 0 44px;
    background:
      var(--mosaic),
      rgba(29, 91, 136, 0.9);
    -webkit-mask-image: linear-gradient(to right, transparent 0, #000 36px, #000 calc(100% - 36px), transparent 100%);
    mask-image: linear-gradient(to right, transparent 0, #000 36px, #000 calc(100% - 36px), transparent 100%);
  }

  .text {
    flex: 1;
    min-width: 0;
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
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.55);
  }

  .side {
    flex: none;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
    white-space: nowrap;
  }

  /* The light pill that says "MVP". */
  .tag {
    height: 17px;
    padding: 0 6px;
    font-size: 12px;
    font-weight: 700;
    line-height: 17px;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: #123a58;
    background: #bfe0fb;
    border-radius: 2px;
  }

  .reply {
    font-size: 12px;
    font-weight: 600;
    line-height: 15px;
    font-variant-numeric: tabular-nums;
    color: #cfe6f8;
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
