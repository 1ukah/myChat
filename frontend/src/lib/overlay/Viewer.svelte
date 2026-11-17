<script lang="ts">
  // A picture opened at full size. While it is open Go makes the overlay
  // window cover the whole monitor: the screen dims, the picture sits in
  // the middle, as large as fits (small pictures grow up to 2×), and the
  // chat box or pop-ups stay where they were, under the dim.
  //
  // Any click closes it (Go catches the click), and Esc while the chat box
  // is open. `onclick` is for the browser mock.
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { closePicture } from "../store.svelte";
  import { isLocal } from "./media";

  let {
    media,
    minimal = false,
    composer = false,
  }: {
    media: C.Media;
    /** The Minimal look. */
    minimal?: boolean;
    /** The chat box is open, so Esc closes the picture too. */
    composer?: boolean;
  } = $props();

  /** Small pictures are enlarged up to this much. */
  const GROW = 2;

  let vw = $state(0);
  let vh = $state(0);
  let natural = $state<{ url: string; w: number; h: number } | null>(null);
  let failed = $state("");

  const usable = $derived(isLocal(media.url) && failed !== media.url);
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
</script>

<!-- The overlay never has keyboard focus, and Go handles the real clicks. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
  class="viewer"
  class:minimal
  bind:clientWidth={vw}
  bind:clientHeight={vh}
  onclick={closePicture}
>
  <figure>
    {#if usable}
      <div class="frame" class:ready={size !== null}>
        <img
          src={media.url}
          alt={media.title || label}
          width={size?.w}
          height={size?.h}
          draggable="false"
          {onload}
          onerror={() => (failed = media.url)}
        />
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
    background: rgba(0, 0, 0, 0.62);
    cursor: zoom-out;
    font-family: var(--font-head);
  }

  figure {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    margin: 0;
  }

  /* Legacy: a thin dark frame and a deep shadow, like the thumbnails. */
  .frame {
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
    background: rgba(8, 14, 22, 0.62);
    font-family: var(--font-ui);
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

  /* Opacity and transform only, and only with animations on. */
  :global(.anim) .viewer {
    animation: dim-in var(--dur) ease-out both;
  }

  :global(.anim) .frame.ready {
    animation: grow-in var(--dur) var(--ease-out) both;
  }

  @keyframes dim-in {
    from {
      opacity: 0;
    }
  }

  @keyframes grow-in {
    from {
      opacity: 0;
      transform: scale(0.96);
    }
  }
</style>
