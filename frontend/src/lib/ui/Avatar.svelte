<script lang="ts" module>
  import { SvelteSet } from "svelte/reactivity";

  /** URLs that failed to load. They are never requested again this session. */
  const failed = new SvelteSet<string>();
  /** URLs that have loaded once: shown at once, with no letter flashing first. */
  const loaded = new Set<string>();
</script>

<script lang="ts">
  // A person, CS2 style: a square picture with a thin coloured border, as in
  // the friends list and the scoreboard. `round` gives the circle CS2 uses
  // for the spectated player. Until the picture has loaded, or if it fails,
  // the dark tile shows the first letter of the name ("?" without a name).
  import { initial } from "../format";

  let {
    src = "",
    name,
    size,
    ring = "none",
    tint = "",
    round = false,
    inline = false,
  }: {
    /** Image URL, or "" for the letter tile. */
    src?: string;
    /** "" shows CS2's "?" tile. */
    name: string;
    /** Side in CSS px. */
    size: number;
    /** Border colour: whisper blue, green (online), grey (offline), light blue, or none. */
    ring?: "whisper" | "green" | "offline" | "accent" | "none";
    /** A CSS colour for the border (a player colour); wins over `ring`. */
    tint?: string;
    /** Circle instead of square (the person in focus). */
    round?: boolean;
    /** Sits in a line of text (chat log) instead of a flex row. */
    inline?: boolean;
  } = $props();

  const usable = $derived(src !== "" && !failed.has(src));
  /** The URL this instance has seen finish loading. */
  let seen = $state("");
  const ready = $derived(usable && (seen === src || loaded.has(src)));

  function onload(): void {
    loaded.add(src);
    seen = src;
  }

  function onerror(): void {
    failed.add(src);
  }
</script>

<span
  class="avatar {ring}"
  class:inline
  class:round
  style:--size="{size}px"
  style:--ring={tint || null}
  aria-hidden="true"
>
  {#if !ready}<span class="letter">{initial(name)}</span>{/if}
  {#if usable}
    <img {src} alt="" loading="eager" decoding="async" draggable="false" class:ready {onload} {onerror} />
  {/if}
</span>

<style>
  .avatar {
    --ring: transparent;
    position: relative;
    flex: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--size);
    height: var(--size);
    background: var(--tile);
    /* Square in both looks; a parent may still set a radius. */
    border-radius: var(--avatar-radius, 0);
    /* About 1 px on the small sizes, 2 px from 44 px up. */
    box-shadow: 0 0 0 max(1px, calc(var(--size) / 22)) var(--ring);
  }

  .avatar.round {
    border-radius: 50%;
  }

  /* Centred on the text of the line it sits in, without making the line taller. */
  .avatar.inline {
    margin: 0 6px 0 2px;
    vertical-align: calc(0.32em - var(--size) / 2);
  }

  .whisper {
    --ring: var(--whisper);
  }

  .green {
    --ring: var(--green);
  }

  .offline {
    --ring: #6b7279;
  }

  .accent {
    --ring: var(--accent);
  }

  .letter {
    font-family: var(--font-head);
    font-size: calc(var(--size) * 0.6);
    font-style: normal;
    font-weight: 700;
    line-height: 1;
    text-shadow: none;
    color: #fff;
  }

  .offline .letter {
    color: var(--muted);
  }

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: inherit;
    opacity: 0;
  }

  img.ready {
    opacity: 1;
  }

  .offline img.ready {
    opacity: 0.55;
  }
</style>
