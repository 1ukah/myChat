<script lang="ts">
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
    variant?: "outline" | "accent" | "primary" | "go" | "danger";
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
    gap: 8px;
    height: 32px;
    padding: 0 12px;
    font-family: var(--font-ui);
    font-size: var(--fs-sm);
    font-weight: 600;
    line-height: 1;
    white-space: nowrap;
    color: #fff;
    background: rgba(255, 255, 255, 0.12);
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  .btn.lg {
    height: 40px;
    padding: 0 20px;
    font-size: var(--fs-md);
  }

  .btn:hover {
    background: rgba(255, 255, 255, 0.22);
  }

  .btn:active {
    background: rgba(255, 255, 255, 0.3);
  }

  .accent,
  .primary {
    color: var(--accent);
    background: var(--tab-fill);
  }

  .accent:hover,
  .primary:hover {
    color: #fff;
    background: var(--tab-fill-hover);
  }

  .accent:active,
  .primary:active {
    background: #2d4c62;
  }

  .go {
    background: var(--green);
  }

  .go:hover {
    background: #5cc661;
  }

  .go:active {
    background: #3c9440;
  }

  .danger {
    color: #ffd6d2;
    background: rgba(200, 53, 47, 0.42);
  }

  .danger:hover {
    color: #fff;
    background: var(--red);
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
      color 120ms;
  }
</style>
