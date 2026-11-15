<script lang="ts">
  // Sets up motion for everything inside it. Descendants animate only under
  // `.anim`, and read the style as CSS variables:
  //   --dx / --dy   where an entering element comes from (and exits to)
  //   --dur / --out entry and exit time
  // so one set of keyframes serves every animation style.
  import type { Snippet } from "svelte";
  import type { Motion } from "./motion";

  let { motion, children }: { motion: Motion; children: Snippet } = $props();
</script>

<div
  class="stage"
  class:anim={motion.on}
  data-style={motion.style}
  style:--dur="{motion.enterMs}ms"
  style:--out="{motion.exitMs}ms"
>
  {@render children()}
</div>

<style>
  /* fade and typewriter: no travel, opacity only. */
  .stage {
    --dx: 0px;
    --dy: 0px;
  }

  /* Enters moving up, from below. */
  .stage[data-style="slide-up"] {
    --dy: 14px;
  }

  .stage[data-style="slide-down"] {
    --dy: -14px;
  }

  /* Enters moving left, from the right. */
  .stage[data-style="slide-left"] {
    --dx: 24px;
  }

  .stage[data-style="slide-right"] {
    --dx: -24px;
  }
</style>
