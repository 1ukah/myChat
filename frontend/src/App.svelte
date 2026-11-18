<script lang="ts">
  import { onMount, type Component, type Snippet } from "svelte";
  import Overlay from "./lib/overlay/Overlay.svelte";
  import Settings from "./lib/settings/Settings.svelte";
  import { app, start, type View } from "./lib/store.svelte";

  const view: View =
    new URLSearchParams(location.search).get("view") === "overlay" ? "overlay" : "settings";

  let Stage = $state.raw<Component<{ children: Snippet }> | null>(null);

  onMount(async () => {
    await start(view);
    if (app.mock && view === "overlay") {
      Stage = (await import("./lib/mock/MockStage.svelte")).default;
    }
  });
</script>

{#if !app.ready}
{:else if view === "settings"}
  <Settings />
{:else if !app.mock}
  <Overlay />
{:else if Stage}
  <Stage>
    <Overlay />
  </Stage>
{/if}
