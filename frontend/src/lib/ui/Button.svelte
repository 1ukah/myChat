<script lang="ts">
  // Flat Panorama button: a square box with a wide-tracked uppercase label.
  import type { Snippet } from "svelte";
  import Icon, { type IconName } from "./Icon.svelte";

  let {
    variant = "outline",
    size = "md",
    icon,
    disabled = false,
    onclick,
    children,
  }: {
    /** outline: neutral · accent: light blue · go: CS2 green · danger: red */
    variant?: "outline" | "accent" | "go" | "danger";
    size?: "md" | "lg";
    icon?: IconName;
    disabled?: boolean;
    onclick: () => void;
    children: Snippet;
  } = $props();
</script>

<button type="button" class="btn {variant} {size}" {disabled} {onclick}>
  {#if icon}<Icon name={icon} size={size === "lg" ? 18 : 16} />{/if}
  <span>{@render children()}</span>
</button>

<style>
  .btn {
    --wails-draggable: no-drag;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    height: 36px;
    padding: 0 20px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 600;
    line-height: 1;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--text);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
    cursor: pointer;
  }

  .btn.lg {
    height: 52px;
    padding: 0 32px;
    font-size: 20px;
    font-weight: 700;
  }

  .btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.55);
  }

  .btn:active {
    background: rgba(255, 255, 255, 0.2);
  }

  .accent {
    color: var(--accent);
    background: var(--accent-fill);
    border-color: var(--accent-line);
  }

  .accent:hover {
    color: #fff;
    background: rgba(134, 197, 255, 0.28);
    border-color: var(--accent);
  }

  .accent:active {
    background: rgba(134, 197, 255, 0.4);
  }

  .go {
    color: #fff;
    background: var(--green);
    border-color: #6dcc76;
  }

  .go:hover {
    background: #4cc257;
    border-color: #9be3a2;
  }

  .go:active {
    background: #379a40;
  }

  .danger {
    color: #ffe1de;
    background: rgba(200, 53, 47, 0.4);
    border-color: var(--red);
  }

  .danger:hover {
    color: #fff;
    background: var(--red);
    border-color: #e9605a;
  }

  .danger:active {
    background: #a82a25;
  }

  .btn:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(.anim) .btn {
    transition:
      background-color 120ms,
      border-color 120ms,
      color 120ms;
  }
</style>
