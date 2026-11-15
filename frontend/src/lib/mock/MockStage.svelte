<script lang="ts">
  // Development-only stage for the overlay in a plain browser. It stands in
  // for what Windows and Go do in the real app: a game picture behind, and a
  // window of exactly the layout size, docked bottom middle, that is resized
  // per mode, blurred when blur is on, zoomed by the size setting, and hidden
  // `hideDelayMs` after `hidden`. The real page stays fully transparent.
  import type { Snippet } from "svelte";
  import { app } from "../store.svelte";

  let { children }: { children: Snippet } = $props();

  const params = new URLSearchParams(location.search);
  const bg = params.get("bg") ?? "mixed";

  let visible = $state(false);
  let height = $state(app.layout.composerHeight);

  $effect(() => {
    const m = app.mode;
    if (m.mode === "hidden") {
      const t = setTimeout(() => (visible = false), m.hideDelayMs);
      return () => clearTimeout(t);
    }
    height = m.mode === "toast" ? app.layout.toastHeight : app.layout.composerHeight;
    visible = true;
  });

  /** Link to this page with some query values changed. */
  function href(change: Record<string, string>): string {
    const q = new URLSearchParams(location.search);
    for (const [k, v] of Object.entries(change)) q.set(k, v);
    return "?" + q.toString();
  }

  const links: [string, [string, Record<string, string>][]][] = [
    [
      "overlay",
      [
        ["composer", { view: "overlay", mock: "composer" }],
        ["toast", { view: "overlay", mock: "toast" }],
        ["suggest", { view: "overlay", mock: "suggest" }],
        ["offline", { view: "overlay", mock: "offline" }],
      ],
    ],
    [
      "settings",
      [
        ["login", { view: "settings", mock: "login" }],
        ["watching", { view: "settings", mock: "watching" }],
        ["nosteam", { view: "settings", mock: "nosteam" }],
        ["expired", { view: "settings", mock: "expired" }],
        ["error", { view: "settings", mock: "error" }],
        ["online", { view: "settings", mock: "online" }],
      ],
    ],
    [
      "style",
      [
        ["fade", { style: "fade" }],
        ["up", { style: "slide-up" }],
        ["down", { style: "slide-down" }],
        ["left", { style: "slide-left" }],
        ["right", { style: "slide-right" }],
        ["type", { style: "typewriter" }],
      ],
    ],
    [
      "speed",
      [
        ["fast", { speed: "fast" }],
        ["normal", { speed: "normal" }],
        ["slow", { speed: "slow" }],
      ],
    ],
    [
      "avatars",
      [
        ["on", { avatars: "1" }],
        ["off", { avatars: "0" }],
      ],
    ],
    [
      "blur",
      [
        ["on", { blur: "1" }],
        ["off", { blur: "0" }],
      ],
    ],
    [
      "anim",
      [
        ["on", { anim: "1" }],
        ["off", { anim: "0" }],
      ],
    ],
    [
      "scale",
      [
        ["75", { scale: "0.75" }],
        ["100", { scale: "1" }],
        ["150", { scale: "1.5" }],
      ],
    ],
    [
      "scene",
      [
        ["mixed", { bg: "mixed" }],
        ["dark", { bg: "dark" }],
        ["bright", { bg: "bright" }],
      ],
    ],
  ];
</script>

<div class="game {bg}" aria-hidden="true">
  <div class="sky"></div>
  <div class="sun"></div>
  <div class="wall"></div>
  <div class="crate a"></div>
  <div class="crate b"></div>
  <div class="beam"></div>
  <div class="floor"></div>
  <div class="cross"></div>
</div>

<div class="dock" style:bottom="{app.config.bottomOffset}px">
  <div
    class="window"
    class:blur={app.mode.blur}
    class:shown={visible}
    style:width="{app.layout.width}px"
    style:height="{height}px"
    style:zoom={app.config.scale}
  >
    {@render children()}
  </div>
</div>

