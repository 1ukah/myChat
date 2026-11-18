<script lang="ts" module>
  import { SvelteSet } from "svelte/reactivity";

  const failed = new SvelteSet<string>();
  const loaded = new Set<string>();
</script>

<script lang="ts">
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
    src?: string;
    name: string;
    size: number;
    ring?: "whisper" | "green" | "offline" | "accent" | "none";
    tint?: string;
    round?: boolean;
    inline?: boolean;
  } = $props();

  const usable = $derived(src !== "" && !failed.has(src));
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
    border-radius: var(--avatar-radius, 0);
    box-shadow: 0 0 0 max(1px, calc(var(--size) / 22)) var(--ring);
  }

  .avatar.round {
    border-radius: 50%;
  }

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
    font-family: inherit;
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
