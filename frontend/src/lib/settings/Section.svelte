<script lang="ts" module>
  /** Context shared by the settings shell: which section is selected. */
  export const SECTIONS = Symbol("sections");
  export interface SectionsContext {
    readonly active: string;
  }
</script>

<script lang="ts">
  // A titled group of rows. The title is grey bold italic, like "Video" in
  // the CS2 settings; the section picked in the sub-tab row is lit.
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
    padding: 0 0 4px;
    font-family: var(--font-title);
    font-size: 23px;
    font-style: italic;
    font-weight: 700;
    line-height: 30px;
    color: var(--muted);
  }

  .active h2 {
    color: #c3cad1;
  }

  :global(.anim) h2 {
    transition: color 160ms;
  }
</style>
