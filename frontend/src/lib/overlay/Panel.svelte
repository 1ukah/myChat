<script lang="ts">
  // The smoky glass panel behind the composer and the toast.
  //
  // In the overlay window it fills the whole viewport edge to edge with square
  // corners: with blur on, Windows blurs the game behind the entire window
  // rectangle. No backdrop-filter here (nothing to blur inside that window).
  import type { Snippet } from "svelte";

  let {
    leave = false,
    solid = false,
    fixed = true,
    children,
  }: {
    /** Mode is `hidden`: play the exit and stay invisible. */
    leave?: boolean;
    /** No blur behind the window: a denser fill keeps text readable. */
    solid?: boolean;
    /** Fill the viewport (overlay window) instead of the parent box (previews). */
    fixed?: boolean;
    children: Snippet;
  } = $props();
</script>

<div class="panel" class:leave class:solid class:fixed>
  {@render children()}
</div>

<style>
  .panel {
    position: absolute;
    inset: 0;
    overflow: hidden;
    /* Darker edges, slightly lighter centre; about 0.55–0.7 over the blur. */
    background: var(--glass);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.2),
      inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  }

  .panel.fixed {
    position: fixed;
  }

  .panel.solid {
    background: linear-gradient(
      to bottom,
      rgba(8, 10, 12, 0.93),
      rgba(18, 21, 25, 0.85) 38%,
      rgba(18, 21, 25, 0.85) 62%,
      rgba(8, 10, 12, 0.94)
    );
  }

  .panel.leave {
    opacity: 0;
  }

  /* Only transform and opacity move. The exit ends 20 ms before Go hides
     the window (--out is hideDelayMs − 20). */
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
