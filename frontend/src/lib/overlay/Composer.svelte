<script lang="ts">
  // The chat box (640×360). There is no <input>: the window never has focus,
  // so the text and the caret are drawn from ComposerState.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import type * as M from "../../../bindings/mychat/models";
  import Avatar from "../ui/Avatar.svelte";
  import ChatLine from "./ChatLine.svelte";

  let {
    composer,
    lines,
    status,
    live,
    avatars = false,
    typewriter = false,
  }: {
    composer: C.ComposerState;
    lines: C.Entry[];
    status: M.Status;
    /** The composer is the current mode (not fading out). */
    live: boolean;
    /** Show friends' avatars (setting `showAvatars`). */
    avatars?: boolean;
    /** Animation style is "typewriter". Never applies to the input line. */
    typewriter?: boolean;
  } = $props();

  const channel = $derived(composer.channel);
  const kind = $derived(
    channel.kind === "whisper" || channel.kind === "party" ? channel.kind : "none",
  );

  // `caret` counts code points, not UTF-16 units.
  const chars = $derived(Array.from(composer.text));
  const caret = $derived(Math.min(Math.max(composer.caret, 0), chars.length));
  const before = $derived(chars.slice(0, caret).join(""));
  const after = $derived(chars.slice(caret).join(""));

  const suggestions = $derived((composer.suggestions ?? []).slice(0, 6));

  const steam = $derived.by(() => {
    switch (status.state) {
      case "online":
        return null;
      case "connecting":
        return { tone: "warn", text: "Connecting to Steam…" };
      case "needs_login":
        return { tone: "bad", text: "Log in from the tray icon" };
      default:
        return { tone: "bad", text: "Steam offline – reconnecting" };
    }
  });

  // ---- keep the caret in view when the text is longer than the field ----

  const EDGE = 24;
  let fieldEl = $state<HTMLElement>();
  let textEl = $state<HTMLElement>();
  let caretEl = $state<HTMLElement>();
  let shift = $state(0);
  /** Flips on every edit to restart the blink, so the caret is solid while typing. */
  let beat = $state(false);

  $effect(() => {
    // Re-measure after anything that moves the caret or resizes the field.
    void composer.text;
    void composer.caret;
    void channel.kind;
    void channel.targetName;
    void avatars;
    if (!fieldEl || !textEl || !caretEl) return;

    const view = fieldEl.clientWidth;
    const x = caretEl.offsetLeft;
    const overflow = textEl.offsetWidth - view;
    let s = untrack(() => shift);
    if (x - s > view - EDGE) s = x - (view - EDGE);
    if (x - s < EDGE) s = x - EDGE;
    shift = Math.round(Math.max(0, Math.min(s, overflow + EDGE)));
    beat = !untrack(() => beat);
  });
</script>

