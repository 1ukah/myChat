<script lang="ts">
  // The settings window (frameless, 980×680, resizable), laid out like the
  // CS2 settings menu: big tabs, a row of small sub-tabs, a centred column of
  // rows, and action links at the bottom right.
  import { onDestroy, setContext } from "svelte";
  import * as Cfg from "../../../bindings/mychat/internal/config/models";
  import { app, hideSettings, minimiseSettings, saveConfig } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Icon from "../ui/Icon.svelte";
  import AccountTab from "./AccountTab.svelte";
  import ChatTab from "./ChatTab.svelte";
  import HelpTab from "./HelpTab.svelte";
  import OverlayTab from "./OverlayTab.svelte";
  import { SECTIONS, type SectionsContext } from "./Section.svelte";
  import { topoTile } from "./topo";

  // Section ids match the <Section id> used inside each tab.
  const TABS = [
    { id: "account", label: "Account", sections: [{ id: "steam", label: "Steam" }] },
    {
      id: "chat",
      label: "Chat",
      sections: [
        { id: "chat-box", label: "Chat box" },
        { id: "pop-ups", label: "Pop-ups" },
      ],
    },
    {
      id: "overlay",
      label: "Overlay",
      sections: [
        { id: "appearance", label: "Appearance" },
        { id: "animation", label: "Animation" },
        { id: "position", label: "Position" },
      ],
    },
    {
      id: "help",
      label: "Help",
      sections: [
        { id: "commands", label: "Commands" },
        { id: "keys", label: "Keys" },
        { id: "about", label: "About" },
      ],
    },
  ] as const;

  type Tab = (typeof TABS)[number];
  type TabId = Tab["id"];

  // `?tab=chat` opens a given tab (handy for previews).
  const asked = new URLSearchParams(location.search).get("tab");
  const first: Tab = TABS.find((t) => t.id === asked) ?? TABS[0];

  let tab = $state<TabId>(first.id);
  let section = $state<string>(first.sections[0].id);
  const current = $derived<Tab>(TABS.find((t) => t.id === tab) ?? TABS[0]);

  setContext<SectionsContext>(SECTIONS, {
    get active() {
      return section;
    },
  });

  // ---- sub-tabs: jump to a section, and follow the scroll position ----

  let scroller = $state<HTMLElement>();
  /** While a jump is scrolling, the scroll position must not pick the section. */
  let spyAfter = 0;

  function pickTab(id: TabId): void {
    if (id === tab) return;
    tab = id;
    section = (TABS.find((t) => t.id === id) ?? TABS[0]).sections[0].id;
    notice = null;
    spyAfter = 0;
    if (scroller) scroller.scrollTop = 0;
  }

  function pickSection(id: string): void {
    section = id;
    spyAfter = performance.now() + 700;
    scroller
      ?.querySelector(`#sec-${id}`)
      ?.scrollIntoView({ behavior: app.config.animations ? "smooth" : "auto", block: "start" });
  }

  function spy(): void {
    if (!scroller || performance.now() < spyAfter) return;
    const top = scroller.getBoundingClientRect().top;
    const atEnd =
      scroller.scrollTop > 0 &&
      scroller.scrollTop + scroller.clientHeight >= scroller.scrollHeight - 2;
    let found: string = current.sections[0].id;
    for (const s of current.sections) {
      const el = scroller.querySelector(`#sec-${s.id}`);
      if (el && (atEnd || el.getBoundingClientRect().top - top <= 40)) found = s.id;
    }
    section = found;
  }

  // ---- bottom bar ----

  let notice = $state<{ text: string; bad: boolean } | null>(null);

  const canReset = $derived(tab === "chat" || tab === "overlay");

  async function resetDefaults(): Promise<void> {
    let patch: Partial<Cfg.Config>;
    if (tab === "chat") {
      patch = { closeAfterSend: true, showToasts: true, toastSeconds: 6 };
      const h = app.config.hotkey;
      if (h.code !== "KeyI" || h.ctrl || h.shift || h.alt) {
        patch.hotkey = new Cfg.Hotkey({ code: "KeyI" });
      }
    } else {
      patch = {
        showAvatars: true,
        blur: true,
        scale: 1,
        animations: true,
        animationStyle: "slide-up",
        animationSpeed: "normal",
        bottomOffset: 150,
      };
    }
    const label = current.label;
    const error = await saveConfig(patch);
    notice = error
      ? { text: error, bad: true }
      : { text: `${label} settings are back to their defaults.`, bad: false };
  }

  // ---- Steam summary in the tab bar ----

  const steam = $derived.by(() => {
    const s = app.status;
    // Same rule as the Account tab: a finished login counts before Steam connects.
    if (!s.accountName || (s.state === "needs_login" && app.login.phase !== "done")) {
      return { tone: "off", ring: "none", name: "Not signed in", text: "Steam" } as const;
    }
    const name = s.personaName || s.accountName;
    switch (s.state) {
      case "online":
        return { tone: "on", ring: "green", name, text: "Online" } as const;
      case "connecting":
      case "needs_login":
        return { tone: "warn", ring: "accent", name, text: "Connecting" } as const;
      default:
        return { tone: "bad", ring: "offline", name, text: "Offline" } as const;
    }
  });

  // ---- background texture ----

  const topo = URL.createObjectURL(new Blob([topoTile()], { type: "image/svg+xml" }));
  onDestroy(() => URL.revokeObjectURL(topo));
