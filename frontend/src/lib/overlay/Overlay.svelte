<script lang="ts">
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

  const MAX_LINES = 48;

  const open = $derived(app.mode.mode === "composer");
  const live = $derived(app.mode.mode !== "hidden");
  const motion = $derived(motionOf(app.mode, app.mode.hideDelayMs));
  const minimal = $derived(asLook(app.mode.overlayStyle) === "minimal");

  let lastShown: "composer" | "toast" | null = null;
  const shown = $derived.by(() => {
    const m = app.mode.mode;
    if (m === "composer" || m === "toast") lastShown = m;
    return lastShown;
  });

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

  let heldToasts: C.Entry[] = [];
  const toasts = $derived.by(() =>
    app.mode.mode === "toast" ? (heldToasts = app.toasts) : heldToasts,
  );

  let heldTotal = 0;
  const total = $derived.by(() => (open ? (heldTotal = app.log.length) : heldTotal));

  const viewer = $derived(live ? app.mode.viewer : null);

  let shownViewer = $state<C.Media | null>(null);
  let closing = $state(false);
  let closeTimer = 0;
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

  const SETTLE_FRAMES = 4;
  let settled = $state(true);

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
    void tick()
      .then(fontsReady)
      .then(overlayReady);

    let pending = 0;
    const onResize = (): void => {
      settled = false;
      cancelAnimationFrame(pending);
      let left = SETTLE_FRAMES;
      const step = (): void => {
        if (--left > 0) {
          pending = requestAnimationFrame(step);
          return;
        }
        settled = true;
        overlayResized();
      };
      pending = requestAnimationFrame(step);
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
        <Viewer media={shownViewer} {motion} {closing} {minimal} {settled} composer={open} />
      {/key}
    {/if}
  </Stage>
</div>

<style>
  .window {
    position: fixed;
    inset: 0;
    container-type: size;
  }

  .area {
    position: absolute;
    left: 50%;
    bottom: clamp(0px, (100cqh - 361px) * 1000, var(--lift));
    width: 640px;
    height: 360px;
    transform: translateX(-50%);
  }
</style>