<div class="composer" class:live>
  <div class="log">
    <div class="lines">
      {#each lines as entry (entry.id)}
        <ChatLine {entry} {avatars} {typewriter} />
      {:else}
        <div class="empty">No messages yet.</div>
      {/each}
    </div>

    {#if composer.scroll > 0}
      <div class="newer head">↓ {composer.scroll} newer · PgDn</div>
    {/if}

    {#if suggestions.length > 0 || composer.error}
      <div class="pop">
        {#if suggestions.length > 0}
          <div class="suggest">
            <div class="shead head">
              <span>Whisper to</span>
              <span>{composer.hint}</span>
            </div>
            {#each suggestions as s, i (s.id)}
              <div class="srow" class:sel={i === composer.selected}>
                {#if avatars}
                  <Avatar size={22} ring={s.online ? "green" : "offline"} src={s.avatar} name={s.name} />
                {:else}
                  <span class="dot" class:on={s.online}></span>
                {/if}
                <bdi class="sname">{s.name}</bdi>
                <span class="sgame">{s.inGame || (s.online ? "Online" : "Offline")}</span>
              </div>
            {/each}
          </div>
        {/if}
        {#if composer.error}
          <div class="error">
            <svg viewBox="0 0 12 12" width="12" height="12" aria-hidden="true">
              <path d="M6 1l5.5 10h-11z" fill="none" stroke="currentColor" stroke-width="1.2" />
              <path d="M6 5v3M6 9.2v.6" fill="none" stroke="currentColor" stroke-width="1.2" />
            </svg>
            <span>{composer.error}</span>
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <div class="input {kind}">
    <div class="label">
      {#if kind === "whisper"}
        {#if avatars}
          <Avatar
            size={18}
            ring={channel.online ? "whisper" : "offline"}
            src={channel.avatar}
            name={channel.targetName}
          />
        {:else}
          <span class="dot" class:on={channel.online}></span>
        {/if}
        <span>To</span>
        <bdi class="target">{channel.targetName}</bdi><span class="colon">:</span>
      {:else if kind === "party"}
        {#if avatars}
          <svg class="group" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
            <path
              d="M5.5 7a2.25 2.25 0 1 0 0-4.5 2.25 2.25 0 0 0 0 4.5z M1.5 13c.3-2.4 1.9-3.6 4-3.6s3.7 1.2 4 3.6 M10.4 6.9a2 2 0 1 0-.5-3.8 M11.2 9.5c1.8.3 3 1.5 3.3 3.5"
            />
          </svg>
        {/if}
        <span>Party</span><span class="colon">:</span>
      {:else}
        <span>Chat</span><span class="colon">:</span>
      {/if}
    </div>

    <div class="field" class:scrolled={shift > 0} bind:this={fieldEl}>
      <div
        class="text"
        bind:this={textEl}
        style:transform={shift > 0 ? `translateX(${-shift}px)` : null}
      >
        <span>{before}</span><span class="caret" class:beat bind:this={caretEl}></span><span
          >{after}</span
        >{#if chars.length === 0 && composer.hint}<span class="hint">{composer.hint}</span>{/if}
      </div>
    </div>
  </div>

  <div class="foot">
    <div class="steam">
      {#if steam}
        <span class="dot {steam.tone}"></span>
        <span>{steam.text}</span>
      {/if}
    </div>
    <!-- Key hints as in the CS2 scoreboard footer: small boxed keys. -->
    <div class="keys">
      <span><kbd>Enter</kbd>Send</span>
      <span><kbd>Esc</kbd>Close</span>
      <span><kbd>Tab</kbd>Next</span>
      <span><kbd>/help</kbd></span>
    </div>
  </div>
</div>

<style>
  .composer {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
  }

  /* ---------- log ---------- */

  .log {
    position: relative;
    flex: 1;
    min-height: 0;
  }

  .lines {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    overflow: hidden;
    padding: 0 16px 8px;
    /* Old lines melt away towards the top. */
    -webkit-mask-image: linear-gradient(to bottom, transparent 0, #000 56px);
    mask-image: linear-gradient(to bottom, transparent 0, #000 56px);
  }

  .empty {
    font-size: 14px;
    font-style: italic;
    color: var(--muted);
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.85);
  }

  .newer {
    position: absolute;
    right: 12px;
    bottom: 8px;
    height: 20px;
    padding: 0 8px;
    font-size: 12px;
    line-height: 20px;
    color: var(--text);
    background: rgba(10, 12, 15, 0.9);
    box-shadow: inset 0 0 0 1px var(--line-strong);
  }

  /* ---------- suggestions and error, stacked just above the input ---------- */

  .pop {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
  }

  .suggest {
    padding-bottom: 4px;
    background: rgba(10, 12, 15, 0.94);
    border-top: 1px solid var(--line);
  }

  .shead {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    height: 24px;
    padding: 0 16px;
    font-size: 12px;
    line-height: 24px;
    color: var(--muted);
    white-space: nowrap;
  }

  .srow {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 28px;
    padding: 0 16px;
    font-size: 14px;
    line-height: 28px;
    color: var(--text-soft);
    white-space: nowrap;
  }

  /* Faint bands, as in the scoreboard. */
  .srow:nth-child(even) {
    background: rgba(255, 255, 255, 0.04);
  }

  /* The selected row gets the scoreboard's gold highlight. */
  .srow.sel {
    color: #fff;
    background: var(--gold-wash);
    box-shadow:
      inset 3px 0 0 var(--gold),
      inset 0 1px 0 rgba(224, 178, 82, 0.35),
      inset 0 -1px 0 rgba(224, 178, 82, 0.35);
  }

  .srow.sel .sgame {
    color: #d9c08a;
  }

  .sname {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 600;
  }

  .sgame {
    flex: none;
    margin-left: auto;
    padding-left: 16px;
    font-size: 12px;
    color: var(--muted);
  }

  .error {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 26px;
    padding: 0 16px;
    font-size: 13px;
    line-height: 26px;
    color: #ffb1aa;
    white-space: nowrap;
    background: rgba(58, 16, 14, 0.94);
    box-shadow: inset 3px 0 0 var(--red);
  }

  .error svg {
    flex: none;
    color: var(--red-text);
  }

  .error span {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* ---------- online dots ---------- */

  .dot {
    flex: none;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    box-shadow: inset 0 0 0 1px var(--dim);
  }

  .dot.on {
    background: var(--green);
    box-shadow: none;
  }

  .dot.warn {
    background: var(--gold);
    box-shadow: none;
  }

  .dot.bad {
    background: var(--red);
    box-shadow: none;
  }

  /* ---------- input bar ---------- */

  .input {
    --chan: var(--muted);
    position: relative;
    flex: none;
    display: flex;
    align-items: center;
    gap: 10px;
    height: 46px;
    padding: 0 16px;
    background: rgba(0, 0, 0, 0.4);
    border-top: 1px solid rgba(255, 255, 255, 0.16);
    box-shadow: inset 3px 0 0 var(--chan);
  }

  .input.whisper {
    --chan: var(--whisper);
  }

  .input.party {
    --chan: var(--party);
  }

  .label {
    flex: none;
    display: flex;
    align-items: center;
    gap: 6px;
    max-width: 50%;
    font-family: var(--font-head);
    font-size: 19px;
    font-weight: 700;
    line-height: 24px;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--chan);
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  }

  .target {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .colon {
    margin-left: -6px;
  }

  .field {
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  .field.scrolled {
    -webkit-mask-image: linear-gradient(to right, transparent 0, #000 20px);
    mask-image: linear-gradient(to right, transparent 0, #000 20px);
  }

  .text {
    position: relative;
    display: inline-block;
    vertical-align: top;
    font-size: 16px;
    line-height: 24px;
    white-space: pre;
    color: #fff;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  }

  .hint {
    color: var(--muted);
  }

  /* A drawn caret: a thin bar that takes no room in the line. */
  .caret {
    display: inline-block;
    width: 2px;
    height: 20px;
    margin: 0 -1px;
    vertical-align: -4px;
    background: #fff;
  }

  /* Blink is CSS only, and only while the composer is open with animations
     on. Two identical keyframe sets let each edit restart it. */
  :global(.anim) .live .caret {
    animation: blink-a 1.1s step-end infinite;
  }

  :global(.anim) .live .caret.beat {
    animation-name: blink-b;
  }

  @keyframes blink-a {
    0%,
    100% {
      opacity: 1;
    }
    55% {
      opacity: 0;
    }
  }

  @keyframes blink-b {
    0%,
    100% {
      opacity: 1;
    }
    55% {
      opacity: 0;
    }
  }

  /* ---------- footer strip ---------- */

  .foot {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    height: 24px;
    padding: 0 16px;
    font-size: 12px;
    line-height: 24px;
    white-space: nowrap;
    color: var(--muted);
    background: rgba(0, 0, 0, 0.5);
    border-top: 1px solid var(--line-soft);
  }

  .steam {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    color: var(--text-soft);
  }

  .keys {
    flex: none;
    display: flex;
    align-items: center;
    gap: 12px;
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .keys span {
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .keys kbd {
    height: 16px;
    padding: 0 5px;
    font: inherit;
    font-size: 12px;
    font-weight: 700;
    line-height: 16px;
    color: var(--text);
    background: rgba(255, 255, 255, 0.08);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.34);
  }

  .group {
    flex: none;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
  }

  /* ---------- motion ---------- */

  :global(.anim) .pop {
    animation: pop-in calc(var(--dur) * 0.6) var(--ease-out) both;
  }

  @keyframes pop-in {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
</style>
