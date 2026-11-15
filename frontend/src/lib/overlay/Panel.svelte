<script lang="ts">
  // The moving layer that holds the chat box or the toast.
  //
  // It paints nothing itself: the overlay window is per-pixel transparent,
  // and the content draws its own CS2-style translucent boxes, leaving
  // see-through margins. It only sets how dense those fills are and plays
  // the entry and exit. No backdrop-filter (nothing to blur in that window).
  import type { Snippet } from "svelte";

  let {
    leave = false,
    solid = false,
    fixed = true,
    children,
  }: {
    /** Mode is `hidden`: play the exit and stay invisible. */
    leave?: boolean;
    /** No blur behind the window (the default): denser fills keep text readable. */
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
    /* With blur behind, the picture is already calm: lighter fills. */
    --box-fill: rgba(0, 0, 0, 0.42);
    --frame-fill: rgba(0, 0, 0, 0.2);
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  .panel.fixed {
    position: fixed;
  }

  /* Like CS2's own chat box: frame and box together are about 0.65 black. */
  .panel.solid {
    --box-fill: rgba(0, 0, 0, 0.5);
    --frame-fill: var(--hud-frame);
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
