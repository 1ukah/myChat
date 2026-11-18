<script lang="ts">
  import { tick } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { clickSpot } from "../store.svelte";
  import { fitSize, isLocal } from "./media";
  import { pictureSpot, scheduleSpots } from "./spots";

  let {
    media,
    show,
    maxW,
    maxH,
    minimal = false,
    zoom = false,
    onsize,
  }: {
    media: C.Media;
    show: boolean;
    maxW: number;
    maxH: number;
    minimal?: boolean;
    zoom?: boolean;
    onsize?: () => void;
  } = $props();

  let loaded = $state("");
  let failed = $state("");
  let natural = $state<{ url: string; w: number; h: number } | null>(null);

  const label = $derived(media.kind === "gif" ? "GIF" : "Image");
  const usable = $derived(show && isLocal(media.url) && failed !== media.url);
  const known = $derived(media.width > 0 && media.height > 0);
  const size = $derived.by(() => {
    if (known) return fitSize(media.width, media.height, maxW, maxH);
    const n = natural?.url === media.url ? natural : null;
    return fitSize(n?.w ?? 0, n?.h ?? 0, maxW, maxH);
  });

  function onload(e: Event): void {
    const img = e.currentTarget as HTMLImageElement;
    if (!known) {
      natural = { url: media.url, w: img.naturalWidth, h: img.naturalHeight };
      if (onsize) void tick().then(onsize);
      if (zoom) void tick().then(scheduleSpots);
    }
    loaded = media.url;
  }

  function onerror(): void {
    failed = media.url;
    if (onsize) void tick().then(onsize);
    if (zoom) void tick().then(scheduleSpots);
  }
</script>

{#if usable}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <span
    class="pic"
    class:minimal
    class:zoom
    class:ready={loaded === media.url}
    data-spot={zoom ? pictureSpot(media) : undefined}
    style:width="{size.w}px"
    style:aspect-ratio="{size.w} / {size.h}"
    onclick={zoom ? () => clickSpot(pictureSpot(media)) : undefined}
  >
    <img
      src={media.url}
      alt={media.title || label}
      width={size.w}
      height={size.h}
      decoding="async"
      draggable="false"
      {onload}
      {onerror}
    />
  </span>
{:else}
  <span class="label" class:minimal>
    <svg viewBox="0 0 12 12" width="11" height="11" aria-hidden="true">
      <path d="M1.5 2.5h9v7h-9z M2 8.5l2.6-2.8 2 2 1.4-1.3 2.5 2.4" fill="none" stroke="currentColor" stroke-width="1.1" />
    </svg>
    <span>{label}</span>
  </span>
{/if}

<style>
  .pic {
    position: relative;
    flex: none;
    display: inline-block;
    max-width: calc(100% - 2px);
    margin: 1px;
    overflow: hidden;
    vertical-align: top;
    background: rgba(255, 255, 255, 0.07);
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.85);
  }

  .pic::before {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(
      100deg,
      rgba(255, 255, 255, 0) 30%,
      rgba(255, 255, 255, 0.07) 50%,
      rgba(255, 255, 255, 0) 70%
    );
  }

  .pic.ready::before {
    display: none;
  }

  .pic.zoom {
    cursor: zoom-in;
  }

  img {
    position: relative;
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0;
    user-select: none;
  }

  .ready img {
    opacity: 1;
  }

  .pic.minimal {
    max-width: 100%;
    margin: 0;
    background: rgba(168, 222, 255, 0.09);
    box-shadow: none;
    -webkit-mask-image:
      linear-gradient(to right, transparent 0, #000 3px, #000 calc(100% - 3px), transparent 100%),
      linear-gradient(to bottom, transparent 0, #000 3px, #000 calc(100% - 3px), transparent 100%);
    mask-image:
      linear-gradient(to right, transparent 0, #000 3px, #000 calc(100% - 3px), transparent 100%),
      linear-gradient(to bottom, transparent 0, #000 3px, #000 calc(100% - 3px), transparent 100%);
    -webkit-mask-composite: source-in;
    mask-composite: intersect;
  }

  .pic.minimal::before {
    background: linear-gradient(
      100deg,
      rgba(190, 230, 255, 0) 30%,
      rgba(190, 230, 255, 0.1) 50%,
      rgba(190, 230, 255, 0) 70%
    );
  }

  .label {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    height: 16px;
    font-size: 12px;
    font-weight: 600;
    line-height: 16px;
    letter-spacing: 0.01em;
    vertical-align: top;
    color: var(--muted);
  }

  .label svg {
    flex: none;
    opacity: 0.85;
  }

  .label.minimal {
    font-weight: 400;
    letter-spacing: 0;
    color: #a9b9c8;
  }

  :global(.anim) img {
    transition: opacity 180ms ease-out;
  }

  :global(.anim) .pic:not(.ready)::before {
    animation: wait 1.4s ease-in-out infinite alternate;
  }

  @keyframes wait {
    from {
      opacity: 0.2;
    }
    to {
      opacity: 1;
    }
  }
</style>
