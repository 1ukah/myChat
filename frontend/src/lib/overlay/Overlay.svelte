<script lang="ts">
  // The in-game overlay window. It is never focused and never sees the mouse:
  // everything here is drawn from state that Go pushes. See docs/UI_CONTRACT.md.
  //
  // Clicks on pictures are caught by Go, which needs to know where they are
  // (spots.ts). A clicked picture opens at full size (Viewer): the window
  // then covers the whole monitor, and the 640×360 area keeps its place on
  // screen, at the bottom middle, `bottomOffset` px up.
  import { onMount, tick, untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import { app, overlayReady, overlayResized } from "../store.svelte";
  import Composer from "./Composer.svelte";
  import { pictureOf } from "./media";
  import { asLook, motionOf } from "./motion";
  import Panel from "./Panel.svelte";
  import { scheduleSpots, spotSource } from "./spots";
  import Stage from "./Stage.svelte";
  import Toasts from "./Toasts.svelte";
  import Viewer from "./Viewer.svelte";

  /** More than fits in 360 px even when every line is short. */
  const MAX_LINES = 48;

  const open = $derived(app.mode.mode === "composer");
  const live = $derived(app.mode.mode !== "hidden");
  const motion = $derived(motionOf(app.mode, app.mode.hideDelayMs));
  /** The chat look: Legacy (CS2 style) or Minimal (dark Panorama glass bands). */
  const minimal = $derived(asLook(app.mode.overlayStyle) === "minimal");

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

  // The pop-ups are read only in toast mode, so they stay as they were while
  // the exit animation fades them out.
  let heldToasts: C.Entry[] = [];
  const toasts = $derived.by(() =>
    app.mode.mode === "toast" ? (heldToasts = app.toasts) : heldToasts,
  );

  let heldTotal = 0;
  const total = $derived.by(() => (open ? (heldTotal = app.log.length) : heldTotal));

  // ---- pictures that open when clicked ----

  const viewer = $derived(live ? app.mode.viewer : null);

  // The picture drawn: a closed one stays while its closing animation plays
  // (Go shrinks the window only after it).
  let shownViewer = $state<C.Media | null>(null);
  let closing = $state(false);
  let closeTimer = 0;
  /** Changes on each opening, so a picture reopened mid-close starts over. */
  let opening = $state(0);
  $effect(() => {
    const v = viewer;
    untrack(() => {
      clearTimeout(closeTimer);
      if (v) {
        if (!shownViewer || closing || shownViewer.id !== v.id) opening++;
        shownViewer = v;
        closing = false;
      } else if (shownViewer && motion.on && motion.exitMs > 0) {
        closing = true;
        closeTimer = window.setTimeout(() => {
          shownViewer = null;
          closing = false;
        }, motion.exitMs);
      } else {
        shownViewer = null;
        closing = false;
      }
    });
  });

  /** How far the 640×360 area sits above the bottom of the monitor, in CSS px. */
  const lift = $derived(Math.max(0, app.config.bottomOffset) / (app.config.scale || 1));

  let area = $state<HTMLElement>();

  spotSource(() => {
    if (!area || !live) return null;
    const media = new Map<string, C.Media>();
    for (const e of open ? lines : toasts) {
      const m = pictureOf(e);
      if (m) media.set(m.id, m);
    }
    return { area, media, key: `${app.mode.mode}:${viewer?.id ?? ""}`, settle: motion.enterMs };
  });

  $effect(() => {
    // Everything that adds, removes or moves a picture.
    void app.mode.mode;
    void viewer;
    void lines;
    void toasts;
    void composer.scroll;
    void app.config.showImages;
    void app.config.toastLayout;
    void app.mode.showAvatars;
    void minimal;
    untrack(scheduleSpots);
  });

  function fontsReady(): Promise<unknown> {
    // Legacy is set in Saira and Noto Sans, Minimal in Inter (one variable
    // file covers every weight). Both are loaded: the look can change later.
    const faces = [
      '600 16px "Saira Semi Condensed"',
      '700 16px "Saira Semi Condensed"',
      '400 16px "Noto Sans"',
      '600 16px "Noto Sans"',
      '700 16px "Noto Sans"',
      '400 16px "Inter Variable"',
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

    // Go hides the window while it changes size (opening or closing a
    // picture); it shows again once a frame at the new size is drawn.
    let pending = 0;
    const onResize = (): void => {
      cancelAnimationFrame(pending);
      pending = requestAnimationFrame(() => {
        pending = requestAnimationFrame(overlayResized);
      });
    };
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("resize", onResize);
      cancelAnimationFrame(pending);
    };
  });
</script>

<div class="window">
  <Stage {motion}>
    <div class="area" style:--lift="{lift}px" bind:this={area}>
      {#if shown}
        {#key shown}
          <Panel leave={!live} solid={!app.mode.blur}>
            {#if shown === "composer"}
              <Composer
                {composer}
                {lines}
                {total}
                {status}
                live={open}
                avatars={app.mode.showAvatars}
                images={app.config.showImages}
                {minimal}
                typewriter={motion.typewriter}
              />
            {:else}
              <Toasts
                entries={toasts}
                layout={app.config.toastLayout}
                hotkey={app.config.hotkeyLabel}
                {live}
                avatars={app.mode.showAvatars}
                images={app.config.showImages}
                {minimal}
                zoom
                {motion}
              />
            {/if}
          </Panel>
        {/key}
      {/if}
    </div>
    {#if shownViewer}
      {#key opening}
        <Viewer media={shownViewer} {motion} {closing} {minimal} composer={open} />
      {/key}
    {/if}
  </Stage>
</div>

<style>
  /* The whole window, as a size container for the area below. */
  .window {
    position: fixed;
    inset: 0;
    container-type: size;
  }

  /* The 640×360 area. It fills the window as usual; when the window covers
     the whole monitor (a picture is open) it stays at the bottom middle,
     `--lift` up, exactly where the window was, so nothing moves. The
     clamp() is 0 for a 360 px high window and --lift for anything taller.
     The transform also makes it the box for the panel's position: fixed. */
  .area {
    position: absolute;
    left: 50%;
    bottom: clamp(0px, (100cqh - 361px) * 1000, var(--lift));
    width: 640px;
    height: 360px;
    transform: translateX(-50%);
  }
</style>
