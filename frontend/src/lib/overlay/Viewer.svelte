<script lang="ts">
  // A picture opened at full size. While it is open Go makes the overlay
  // window cover the whole monitor: the screen dims, the picture sits in
  // the middle, as large as fits (small pictures grow up to 2×), and the
  // chat box or pop-ups stay where they were, under the dim.
  //
  // Any click closes it (Go catches the click), and Esc while the chat box
  // is open. `onclick` is for the browser mock.
  //
  // With animations on, the picture grows out of its thumbnail into the
  // middle while the dim fades in, and shrinks back into it on close.
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { closePicture } from "../store.svelte";
  import { isLocal } from "./media";
  import type { Motion } from "./motion";
  import { pictureSpot } from "./spots";

  let {
    media,
    motion,
    closing = false,
    minimal = false,
    settled = true,
    composer = false,
  }: {
    media: C.Media;
    motion: Motion;
    /** Closed: the closing animation plays, then the overlay removes it. */
    closing?: boolean;
    /** The Minimal look. */
    minimal?: boolean;
    /** The window has drawn at its current size and Go shows it again. */
    settled?: boolean;
    /** The chat box is open, so Esc closes the picture too. */
    composer?: boolean;
  } = $props();

  /** Small pictures are enlarged up to this much. */
  const GROW = 2;
  /** The zoom gets a little longer than the chat box's entry. */
  const openMs = $derived(Math.round(motion.enterMs * 1.5));
  const EASE = "cubic-bezier(0.2, 0.8, 0.2, 1)";

  let vw = $state(0);
  let vh = $state(0);
  let natural = $state<{ url: string; w: number; h: number } | null>(null);
  let failed = $state("");
  /** The original (`media.full`) once it has loaded, or failed to. */
  let fullLoaded = $state("");
  let fullFailed = $state("");

  const usable = $derived(isLocal(media.url) && failed !== media.url);
  /** The thumbnail shows at once; the original, fetched only now, covers
      it once loaded. Without one, the thumbnail is all there is. */
  const hasFull = $derived(!!media.full && isLocal(media.full) && fullFailed !== media.full);
  const fullReady = $derived(hasFull && fullLoaded === media.full);
  const src = $derived.by(() => {
    if (media.width > 0 && media.height > 0) return { w: media.width, h: media.height };
    return natural?.url === media.url ? natural : null;
  });
  /** Unknown until the image has loaded: then it shows. */
  const size = $derived.by(() => {
    if (!src) return null;
    const k = Math.min(GROW, Math.max(0, vw - 96) / src.w, Math.max(0, vh - 136) / src.h);
    return { w: Math.max(1, Math.round(src.w * k)), h: Math.max(1, Math.round(src.h * k)) };
  });
  const label = $derived(media.kind === "gif" ? "GIF" : "Image");

  function onload(e: Event): void {
    const img = e.currentTarget as HTMLImageElement;
    natural = { url: media.url, w: img.naturalWidth, h: img.naturalHeight };
  }

  // Go makes the window cover the monitor just before this mounts, and the
  // window may still be its usual 640×360 for a moment. Nothing shows or
  // animates until it has grown and is on screen again (`settled`).
  const big = $derived(vh > 361 && settled);

  let frame = $state<HTMLElement>();
  let opened = false;

  /** The transform that puts the frame (at rest) over the thumbnail, or
      null when the thumbnail is not on screen. */
  function overThumbnail(el: HTMLElement): string | null {
    const thumb = document.querySelector(`[data-spot="${CSS.escape(pictureSpot(media))}"]`);
    if (!thumb) return null;
    const t = thumb.getBoundingClientRect();
    const f = el.getBoundingClientRect();
    if (t.width < 1 || t.height < 1 || f.width < 1 || f.height < 1) return null;
    const k = Math.min(t.width / f.width, t.height / f.height);
    const dx = t.left + t.width / 2 - (f.left + f.width / 2);
    const dy = t.top + t.height / 2 - (f.top + f.height / 2);
    return `translate(${dx}px, ${dy}px) scale(${k})`;
  }

  $effect(() => {
    if (!big || !size || !frame || opened) return;
    opened = true;
    if (!motion.on) return;
    const from = overThumbnail(frame);
    frame.animate(
      from
        ? [{ transform: from }, { transform: "none" }]
        : [{ transform: "scale(0.85)", opacity: 0 }, { transform: "none", opacity: 1 }],
      { duration: openMs, easing: EASE },
    );
  });

  $effect(() => {
    if (!closing || !frame || !motion.on || !opened) return;
    const el = frame;
    // Carry on from wherever the opening zoom is.
    const was = getComputedStyle(el).transform;
    for (const a of el.getAnimations()) a.cancel();
    const to = overThumbnail(el);
    el.animate(
      to
        ? [
            { transform: was, opacity: 1 },
            { transform: to, opacity: 1, offset: 0.7 },
            { transform: to, opacity: 0 },
          ]
        : [{ transform: was, opacity: 1 }, { transform: "scale(0.85)", opacity: 0 }],
      { duration: motion.exitMs, easing: EASE, fill: "forwards" },
    );
  });
