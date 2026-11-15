<script lang="ts" module>
  import { SvelteSet } from "svelte/reactivity";

  /** URLs that failed to load. They are never requested again this session. */
  const failed = new SvelteSet<string>();
  /** URLs that have loaded once: shown at once, with no letter flashing first. */
  const loaded = new Set<string>();
</script>

<script lang="ts">
  // A person, spectator-card style: a circle picture with a thin coloured
  // ring. Until the picture has loaded, or if it fails, the circle shows the
  // first letter of the name.
  import { initial } from "../format";

  let {
    src = "",
    name,
    size,
    ring = "none",
    inline = false,
  }: {
    /** Image URL, or "" for the letter circle. */
    src?: string;
    name: string;
    /** Diameter in CSS px. */
    size: number;
    /** Ring colour: whisper violet, green (party, online), grey (offline), light blue, or no ring. */
    ring?: "whisper" | "green" | "offline" | "accent" | "none";
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

<span class="avatar {ring}" class:inline style:--size="{size}px" aria-hidden="true">
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
    border-radius: 50%;
    background: #252a31;
    /* About 1 px on the small sizes, 2 px on the toast. */
    box-shadow: 0 0 0 max(1px, calc(var(--size) / 22)) var(--ring);
  }

  /* Centred on the text of the line it sits in, without making the line taller. */
  .avatar.inline {
    margin: 0 5px 0 1px;
    vertical-align: calc(0.3em - var(--size) / 2);
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
    font-size: calc(var(--size) * 0.56);
    font-style: normal;
    font-weight: 700;
    line-height: 1;
    text-shadow: none;
    color: var(--text-soft);
  }

  .offline .letter {
    color: var(--muted);
  }

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    opacity: 0;
  }

  img.ready {
    opacity: 1;
  }

  .offline img.ready {
    opacity: 0.55;
  }
</style>
