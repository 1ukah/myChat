<script lang="ts" module>
  /** Context shared by the settings shell: which section is selected. */
  export const SECTIONS = Symbol("sections");
  export interface SectionsContext {
    readonly active: string;
  }
</script>

<script lang="ts">
  // A titled group of rows. The title is a plain semibold heading; the
  // section picked in the sub-tab row is lit.
  import { getContext, type Snippet } from "svelte";

  let { id, title, children }: { id: string; title: string; children: Snippet } = $props();

  const sections = getContext<SectionsContext | undefined>(SECTIONS);
</script>

<section id="sec-{id}" class:active={sections?.active === id}>
  <h2>{title}</h2>
  <div class="rows">{@render children()}</div>
</section>

<style>
  section {
    scroll-margin-top: 0;
    padding-top: 24px;
  }

  h2 {
    margin: 0;
    padding: 0 0 6px;
    font-size: var(--fs-lg);
    font-weight: 600;
    line-height: 28px;
    letter-spacing: -0.01em;
    color: #b6bec6;
  }

  .active h2 {
    color: #fff;
  }

  :global(.anim) h2 {
    transition: color 160ms;
  }
</style>
