<script lang="ts">
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clock, playerColor } from "../format";
  import { app, clickSpot, isFresh } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Embed from "./Embed.svelte";
  import { CHAT_THUMB, pictureOf } from "./media";
  import Picture from "./Picture.svelte";
  import RichText from "./RichText.svelte";
  import { nameSpot } from "./spots";

  let {
    entry,
    avatars = false,
    images = true,
    thumbH = CHAT_THUMB.h,
    minimal = false,
    typewriter = false,
  }: {
    entry: C.Entry;
    avatars?: boolean;
    images?: boolean;
    thumbH?: number;
    minimal?: boolean;
    typewriter?: boolean;
  } = $props();

  const fresh = untrack(() => isFresh(entry.id));
  const typed = untrack(() => fresh && typewriter && entry.kind === "in");

  const kind = $derived(entry.kind);
  const pending = $derived(entry.status === "pending");
  const failed = $derived(entry.status === "failed");
  const tint = $derived(playerColor(entry.peerId || entry.peerName));
  const pic = $derived(pictureOf(entry));
  const spot = $derived(nameSpot(entry.peerId, entry.id));
  const hover = $derived(spot !== "" && app.hover === spot);
</script>

{#snippet face()}{#if avatars}<Avatar inline size={16} {tint} src={entry.peerAvatar} name={entry.peerName} />{:else}<span class="dot" style:color={tint}>{minimal ? "■" : "●"}</span>{" "}{/if}{/snippet}

{#snippet name()}<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions --><bdi class="name" class:link={spot !== ""} class:hover data-spot={spot || undefined} onclick={spot ? () => clickSpot(spot) : undefined}>{entry.peerName}</bdi>{/snippet}

<div class="line {kind}" class:minimal class:pending class:fresh>
  <span class="time">{clock(entry.time)}</span>
  <span class="msg">{#if kind === "in"}<span class="who">{minimal ? "" : "[FROM] "}{@render face()}{@render name()}</span>{": "}{:else if kind === "out"}<span class="who">{minimal ? "To " : "[TO] "}{@render face()}{@render name()}</span>{": "}{/if}<bdi class="text"><RichText text={entry.text} at={entry.id} {typed} /></bdi></span>
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
  {#if pic}
    <div class="media" class:plain={!images}><span class="time ghost" aria-hidden="true">{clock(entry.time)}</span>{" "}<Picture media={pic} show={images} maxW={CHAT_THUMB.w} maxH={thumbH} {minimal} zoom /></div>
  {/if}
  {#if entry.embed}
    <div class="embed"><span class="time ghost" aria-hidden="true">{clock(entry.time)}</span><Embed embed={entry.embed} at={entry.id} {images} {minimal} /></div>
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

  .link {
    cursor: pointer;
  }

  .link.hover,
  .link:hover {
    color: color-mix(in srgb, var(--c) 65%, #fff);
    text-decoration: underline;
    text-decoration-thickness: 1px;
    text-underline-offset: 2px;
  }

  .media {
    margin: 2px 0 4px;
    line-height: 0;
    white-space: nowrap;
  }

  .media.plain {
    margin: 0 0 2px;
  }

  .embed {
    display: flex;
    gap: 0.25em;
    margin: 2px 0 4px;
  }

  .embed .ghost {
    flex: none;
  }

  .ghost {
    visibility: hidden;
  }

  .dot {
    font-size: 11px;
    vertical-align: 1px;
  }

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

  .minimal {
    --link: #c4e6ff;
    --link-hover: #eef8ff;
    font-family: var(--font-ui);
    font-size: 14px;
    font-weight: 400;
    line-height: 20px;
    color: #f4f8fb;
    text-shadow: 0 1px 1px rgba(4, 10, 18, 0.55), 0 0 5px rgba(178, 226, 255, 0.5);
  }

  .minimal.in {
    --c: #a3d6ff;
  }

  .minimal.out {
    --c: #b9c8d6;
  }

  .minimal .who {
    font-weight: 700;
  }

  .minimal.out .text {
    color: #dbe5ee;
  }

  .minimal.system {
    font-weight: 400;
    color: #b7c6d4;
  }

  .minimal.error {
    font-weight: 600;
    color: #ff9d94;
  }

  .minimal .time {
    font-size: 11px;
    font-weight: 400;
    color: #a9b9c8;
  }

  .minimal .fail {
    font-size: 12px;
    font-weight: 600;
    color: #ff9d94;
  }

  :global(.anim) .fresh {
    animation: line-in calc(var(--dur) * 0.75) var(--ease-out) both;
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
