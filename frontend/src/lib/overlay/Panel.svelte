<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    leave = false,
    solid = false,
    fixed = true,
    children,
  }: {
    leave?: boolean;
    solid?: boolean;
    fixed?: boolean;
    children: Snippet;
  } = $props();
</script>

<div class="panel" class:leave class:solid class:fixed>
  {@render children()}
</div>

<style>
  .panel {
    --box-fill: rgba(0, 0, 0, 0.42);
    --frame-fill: rgba(0, 0, 0, 0.2);
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  .panel.fixed {
    position: fixed;
  }

  .panel.solid {
    --box-fill: rgba(0, 0, 0, 0.5);
    --frame-fill: var(--hud-frame);
  }

  .panel.leave {
    opacity: 0;
  }

  :global(.anim) .panel {
    animation: panel-in var(--dur) var(--ease-out) both;
  }

  :global(.anim) .panel.leave {
    animation: panel-out var(--out) ease-in both;
  }

  @keyframes panel-in {
    from {
      opacity: 0;
      transform: translate(var(--dx), var(--dy));
    }
    to {
      opacity: 1;
      transform: none;
    }
  }

  @keyframes panel-out {
    from {
      opacity: 1;
      transform: none;
    }
    to {
      opacity: 0;
      transform: translate(var(--dx), var(--dy));
    }
  }
</style>
