<script lang="ts">
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { app, clickSpot } from "../store.svelte";
  import { isLocal } from "./media";
  import { linkSpot } from "./spots";

  let {
    embed,
    at,
    images = true,
    minimal = false,
    links = true,
    compact = false,
  }: {
    embed: C.Embed;
    at: number;
    images?: boolean;
    minimal?: boolean;
    links?: boolean;
    compact?: boolean;
  } = $props();

  const SITE_COLORS: Record<string, string> = { youtube: "#ff4e45", x: "#e7e9ea", reddit: "#ff4500" };

  const spot = $derived(links ? linkSpot(at, "card") : "");
  const thumb = $derived(images && embed.thumb && isLocal(embed.thumb.url) ? embed.thumb : null);
  const box = $derived(compact ? { w: 96, h: 54 } : { w: 112, h: 63 });
  const head = $derived([embed.siteName, embed.author].filter(Boolean).join(" · "));

  let loaded = $state("");
  let failed = $state("");
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
  class="card"
  class:minimal
  class:compact
  class:link={spot !== ""}
  class:hover={spot !== "" && app.hover === spot}
  style:--site={SITE_COLORS[embed.site] ?? "var(--muted)"}
  data-spot={spot || undefined}
  data-url={spot ? embed.url : undefined}
  onclick={spot ? () => clickSpot(spot) : undefined}
>
  <div class="info">
    {#if head}<div class="from"><bdi>{head}</bdi></div>{/if}
    {#if embed.title}<div class="title"><bdi>{embed.title}</bdi></div>{/if}
    {#if embed.text}<div class="desc"><bdi>{embed.text}</bdi></div>{/if}
  </div>
  {#if thumb && failed !== thumb.url}
    <span class="thumb" class:ready={loaded === thumb.url} style:width="{box.w}px" style:height="{box.h}px">
      <img
        src={thumb.url}
        alt=""
        width={box.w}
        height={box.h}
        decoding="async"
        draggable="false"
        onload={() => (loaded = thumb.url)}
        onerror={() => (failed = thumb.url)}
      />
      {#if embed.video}
        <span class="play" aria-hidden="true">
          <svg viewBox="0 0 10 10" width="10" height="10"><path d="M2.5 1.5v7l6-3.5z" fill="currentColor" /></svg>
        </span>
      {/if}
    </span>
  {/if}
</div>

<style>
  .card {
    display: flex;
    gap: 10px;
    width: 100%;
    max-width: 420px;
    min-width: 0;
    padding: 6px 8px 6px 10px;
    font-family: var(--font-ui);
    line-height: normal;
    white-space: normal;
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.85);
    background: rgba(0, 0, 0, 0.38);
    box-shadow:
      inset 2px 0 0 var(--site),
      inset 0 0 0 1px rgba(255, 255, 255, 0.06);
  }

  .card.link {
    cursor: pointer;
  }

  .info {
    flex: 1;
    min-width: 0;
  }

  .from {
    overflow: hidden;
    font-size: 11px;
    font-weight: 600;
    line-height: 15px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .title,
  .desc {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    overflow: hidden;
    overflow-wrap: anywhere;
  }

  .title {
    -webkit-line-clamp: 2;
    line-clamp: 2;
    font-size: 13px;
    font-weight: 700;
    line-height: 17px;
    color: #9ed2ff;
  }

  .desc {
    -webkit-line-clamp: 2;
    line-clamp: 2;
    margin-top: 1px;
    font-size: 12px;
    font-weight: 400;
    line-height: 16px;
    color: #d5dbe1;
  }

  .compact .title,
  .compact .desc {
    -webkit-line-clamp: 1;
    line-clamp: 1;
  }

  .card.link.hover .title,
  .card.link:hover .title {
    text-decoration: underline;
    text-decoration-thickness: 1px;
    text-underline-offset: 2px;
  }

  .card.link.hover,
  .card.link:hover {
    background: rgba(0, 0, 0, 0.5);
  }

  .thumb {
    position: relative;
    flex: none;
    align-self: center;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.07);
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.85);
  }

  img {
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

  .play {
    position: absolute;
    top: 50%;
    left: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    padding-left: 2px;
    color: #fff;
    background: rgba(0, 0, 0, 0.62);
    border-radius: 50%;
    transform: translate(-50%, -50%);
  }

  .minimal {
    text-shadow: 0 1px 1px rgba(4, 10, 18, 0.55), 0 0 5px rgba(178, 226, 255, 0.35);
    background: rgba(168, 222, 255, 0.08);
    box-shadow: inset 2px 0 0 var(--site);
  }

  .minimal .from {
    font-weight: 400;
    color: #a9b9c8;
  }

  .minimal .title {
    color: #bfe4ff;
  }

  .minimal .desc {
    color: #dbe5ee;
  }

  .minimal.link.hover,
  .minimal.link:hover {
    background: rgba(168, 222, 255, 0.14);
  }

  .minimal .thumb {
    background: rgba(168, 222, 255, 0.09);
    box-shadow: none;
  }

  :global(.anim) img {
    transition: opacity 180ms ease-out;
  }
</style>