</script>

<!-- The overlay never has keyboard focus, and Go handles the real clicks. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
  class="viewer"
  class:minimal
  class:open={big}
  class:closing
  style:--zoom="{openMs}ms"
  bind:clientWidth={vw}
  bind:clientHeight={vh}
  onclick={closePicture}
>
  <div class="dim"></div>
  <figure>
    {#if usable}
      <div class="frame" class:ready={size !== null} bind:this={frame}>
        <img
          src={media.url}
          alt={media.title || label}
          width={size?.w}
          height={size?.h}
          draggable="false"
          {onload}
          onerror={() => (failed = media.url)}
        />
        {#if hasFull}
          <img
            class="full"
            class:shown={fullReady}
            src={media.full}
            alt=""
            width={size?.w}
            height={size?.h}
            draggable="false"
            onload={() => (fullLoaded = media.full)}
            onerror={() => (fullFailed = media.full)}
          />
        {/if}
      </div>
    {:else}
      <div class="missing">{label} could not be loaded</div>
    {/if}
    <figcaption>
      {#if media.title}<bdi class="title">{media.title}</bdi>{/if}
      <span class="hint">
        {#if minimal}
          <b>Click</b>{#if composer}{" or "}<b>Esc</b>{/if} to close
        {:else}
          [CLICK]{#if composer}{" / [ESC]"}{/if} Close
        {/if}
      </span>
    </figcaption>
  </figure>
</div>

<style>
  .viewer {
    position: fixed;
    z-index: 10;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: zoom-out;
    font-family: var(--font-head);
  }

  /* Until the window covers the monitor (see `big`). */
  .viewer:not(.open) {
    visibility: hidden;
  }

  .dim {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.62);
  }

  figure {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    margin: 0;
  }

  /* Legacy: a thin dark frame and a deep shadow, like the thumbnails. */
  .frame {
    position: relative;
    display: flex;
    background: rgba(255, 255, 255, 0.05);
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.85),
      0 12px 48px rgba(0, 0, 0, 0.6);
    opacity: 0;
  }

  .frame.ready {
    opacity: 1;
  }

  img {
    display: block;
    user-select: none;
  }

  img.full {
    position: absolute;
    inset: 0;
    opacity: 0;
  }

  img.full.shown {
    opacity: 1;
  }

  /* Until its size is known the picture is not drawn (it is never shown
     at a size it will not keep). */
  .frame:not(.ready) img {
    position: absolute;
    visibility: hidden;
  }

  .missing {
    padding: 12px 18px;
    font-size: 15px;
    font-weight: 600;
    color: var(--muted);
    background: rgba(0, 0, 0, 0.6);
  }

  figcaption {
    display: flex;
    align-items: baseline;
    gap: 14px;
    max-width: 90vw;
    font-size: 12px;
    font-weight: 700;
    line-height: 20px;
    white-space: nowrap;
    color: #e6e9ec;
    text-shadow:
      1px 1px 0 #000,
      0 0 3px rgba(0, 0, 0, 0.9);
  }

  .title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
    color: #fff;
  }

  .hint {
    flex: none;
    padding: 0 6px;
    letter-spacing: 0.02em;
    color: #c5ccd3;
    background: rgba(0, 0, 0, 0.45);
  }

  /* Minimal: a cool dark dim, the lit hairline and a soft cyan glow
     around the picture, and the hint on a strip of glass. */
  .minimal {
    font-family: var(--font-ui);
  }

  .minimal .dim {
    background: rgba(8, 14, 22, 0.62);
  }

  .minimal .frame {
    background: rgba(168, 222, 255, 0.09);
    box-shadow:
      0 0 0 1px rgba(190, 230, 255, 0.35),
      0 0 24px rgba(120, 196, 255, 0.28),
      0 12px 48px rgba(0, 0, 0, 0.55);
  }

  .minimal figcaption {
    font-weight: 400;
    color: #c9d6e2;
    text-shadow: 0 1px 1px rgba(4, 10, 18, 0.55), 0 0 5px rgba(178, 226, 255, 0.5);
  }

  .minimal .title {
    font-weight: 600;
    color: #f4f8fb;
  }

  .minimal .hint {
    padding: 0 14px;
    letter-spacing: 0;
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

  /* Opacity only, and only with animations on. The picture's zoom is
     played from the script, since it starts at the thumbnail. */
  :global(.anim) .viewer.open .dim,
  :global(.anim) .viewer.open figcaption {
    animation: fade-in var(--zoom) ease-out both;
  }

  :global(.anim) .viewer.closing .dim,
  :global(.anim) .viewer.closing figcaption {
    animation: fade-out var(--out) ease-in both;
  }

  @keyframes fade-in {
    from {
      opacity: 0;
    }
  }

  @keyframes fade-out {
    to {
      opacity: 0;
    }
  }
</style>
