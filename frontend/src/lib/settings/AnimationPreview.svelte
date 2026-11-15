<script lang="ts">
  // Live preview tile: the real toast, at reduced size, replaying the chosen
  // animation style and speed. It uses the overlay's own Stage, Panel and
  // Toast components, so what plays here is what plays in game.
  import { onDestroy, untrack } from "svelte";
  import * as C from "../../../bindings/mychat/internal/chat/models";
  import { motionOf } from "../overlay/motion";
  import Panel from "../overlay/Panel.svelte";
  import Stage from "../overlay/Stage.svelte";
  import Toast from "../overlay/Toast.svelte";
  import { app } from "../store.svelte";
  import Icon from "../ui/Icon.svelte";

  const sample = new C.Entry({
    id: 1,
    time: Date.now(),
    kind: "in",
    peerName: "sant1",
    text: "one more after this? gg wp, that last round was close",
  });

  const motion = $derived(motionOf(app.config));

  /** The sample toast is playing its exit before coming back in. */
  let leaving = $state(false);
  /** Bumped to rebuild the toast, which restarts the entry (and the typing). */
  let run = $state(0);
  let timer: ReturnType<typeof setTimeout> | undefined;

  function replay(): void {
    clearTimeout(timer);
    if (!motion.on) return;
    leaving = true;
    timer = setTimeout(() => {
      leaving = false;
      run++;
    }, motion.exitMs + 160);
  }

  // Replay whenever the style or the speed changes (not on first render).
  let first = true;
  $effect(() => {
    void motion.style;
    void motion.enterMs;
    void motion.on;
    if (first) {
      first = false;
      return;
    }
    untrack(replay);
  });

  onDestroy(() => clearTimeout(timer));
</script>

<div class="preview">
  <div class="scene" aria-hidden="true">
    <div class="window" class:blur={app.config.blur}>
      <Stage {motion}>
        {#key run}
          <Panel fixed={false} leave={leaving}>
            <Toast
              entry={sample}
              hotkey={app.config.hotkeyLabel}
              live={!leaving}
              avatars={app.config.showAvatars}
              typewriter={motion.typewriter}
              typeDelay={Math.round(motion.enterMs * 0.6)}
            />
          </Panel>
        {/key}
      </Stage>
    </div>
  </div>
  <button type="button" class="replay" disabled={!motion.on} onclick={replay}>
    <Icon name="retry" />
    <span>{motion.on ? "Replay" : "Animations off"}</span>
  </button>
</div>

<style>
  .preview {
    position: relative;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--line);
  }

  /* A stand-in game picture with dark and bright areas behind the toast. */
  .scene {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 132px;
    margin-top: 12px;
    overflow: hidden;
    background:
      radial-gradient(60px 60px at 78% 22%, rgba(255, 244, 214, 0.9), rgba(255, 244, 214, 0) 70%),
      linear-gradient(100deg, rgba(0, 0, 0, 0) 58%, #cdbb93 58%, #efe4c8 100%),
      linear-gradient(#34495f 0%, #6f8799 52%, #2a241b 52%, #14110d 100%);
  }

  /* The overlay window in miniature: 640×92 at 75 %. */
  .window {
    position: relative;
    flex: none;
    width: 640px;
    height: 92px;
    zoom: 0.75;
    contain: layout paint;
  }

  /* In game Windows blurs what is behind the window; here CSS can. */
  .window.blur {
    -webkit-backdrop-filter: blur(14px);
    backdrop-filter: blur(14px);
  }

  .replay {
    position: absolute;
    top: 20px;
    right: 8px;
    display: flex;
    align-items: center;
    gap: 6px;
    height: 26px;
    padding: 0 9px;
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.01em;
    text-transform: uppercase;
    color: #fff;
    background: rgba(0, 0, 0, 0.6);
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  .replay:hover {
    color: var(--accent);
    background: var(--tab-fill);
  }

  .replay:disabled {
    color: var(--muted);
    pointer-events: none;
  }

  .replay:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
</style>
