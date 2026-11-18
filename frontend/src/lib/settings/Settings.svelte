<script lang="ts">
  import { onDestroy, setContext, tick } from "svelte";
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

  const SETTINGS_TABS = TABS.filter((t) => t.id !== "help");

  const asked = new URLSearchParams(location.search).get("tab");
  const first: Tab = TABS.find((t) => t.id === asked) ?? TABS[0];

  let tab = $state<TabId>(first.id);
  let section = $state<string>(first.sections[0].id);
  const current = $derived<Tab>(TABS.find((t) => t.id === tab) ?? TABS[0]);
  const inHelp = $derived(tab === "help");
  let lastSettingsTab: TabId = first.id === "help" ? "account" : first.id;

  setContext<SectionsContext>(SECTIONS, {
    get active() {
      return section;
    },
  });

  let scroller = $state<HTMLElement>();
  let spyAfter = 0;

  function pickTab(id: TabId): void {
    if (id === tab) return;
    tab = id;
    if (id !== "help") lastSettingsTab = id;
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

  const INDEX: readonly { tab: TabId; section: string; label: string; words?: string }[] = [
    { tab: "account", section: "steam", label: "Steam account", words: "sign in login log out switch saved account password steam guard qr code" },
    { tab: "chat", section: "chat-box", label: "Open chat", words: "hotkey key bind shortcut" },
    { tab: "chat", section: "chat-box", label: "Close after sending" },
    { tab: "chat", section: "chat-box", label: "Double-tap Enter to close", words: "double press twice empty line" },
    { tab: "chat", section: "pop-ups", label: "Show new message pop-ups", words: "toast notification" },
    { tab: "chat", section: "pop-ups", label: "Message sound", words: "sound audio chime ding notification steam mute volume" },
    { tab: "chat", section: "pop-ups", label: "Pop-up duration", words: "toast seconds time" },
    { tab: "chat", section: "pop-ups", label: "Multiple messages", words: "toast pop-up layout stack group grouped discord sender" },
    { tab: "chat", section: "pop-ups", label: "Messages shown", words: "toast pop-up max count number how many recent" },
    { tab: "chat", section: "pop-ups", label: "Show images", words: "picture photo gif giphy embed media thumbnail chat box pop-up" },
    { tab: "chat", section: "pop-ups", label: "Also load image links", words: "picture photo gif giphy embed link url address site ip privacy" },
    { tab: "chat", section: "pop-ups", label: "Link previews", words: "embed card youtube video twitter x tweet post page link url site ip privacy" },
    { tab: "overlay", section: "appearance", label: "Chat theme", words: "style theme look legacy minimal light cs2" },
    { tab: "overlay", section: "appearance", label: "Show avatars", words: "pictures" },
    { tab: "overlay", section: "appearance", label: "Size", words: "scale zoom" },
    { tab: "overlay", section: "animation", label: "Animations", words: "motion" },
    { tab: "overlay", section: "animation", label: "Animation style", words: "fade slide typewriter" },
    { tab: "overlay", section: "animation", label: "Animation speed", words: "fast normal slow" },
    { tab: "overlay", section: "position", label: "Distance from bottom", words: "offset place" },
    { tab: "overlay", section: "position", label: "Preview in game position", words: "test pop-up toast" },
    { tab: "help", section: "commands", label: "Commands", words: "/r /w /help /clear whisper reply" },
    { tab: "help", section: "keys", label: "Keys", words: "enter esc tab arrows page paste" },
    { tab: "help", section: "about", label: "Quit myChat", words: "exit close about" },
  ];

  let searching = $state(false);
  let query = $state("");
  let searchBox = $state<HTMLInputElement>();

  const results = $derived.by(() => {
    const words = query.toLowerCase().split(/\s+/).filter(Boolean);
    return INDEX.filter((e) => {
      const hay = `${e.label} ${e.words ?? ""} ${e.tab} ${e.section}`.toLowerCase();
      return words.every((w) => hay.includes(w));
    });
  });

  const tabLabel = (id: TabId): string => TABS.find((t) => t.id === id)?.label ?? id;
  const sectionLabel = (id: TabId, sec: string): string =>
    (TABS.find((t) => t.id === id)?.sections as readonly { id: string; label: string }[] | undefined)?.find(
      (s) => s.id === sec,
    )?.label ?? sec;

  async function toggleSearch(): Promise<void> {
    searching = !searching;
    query = "";
    if (searching) {
      await tick();
      searchBox?.focus();
    }
  }

  async function openResult(e: (typeof INDEX)[number]): Promise<void> {
    searching = false;
    query = "";
    pickTab(e.tab);
    await tick();
    pickSection(e.section);
  }

  function searchKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.preventDefault();
      searching = false;
      query = "";
    } else if (e.key === "Enter" && results.length > 0) {
      e.preventDefault();
      void openResult(results[0]);
    }
  }

  let notice = $state<{ text: string; bad: boolean } | null>(null);

  const canReset = $derived(!searching && (tab === "chat" || tab === "overlay"));

  async function resetDefaults(): Promise<void> {
    let patch: Partial<Cfg.Config>;
    if (tab === "chat") {
      patch = {
        closeAfterSend: true,
        doubleEnterClose: true,
        showToasts: true,
        toastSeconds: 6,
        toastLayout: "group",
        toastMax: 3,
        showImages: true,
        imageLinks: false,
        linkPreviews: true,
      };
      const h = app.config.hotkey;
      if (h.code !== "KeyI" || h.ctrl || h.shift || h.alt) {
        patch.hotkey = new Cfg.Hotkey({ code: "KeyI" });
      }
    } else {
      patch = {
        overlayStyle: "legacy",
        showAvatars: true,
        blur: false,
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

  const steam = $derived.by(() => {
    const s = app.status;
    if (!s.accountName || (s.state === "needs_login" && app.login.phase !== "done")) {
      return { tone: "off", ring: "none", name: "Not signed in", text: "Steam" } as const;
    }
    const name = s.personaName || (app.config.showAccountName ? s.accountName : "") || "Steam";
    switch (s.state) {
      case "online":
        return { tone: "in", ring: "accent", name, text: "Logged in" } as const;
      case "connecting":
      case "needs_login":
        return { tone: "warn", ring: "accent", name, text: "Connecting" } as const;
      default:
        return { tone: "bad", ring: "offline", name, text: "Offline" } as const;
    }
  });

  const topo = URL.createObjectURL(new Blob([topoTile()], { type: "image/svg+xml" }));
  onDestroy(() => URL.revokeObjectURL(topo));
</script>

<div class="settings" class:anim={app.config.animations} style:--topo="url({topo})">
  <header class="top">
    <div class="mark">
      <span>myChat</span>
    </div>

    <nav class="nav" aria-label="myChat">
      <button
        type="button"
        class="navitem"
        class:active={!inHelp}
        aria-current={!inHelp ? "page" : undefined}
        onclick={() => pickTab(lastSettingsTab)}
      >
        Settings
      </button>
      <button
        type="button"
        class="navitem"
        class:active={inHelp}
        aria-current={inHelp ? "page" : undefined}
        onclick={() => pickTab("help")}
      >
        Help
      </button>
    </nav>

    <button type="button" class="steam" aria-label="Steam account" onclick={() => pickTab("account")}>
      <span class="who">
        <bdi class="name">{steam.name}</bdi>
        <span class="state {steam.tone}"><span class="dot {steam.tone}"></span>{steam.text}</span>
      </span>
      {#if steam.tone === "off"}
        <span class="nobody"><Icon name="user" size={18} /></span>
      {:else}
        <Avatar size={28} ring={steam.ring} src={app.status.avatar} name={steam.name} />
      {/if}
    </button>
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
    <div class="tabrow">
      <button
        type="button"
        class="search"
        class:active={searching}
        aria-label={searching ? "Close search" : "Search settings"}
        aria-pressed={searching}
        onclick={toggleSearch}
      >
        <Icon name="search" size={20} />
      </button>
      {#if searching}
        <input
          class="query"
          type="text"
          placeholder="Search settings"
          spellcheck="false"
          autocomplete="off"
          aria-label="Search settings"
          bind:this={searchBox}
          bind:value={query}
          onkeydown={searchKey}
        />
      {:else if inHelp}
        {#each current.sections as s (s.id)}
          <button
            type="button"
            class="tab"
            class:active={section === s.id}
            aria-current={section === s.id ? "true" : undefined}
            onclick={() => pickSection(s.id)}
          >
            {s.label}
          </button>
        {/each}
      {:else}
        <div class="tablist" role="tablist" aria-label="Settings">
          {#each SETTINGS_TABS as t (t.id)}
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
      {/if}
    </div>

    {#if !searching && !inHelp}
      <div class="tabrow second">
        {#each current.sections as s (s.id)}
          <button
            type="button"
            class="tab"
            class:active={section === s.id}
            aria-current={section === s.id ? "true" : undefined}
            onclick={() => pickSection(s.id)}
          >
            {s.label}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <main class="content" bind:this={scroller} onscroll={spy}>
    <div class="column">
      {#if searching}
        <div class="page">
          <h2 class="found">{query.trim() ? "Search results" : "All settings"}</h2>
          {#each results as r (r.tab + r.section + r.label)}
            <button type="button" class="result" onclick={() => openResult(r)}>
              <span class="rlabel">{r.label}</span>
              <span class="rwhere">{tabLabel(r.tab)} / {sectionLabel(r.tab, r.section)}</span>
              <span class="rgo"><Icon name="chevron" /></span>
            </button>
          {:else}
            <p class="none">Nothing matches “{query.trim()}”.</p>
          {/each}
        </div>
      {:else}
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
      {/if}
    </div>
  </main>

  <footer class="bottom">
    <div class="notice" class:bad={notice?.bad} role="status">
      {notice ? notice.text : "Changes save automatically."}
    </div>
    {#if canReset}
      <button type="button" class="reset" onclick={resetDefaults}>
        <Icon name="history" size={18} />
        <span>Reset</span>
      </button>
    {/if}
  </footer>
</div>

<style>
  .settings {
    position: fixed;
    inset: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: 52px auto minmax(0, 1fr) 52px;
    grid-template-areas:
      "top"
      "tabs"
      "content"
      "bottom";
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    line-height: 20px;
    color: var(--text);
    background:
      radial-gradient(44% 50% at 14% 36%, rgba(15, 19, 24, 0.7), transparent 72%),
      radial-gradient(38% 44% at 84% 66%, rgba(15, 19, 24, 0.62), transparent 72%),
      radial-gradient(30% 26% at 50% 20%, rgba(15, 19, 24, 0.4), transparent 72%),
      radial-gradient(46% 42% at 54% 104%, rgba(176, 186, 196, 0.36), transparent 72%),
      radial-gradient(36% 34% at 88% 12%, rgba(140, 156, 172, 0.3), transparent 72%),
      radial-gradient(32% 36% at 40% 58%, rgba(128, 144, 160, 0.24), transparent 72%),
      #3c4651;
  }

  .top {
    --wails-draggable: drag;
    grid-area: top;
    position: relative;
    display: flex;
    align-items: center;
    padding-left: 18px;
    background:
      var(--topo) 0 0 / 560px 560px,
      rgba(34, 43, 52, 0.82);
    border-bottom: 1px solid rgba(0, 0, 0, 0.35);
  }

  .mark {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 15px;
    font-weight: 600;
    line-height: 24px;
    letter-spacing: -0.01em;
    color: #fff;
  }

  .nav {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 50%;
    display: flex;
    align-items: center;
    transform: translateX(-50%);
  }

  .navitem {
    --wails-draggable: no-drag;
    position: relative;
    height: 100%;
    padding: 0 20px;
    font-family: var(--font-ui);
    font-size: 16px;
    font-weight: 600;
    line-height: 52px;
    letter-spacing: -0.01em;
    color: #c3cad1;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .navitem + .navitem::before {
    content: "";
    position: absolute;
    top: 18px;
    bottom: 18px;
    left: 0;
    width: 1px;
    background: rgba(255, 255, 255, 0.3);
  }

  .navitem:hover {
    color: #fff;
  }

  .navitem.active {
    color: var(--accent);
  }

  .steam {
    --wails-draggable: no-drag;
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    height: 100%;
    margin-left: auto;
    padding: 0 12px;
    text-align: right;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .nobody {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: #fff;
    background: var(--tile);
  }

  .who {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    min-width: 0;
  }

  .name {
    max-width: 170px;
    overflow: hidden;
    font-size: var(--fs-sm);
    font-weight: 600;
    line-height: 17px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: #fff;
  }

  .state {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    line-height: 14px;
    color: var(--muted);
  }

  .dot {
    flex: none;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--dim);
  }

  .state.in {
    color: var(--accent);
  }

  .dot.in {
    background: var(--accent);
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

  .window {
    display: flex;
    height: 100%;
  }

  .window button {
    --wails-draggable: no-drag;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 100%;
    padding: 0;
    color: #c9d0d6;
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

  .tabs {
    grid-area: tabs;
  }

  .tabrow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    height: 44px;
    padding: 0 16px;
    background: rgba(9, 12, 15, 0.42);
  }

  .tabrow.second {
    height: 40px;
    background: none;
  }

  .tablist {
    display: flex;
    gap: 6px;
  }

  .tab {
    height: 28px;
    padding: 0 12px;
    font-family: var(--font-ui);
    font-size: var(--fs-sm);
    font-weight: 500;
    line-height: 28px;
    white-space: nowrap;
    color: var(--text-soft);
    background: none;
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  .tablist .tab {
    font-size: var(--fs-md);
    font-weight: 600;
  }

  .tab:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.12);
  }

  .tab.active {
    font-weight: 600;
    color: var(--accent);
    background: var(--tab-fill);
  }

  .search {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    margin-right: 8px;
    padding: 0;
    color: #fff;
    background: none;
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  .search:hover,
  .search.active {
    background: rgba(255, 255, 255, 0.2);
  }

  .query {
    width: 360px;
    height: 30px;
    padding: 0 10px;
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    color: #fff;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
    outline: none;
  }

  .query::placeholder {
    color: var(--muted);
  }

  .query:focus {
    border-color: var(--accent);
  }

  .content {
    grid-area: content;
    overflow-x: hidden;
    overflow-y: auto;
    scrollbar-gutter: stable both-edges;
  }

  .content::-webkit-scrollbar {
    width: 10px;
  }

  .content::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.22);
    border: 3px solid transparent;
    background-clip: content-box;
  }

  .content::-webkit-scrollbar-thumb:hover {
    background-color: rgba(255, 255, 255, 0.4);
  }

  .column {
    width: min(732px, 100% - 48px);
    min-height: 100%;
    margin: 0 auto;
    padding: 0 0 40px;
  }

  .found {
    margin: 0;
    padding: 24px 0 6px;
    font-size: var(--fs-lg);
    font-weight: 600;
    line-height: 28px;
    letter-spacing: -0.01em;
    color: #fff;
  }

  .result {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    min-height: 48px;
    padding: 0 4px 0 0;
    text-align: left;
    background: none;
    border: 0;
    border-bottom: 1px solid var(--line);
    cursor: pointer;
  }

  .rlabel {
    flex: 1;
    min-width: 0;
    font-size: var(--fs-md);
    font-weight: 500;
    color: var(--text);
  }

  .rwhere {
    font-size: var(--fs-sm);
    font-weight: 500;
    white-space: nowrap;
    color: var(--muted);
  }

  .rgo {
    color: #fff;
    transform: rotate(-90deg);
  }

  .result:hover .rlabel,
  .result:hover .rwhere,
  .result:hover .rgo {
    color: var(--accent);
  }

  .none {
    margin: 0;
    padding: 14px 0;
    font-size: var(--fs-md);
    color: var(--muted);
    border-bottom: 1px solid var(--line);
  }

  .bottom {
    grid-area: bottom;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 0 18px 0 24px;
    border-top: 1px solid var(--line-soft);
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

  .reset {
    flex: none;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 32px;
    padding: 0 10px;
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    font-weight: 600;
    color: #fff;
    border-radius: 2px;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .reset:hover {
    color: var(--accent);
  }

  button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .anim .navitem,
  .anim .tab,
  .anim .search,
  .anim .reset,
  .anim .window button {
    transition:
      color 120ms,
      background-color 120ms;
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