</script>

<div class="settings" class:anim={app.config.animations} style:--topo="url({topo})">
  <header class="titlebar">
    <div class="mark"><span>my</span>Chat</div>
    <div class="tagline">Steam chat overlay</div>
    <div class="window">
      <button type="button" aria-label="Minimise" onclick={minimiseSettings}>
        <Icon name="minimise" />
      </button>
      <button type="button" class="close" aria-label="Close to tray" onclick={hideSettings}>
        <Icon name="close" />
      </button>
    </div>
  </header>

  <div class="tabs">
    <div class="tablist" role="tablist" aria-label="Settings">
      {#each TABS as t (t.id)}
        <button
          type="button"
          role="tab"
          class="tab"
          class:active={tab === t.id}
          aria-selected={tab === t.id}
          onclick={() => pickTab(t.id)}
        >
          {t.label}
        </button>
      {/each}
    </div>
    <button type="button" class="steam" aria-label="Steam account" onclick={() => pickTab("account")}>
      {#if steam.tone === "off"}
        <span class="nobody"><Icon name="user" /></span>
      {:else}
        <Avatar size={30} ring={steam.ring} name={steam.name} />
      {/if}
      <span class="who">
        <bdi class="name">{steam.name}</bdi>
        <span class="state"><span class="dot {steam.tone}"></span>{steam.text}</span>
      </span>
    </button>
  </div>

  <div class="subtabs">
    {#each current.sections as s (s.id)}
      <button
        type="button"
        class="subtab"
        class:active={section === s.id}
        aria-current={section === s.id ? "true" : undefined}
        onclick={() => pickSection(s.id)}
      >
        {s.label}
      </button>
    {/each}
  </div>

  <main class="content" bind:this={scroller} onscroll={spy}>
    <div class="column">
      {#key tab}
        <div class="page">
          {#if tab === "account"}
            <AccountTab />
          {:else if tab === "chat"}
            <ChatTab />
          {:else if tab === "overlay"}
            <OverlayTab />
          {:else}
            <HelpTab />
          {/if}
        </div>
      {/key}
    </div>
  </main>

  <footer class="bottom">
    <div class="notice" class:bad={notice?.bad} role="status">
      {notice ? notice.text : "Changes are saved as you make them."}
    </div>
    {#if canReset}
      <button type="button" class="link" onclick={resetDefaults}>
        <Icon name="reset" />
        <span>Reset defaults</span>
      </button>
    {/if}
  </footer>
</div>

<style>
  .settings {
    position: fixed;
    inset: 0;
    display: grid;
    grid-template-rows: 36px 64px 44px minmax(0, 1fr) 44px;
    color: var(--text);
    /* Contour lines over a smoky dark base. */
    background:
      var(--topo) 0 0 / 560px 560px,
      radial-gradient(120% 90% at 50% 0%, #20252c 0%, #15171b 55%, #101215 100%),
      var(--bg);
  }

  /* ---------- title bar ---------- */

  .titlebar {
    --wails-draggable: drag;
    display: flex;
    align-items: center;
    gap: 12px;
    padding-left: 16px;
    background: rgba(0, 0, 0, 0.45);
  }

  .mark {
    font-family: var(--font-head);
    font-size: 21px;
    font-style: italic;
    font-weight: 700;
    line-height: 36px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: #fff;
  }

  .mark span {
    color: var(--accent);
  }

  .tagline {
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--dim);
  }

  .window {
    display: flex;
    margin-left: auto;
  }

  .window button {
    --wails-draggable: no-drag;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 46px;
    height: 36px;
    padding: 0;
    color: var(--muted);
    background: none;
    border: 0;
    cursor: pointer;
  }

  .window button:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.12);
  }

  .window button.close:hover {
    background: var(--red);
  }

  /* ---------- main tabs ---------- */

  .tabs {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    gap: 16px;
    padding: 0 16px;
    background: rgba(0, 0, 0, 0.3);
    border-bottom: 1px solid var(--line);
  }

  .tablist {
    display: flex;
  }

  .tab {
    position: relative;
    padding: 0 24px;
    font-family: var(--font-head);
    font-size: 30px;
    font-weight: 600;
    line-height: 63px;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    color: var(--muted);
    background: none;
    border: 0;
    cursor: pointer;
  }

  /* Thin vertical separators between tabs. */
  .tab + .tab::before {
    content: "";
    position: absolute;
    top: 20px;
    bottom: 19px;
    left: 0;
    width: 1px;
    background: var(--line);
  }

  .tab::after {
    content: "";
    position: absolute;
    right: 24px;
    bottom: -1px;
    left: 24px;
    height: 2px;
    background: var(--accent);
    opacity: 0;
  }

  .tab:hover {
    color: var(--text);
  }

  .tab.active {
    color: var(--accent);
  }

  .tab.active::after {
    opacity: 1;
  }

  /* Steam summary: square letter avatar, name, small status line. */
  .steam {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    padding: 0 8px;
    text-align: left;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .nobody {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    color: var(--muted);
    background: rgba(255, 255, 255, 0.05);
    border-radius: 50%;
    box-shadow: 0 0 0 1px var(--line-strong);
  }

  .who {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    max-width: 200px;
    overflow: hidden;
    font-size: 14px;
    font-weight: 600;
    line-height: 18px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: #fff;
  }

  .state {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    line-height: 16px;
    color: var(--muted);
  }

  .dot {
    flex: none;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--dim);
  }

  .dot.on {
    background: var(--green);
  }

  .dot.warn {
    background: var(--gold);
  }

  .dot.bad {
    background: var(--red);
  }

  .steam:hover .name {
    color: var(--accent);
  }

  /* ---------- sub-tabs ---------- */

  .subtabs {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 28px;
    background: rgba(0, 0, 0, 0.18);
    border-bottom: 1px solid var(--line-soft);
  }

  .subtab {
    height: 26px;
    padding: 0 12px;
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 600;
    line-height: 24px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--muted);
    background: none;
    border: 1px solid transparent;
    border-radius: 2px;
    cursor: pointer;
  }

  .subtab:hover {
    color: var(--text);
    border-color: var(--line);
  }

  /* The selected sub-tab is a filled, outlined muted-blue box. */
  .subtab.active {
    color: #fff;
    background: var(--accent-fill);
    border-color: var(--accent-line);
  }

  /* ---------- content ---------- */

  .content {
    overflow-x: hidden;
    overflow-y: auto;
    scrollbar-gutter: stable both-edges;
  }

  .content::-webkit-scrollbar {
    width: 10px;
  }

  .content::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.14);
    border: 3px solid transparent;
    background-clip: content-box;
  }

  .content::-webkit-scrollbar-thumb:hover {
    background-color: rgba(255, 255, 255, 0.3);
  }

  /* The centred column is a darker panel; the contour texture shows through
     it faintly and fully in the side panels left and right. */
  .column {
    width: min(760px, 100% - 48px);
    min-height: 100%;
    margin: 0 auto;
    padding: 0 24px 40px;
    /* Smoky glass: a little lighter through the middle, darker at the edges. */
    background:
      linear-gradient(
        to right,
        rgba(8, 10, 12, 0.3),
        rgba(8, 10, 12, 0) 18%,
        rgba(8, 10, 12, 0) 82%,
        rgba(8, 10, 12, 0.3)
      ),
      var(--panel);
    border-right: 1px solid var(--line);
    border-left: 1px solid var(--line);
  }

  /* ---------- bottom bar ---------- */

  .bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 0 24px;
    background: rgba(0, 0, 0, 0.4);
    border-top: 1px solid var(--line);
  }

  .notice {
    min-width: 0;
    overflow: hidden;
    font-size: 12px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .notice.bad {
    color: var(--red-text);
  }

  .link {
    flex: none;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 28px;
    padding: 0 4px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--text-soft);
    background: none;
    border: 0;
    cursor: pointer;
  }

  .link:hover {
    color: var(--accent);
  }

  /* ---------- focus and motion ---------- */

  button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .anim .tab,
  .anim .subtab,
  .anim .link,
  .anim .window button {
    transition:
      color 120ms,
      background-color 120ms,
      border-color 120ms;
  }

  .anim .tab::after {
    transition: opacity 120ms;
  }

  .anim .page {
    animation: page-in 160ms var(--ease-out) both;
  }

  @keyframes page-in {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
</style>
