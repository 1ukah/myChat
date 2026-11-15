<script lang="ts">
  // The in-game overlay window. It is never focused and never sees the mouse:
  // everything here is drawn from state that Go pushes. See docs/UI_CONTRACT.md.
  import { onMount, tick } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { app, overlayReady } from "../store.svelte";
  import Composer from "./Composer.svelte";
  import { motionOf } from "./motion";
  import Panel from "./Panel.svelte";
  import Stage from "./Stage.svelte";
  import Toast from "./Toast.svelte";

  /** More than fits in 360 px even when every line is short. */
  const MAX_LINES = 48;

  const open = $derived(app.mode.mode === "composer");
  const live = $derived(app.mode.mode !== "hidden");
  const motion = $derived(motionOf(app.mode, app.mode.hideDelayMs));

  // What was last on screen. After `hidden` it stays mounted so the exit
  // animation has something to fade, and so reopening costs no DOM work.
  let lastShown: "composer" | "toast" | null = null;
  const shown = $derived.by(() => {
    const m = app.mode.mode;
    if (m === "composer" || m === "toast") lastShown = m;
    return lastShown;
  });

  // The composer's inputs are read only while it is open. Once hidden these
  // return the last frame and stop depending on the store, so events that
  // arrive in the background (new messages, the input reset) touch no DOM.
  let heldComposer = app.composer;
  const composer = $derived.by(() => (open ? (heldComposer = app.composer) : heldComposer));

  let heldStatus = app.status;
  const status = $derived.by(() => (open ? (heldStatus = app.status) : heldStatus));

  let heldLines: C.Entry[] = [];
  const lines = $derived.by(() => {
    if (open) {
      const total = app.log.length;
      const end = Math.max(Math.min(total, 1), total - Math.max(0, app.composer.scroll));
      heldLines = app.log.slice(Math.max(0, end - MAX_LINES), end);
    }
    return heldLines;
  });

  function fontsReady(): Promise<unknown> {
    const faces = [
      '600 16px "Barlow Condensed"',
      '700 16px "Barlow Condensed"',
      '400 16px "Noto Sans"',
      '600 16px "Noto Sans"',
      'italic 400 16px "Noto Sans"',
    ];
    const loaded = Promise.all(faces.map((f) => document.fonts.load(f))).catch(() => undefined);
    const timeout = new Promise((resolve) => setTimeout(resolve, 1500));
    return Promise.race([loaded, timeout]);
  }

  onMount(() => {
    // Tell Go the page has rendered. Fonts are decoded first so the very
    // first open in game shows final text, with no swap or reflow.
    void tick()
      .then(fontsReady)
      .then(overlayReady);
  });
</script>

<Stage {motion}>
  {#if shown}
    {#key shown}
      <Panel leave={!live} solid={!app.mode.blur}>
        {#if shown === "composer"}
          <Composer
            {composer}
            {lines}
            {status}
            live={open}
            avatars={app.mode.showAvatars}
            typewriter={motion.typewriter}
          />
        {:else if app.toast}
          <Toast
            entry={app.toast}
            hotkey={app.config.hotkeyLabel}
            {live}
            avatars={app.mode.showAvatars}
            typewriter={motion.typewriter}
            typeDelay={Math.round(motion.enterMs * 0.6)}
          />
        {/if}
      </Panel>
    {/key}
  {/if}
</Stage>
