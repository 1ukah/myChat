<script lang="ts">
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
  .stage {
    --dx: 0px;
    --dy: 0px;
  }

  .stage[data-style="slide-up"] {
    --dy: 14px;
  }

  .stage[data-style="slide-down"] {
    --dy: -14px;
  }

  .stage[data-style="slide-left"] {
    --dx: 24px;
  }

  .stage[data-style="slide-right"] {
    --dx: -24px;
  }
</style>