<nav class="dev">
  <b>myChat mock</b>
  {#each links as [group, items] (group)}
    <div>
      <span>{group}</span>
      {#each items as [name, change] (name)}
        <a href={href(change)}>{name}</a>
      {/each}
    </div>
  {/each}
  <p>Press <kbd>{app.config.hotkeyLabel}</kbd> to open the chat box, then type. A message with “fail” fails to send.</p>
</nav>

<style>
  /* ---------- a stand-in game picture: dark and bright areas behind the box ---------- */

  .game {
    position: fixed;
    inset: 0;
    overflow: hidden;
    background: linear-gradient(#1a2330 0%, #2b3340 46%, #17191d 46%, #0c0d10 100%);
  }

  .sky {
    position: absolute;
    inset: 0 0 54% 0;
    background: linear-gradient(#35506e, #8fa9bd);
    opacity: 0.55;
  }

  .sun {
    position: absolute;
    top: 6%;
    right: 14%;
    width: 220px;
    height: 220px;
    border-radius: 50%;
    background: radial-gradient(#fff 0%, #ffe9b0 30%, rgba(255, 220, 150, 0) 70%);
  }

  /* A sunlit wall that crosses behind the right half of the overlay. */
  .wall {
    position: absolute;
    right: 0;
    bottom: 0;
    width: 46%;
    height: 62%;
    background: linear-gradient(100deg, #d9c9a3, #f3ead2 60%, #fff);
  }

  .crate {
    position: absolute;
    bottom: 12%;
    background: #c9842b;
    box-shadow: inset 0 0 0 6px #8a5413;
  }

  .crate.a {
    left: 16%;
    width: 150px;
    height: 150px;
  }

  .crate.b {
    left: 31%;
    width: 96px;
    height: 96px;
    background: #4fa3c7;
    box-shadow: inset 0 0 0 6px #1f5f7d;
  }

  .beam {
    position: absolute;
    bottom: 0;
    left: 38%;
    width: 90px;
    height: 120%;
    background: linear-gradient(rgba(255, 255, 255, 0), rgba(255, 255, 255, 0.85));
    transform: rotate(24deg);
    transform-origin: bottom;
  }

  .floor {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    height: 12%;
    background: linear-gradient(#3a3023, #17130e);
  }

  .cross {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 22px;
    height: 22px;
    margin: -11px;
    background:
      linear-gradient(#3dff6a, #3dff6a) center / 2px 100% no-repeat,
      linear-gradient(#3dff6a, #3dff6a) center / 100% 2px no-repeat;
  }

  .game.dark .sky,
  .game.dark .sun,
  .game.dark .wall,
  .game.dark .beam {
    display: none;
  }

  .game.dark .crate {
    filter: brightness(0.35);
  }

  .game.bright {
    background: linear-gradient(#bcd6ea 0%, #eef4f7 46%, #e9dfc6 46%, #cbbd9c 100%);
  }

  .game.bright .sky {
    opacity: 0;
  }

  .game.bright .wall {
    width: 100%;
    height: 40%;
    background: linear-gradient(#f6f1e2, #fff);
  }

  .game.bright .floor {
    background: linear-gradient(#d8cbac, #b7a986);
  }

  /* ---------- the pretend overlay window ---------- */

  .dock {
    position: fixed;
    right: 0;
    left: 0;
    display: flex;
    justify-content: center;
    pointer-events: none;
  }

  /* `contain` makes this box the containing block of the overlay's
     `position: fixed` panel and clips it, as a real window would. */
  .window {
    position: relative;
    flex: none;
    contain: layout paint;
    visibility: hidden;
  }

  .window.shown {
    visibility: visible;
  }

  /* Windows blurs the game behind the whole window when blur is on. */
  .window.shown.blur {
    -webkit-backdrop-filter: blur(18px);
    backdrop-filter: blur(18px);
  }

  /* ---------- small dev menu ---------- */

  .dev {
    position: fixed;
    top: 8px;
    left: 8px;
    width: 264px;
    padding: 8px 10px;
    font: 12px/18px "Segoe UI", sans-serif;
    color: #cfd4d9;
    background: rgba(0, 0, 0, 0.72);
    user-select: text;
  }

  .dev b {
    display: block;
    color: #fff;
  }

  .dev span {
    display: inline-block;
    width: 56px;
    color: #8f979e;
  }

  .dev a {
    margin-right: 8px;
    color: #86c5ff;
  }

  .dev p {
    margin: 6px 0 0;
    color: #8f979e;
  }

  .dev kbd {
    padding: 0 4px;
    font: inherit;
    color: #fff;
    border: 1px solid #666;
  }
</style>
