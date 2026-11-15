<script lang="ts">
  // The chat box (640×360), built like CS2's own: a translucent black frame
  // holding the log box and a separate input row with a grey SEND, and key
  // hints under it. There is no <input>: the window never has focus, so the
  // text and the caret are drawn from ComposerState.
  import { untrack } from "svelte";
  import type * as C from "../../../bindings/mychat/internal/chat/models";
  import type * as M from "../../../bindings/mychat/models";
  import Avatar from "../ui/Avatar.svelte";
  import ChatLine from "./ChatLine.svelte";

  let {
    composer,
    lines,
    total = 0,
    status,
    live,
    avatars = false,
    minimal = false,
    typewriter = false,
  }: {
    composer: C.ComposerState;
    lines: C.Entry[];
    /** Entries in the whole log (for the scrollbar). */
    total?: number;
    status: M.Status;
    /** The composer is the current mode (not fading out). */
    live: boolean;
    /** Show friends' avatars (setting `showAvatars`). */
    avatars?: boolean;
    /** The Minimal look (`overlayStyle`): light, compact, rounded panels. */
    minimal?: boolean;
    /** Animation style is "typewriter". Never applies to the input line. */
    typewriter?: boolean;
  } = $props();

  const channel = $derived(composer.channel);
  const kind = $derived(channel.kind === "whisper" ? "whisper" : "none");

  // `caret` counts code points, not UTF-16 units.
  const chars = $derived(Array.from(composer.text));
  const caret = $derived(Math.min(Math.max(composer.caret, 0), chars.length));
  const before = $derived(chars.slice(0, caret).join(""));
  const after = $derived(chars.slice(caret).join(""));

  const suggestions = $derived((composer.suggestions ?? []).slice(0, 6));

  // CS2 writes "Say to all" in the empty input.
  const placeholder = $derived(
    composer.hint ||
      (kind === "whisper" ? `Say to ${channel.targetName}` : "Type /w name to whisper"),
  );

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

  // ---- thin scrollbar, drawn from the scroll position Go reports ----

  /** About how many lines fit in the log box (the Minimal one is lower). */
  const fit = $derived(minimal ? 10 : 12);
  const bar = $derived.by(() => {
    if (total <= fit) return null;
    const size = Math.max(14, (fit / total) * 100);
    const scroll = Math.min(Math.max(composer.scroll, 0), total - 1);
    return { size, bottom: Math.min(100 - size, (scroll / total) * 100) };
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
    void minimal;
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

<div class="composer" class:live class:minimal>
  <div class="frame">
    <div class="log">
      <div class="lines">
        {#each lines as entry (entry.id)}
          <ChatLine {entry} {avatars} {minimal} {typewriter} />
        {:else}
          <div class="empty">No messages yet.</div>
        {/each}
      </div>

      {#if bar}
        <div class="bar" style:height="{bar.size}%" style:bottom="{bar.bottom}%"></div>
      {/if}

      {#if composer.scroll > 0}
        <div class="newer">↓ {composer.scroll} newer · {minimal ? "PgDn" : "[PGDN]"}</div>
      {/if}

      {#if suggestions.length > 0 || composer.error}
        <div class="pop">
          {#if suggestions.length > 0}
            <div class="suggest">
              <div class="shead">
                <span>Whisper to</span>
                <span>{composer.hint}</span>
              </div>
              {#each suggestions as s, i (s.id)}
                <div class="srow" class:sel={i === composer.selected} class:off={!s.online}>
                  {#if avatars}
                    <Avatar size={20} ring={s.online ? "green" : "offline"} src={s.avatar} name={s.name} />
                  {:else}
                    <span class="dot" class:on={s.online}></span>
                  {/if}
                  <bdi class="sname">{s.name}</bdi>
                  <!-- A friend with a nickname: the Steam name, small, next to it. -->
                  {#if s.altName}<bdi class="salt">{s.altName}</bdi>{/if}
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
      {#if kind === "whisper"}
        <div class="label">
          <span>{minimal ? "To" : "[TO]"}</span>
          {#if avatars}
            <!-- Square, like the avatars in the chat lines; the border tells
                 whether the person is online. -->
            <Avatar
              size={22}
              ring={channel.online ? "whisper" : "offline"}
              src={channel.avatar}
              name={channel.targetName}
            />
          {:else}
            <span class="dot" class:on={channel.online}></span>
          {/if}
          <bdi class="target">{channel.targetName}</bdi><span class="colon">:</span>
        </div>
      {/if}

      <div class="field" class:scrolled={shift > 0} bind:this={fieldEl}>
        <div
          class="text"
          bind:this={textEl}
          style:transform={shift > 0 ? `translateX(${-shift}px)` : null}
        >
          <span>{before}</span><span class="caret" class:beat bind:this={caretEl}></span><span
            >{after}</span
          >{#if chars.length === 0}<span class="hint">{placeholder}</span>{/if}
        </div>
      </div>

      <div class="send" class:ready={chars.length > 0}>Send</div>
    </div>
  </div>

  <div class="foot">
    {#if steam}
      <div class="steam">
        <span class="dot {steam.tone}"></span>
        <span>{steam.text}</span>
      </div>
    {/if}
    <!-- Key hints as CS2 writes them: "[CTRL] Close". Minimal: "Enter send". -->
    <div class="keys">
      {#if minimal}
        <span><b>Enter</b> send</span>
        <span><b>Esc</b> close</span>
        <span><b>Tab</b> next</span>
        <span><b>/help</b></span>
      {:else}
        <span>[ENTER] Send</span>
        <span>[ESC] Close</span>
        <span>[TAB] Next</span>
        <span>/help</span>
      {/if}
    </div>
  </div>
</div>

<style>
  .composer {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    font-family: var(--font-head);
  }

  /* The outer frame: a lighter black that holds the two darker boxes. */
  .frame {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 6px;
    background: var(--frame-fill);
  }

  /* ---------- log ---------- */

  .log {
    position: relative;
    flex: 1;
    min-height: 0;
    background: var(--box-fill);
  }

  .lines {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    overflow: hidden;
    padding: 0 18px 6px 12px;
  }

  .empty {
    font-size: 15px;
    font-weight: 600;
    line-height: 21px;
    color: var(--muted);
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.9);
  }

  /* CS2's thin light scrollbar at the right edge. */
  .bar {
    position: absolute;
    right: 5px;
    width: 4px;
    margin: 6px 0;
    background: rgba(255, 255, 255, 0.55);
  }

  .newer {
    position: absolute;
    right: 16px;
    bottom: 6px;
    height: 20px;
    padding: 0 8px;
    font-size: 12px;
    font-weight: 700;
    line-height: 20px;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: var(--accent);
    background: var(--tab-fill);
  }

  /* ---------- suggestions and error, stacked just above the input ---------- */

  .pop {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
  }

  .suggest {
    padding-bottom: 2px;
    /* Opaque: the log lines behind must not show through the names. */
    background: #121417;
  }

  .shead {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    height: 24px;
    padding: 0 12px;
    font-size: 12px;
    font-weight: 700;
    line-height: 24px;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: var(--muted);
    white-space: nowrap;
  }

  .srow {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 26px;
    margin: 0 4px 2px;
    padding: 0 8px;
    font-size: 15px;
    line-height: 26px;
    color: var(--text);
    white-space: nowrap;
    /* Row bands, as in the scoreboard. */
    background: rgba(255, 255, 255, 0.07);
  }

  .srow.off {
    color: var(--muted);
  }

  /* The selected row is the scoreboard's highlighted player. */
  .srow.sel {
    color: #fff;
    background: var(--row-pick);
    box-shadow: inset 0 0 0 1px var(--row-pick-line);
  }

  .sname {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 700;
  }

  /* The Steam name behind a nickname. */
  .salt {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 12px;
    font-weight: 600;
    color: var(--muted);
  }

  .sgame {
    flex: none;
    margin-left: auto;
    padding-left: 16px;
    font-size: 13px;
    font-weight: 600;
    color: var(--muted);
  }

  .srow.sel .sgame {
    color: var(--ct);
  }

  .error {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 26px;
    padding: 0 12px;
    font-size: 14px;
    font-weight: 700;
    line-height: 26px;
    color: #ffc2bc;
    white-space: nowrap;
    background: #47110f;
    box-shadow: inset 3px 0 0 var(--red-text);
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
    width: 8px;
    height: 8px;
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
    background: var(--red-text);
    box-shadow: none;
  }

  /* ---------- input row ---------- */

  .input {
    --chan: var(--muted);
    flex: none;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 44px;
    padding: 0 16px 0 12px;
    background: var(--box-fill);
  }

  .input.whisper {
    --chan: var(--whisper);
  }

  .label {
    flex: none;
    display: flex;
    align-items: center;
    gap: 6px;
    max-width: 50%;
    font-size: 16px;
    font-weight: 700;
    line-height: 24px;
    white-space: nowrap;
    color: var(--chan);
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.9);
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
    font-weight: 600;
    line-height: 24px;
    white-space: pre;
    color: #fff;
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.9);
  }

  /* "Say to all": regular-weight grey. */
  .hint {
    font-family: var(--font-body);
    font-size: 15px;
    font-weight: 400;
    color: var(--muted);
    text-shadow: none;
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

  .send {
    flex: none;
    padding-left: 8px;
    font-size: 17px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: #868c93;
  }

  .send.ready {
    color: #fff;
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

  /* ---------- key hints under the box, straight over the game ---------- */

  .foot {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    height: 22px;
    padding: 0 8px;
    font-size: 12px;
    font-weight: 700;
    line-height: 22px;
    white-space: nowrap;
    color: #e6e9ec;
    /* An outline, so the hints read over bright and dark scenes alike. */
    text-shadow:
      1px 1px 0 #000,
      -1px 1px 0 #000,
      1px -1px 0 #000,
      -1px -1px 0 #000,
      0 0 3px rgba(0, 0, 0, 0.9);
  }

  .steam {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .steam .dot {
    box-shadow: 0 0 0 1px #000;
  }

  .keys {
    flex: none;
    display: flex;
    gap: 14px;
    margin-left: auto;
  }

  /* ---------- Minimal look ----------
     Light and airy: two small frosted-white cards (log, input) with dark
     text, hairline edges, a soft shadow and round corners. Nothing blurs the
     game here, so the frost is a dense white gradient: it stays light over a
     dark scene and the edge and shadow keep it visible over a bright one.
     The log card is only as tall as its lines, up to ten of them. */

  .minimal {
    --avatar-radius: 3px;
    --ink: #151c26;
    --blue: #1f62ad;
    --muted: #677383;
    --dim: #9aa5b3;
    --whisper: var(--blue);
    --green: #2f9e58;
    --gold: #c98a12;
    --red-text: #b8322a;
    --tile: #66758a;
    --glass: linear-gradient(rgba(255, 255, 255, 0.92), rgba(243, 247, 252, 0.82));
    --glass-edge:
      inset 0 0 0 1px rgba(255, 255, 255, 0.8), 0 0 0 1px rgba(24, 38, 62, 0.1),
      0 6px 18px rgba(14, 24, 44, 0.2);
    align-items: center;
    justify-content: flex-end;
    font-family: var(--font-body);
  }

  .minimal .frame {
    flex: none;
    gap: 5px;
    width: 520px;
    padding: 0;
    background: none;
  }

  .minimal .log {
    flex: none;
    background: var(--glass);
    border-radius: 12px;
    box-shadow: var(--glass-edge);
  }

  .minimal .lines {
    position: relative;
    inset: auto;
    max-height: 214px;
    padding: 7px 16px 7px 12px;
  }

  .minimal .empty {
    font-size: 14px;
    font-weight: 400;
    line-height: 20px;
    text-shadow: none;
  }

  .minimal .bar {
    right: 4px;
    width: 3px;
    background: rgba(24, 38, 62, 0.28);
    border-radius: 2px;
  }

  .minimal .newer {
    right: 12px;
    height: 18px;
    padding: 0 9px;
    font-size: 11px;
    font-weight: 600;
    line-height: 18px;
    letter-spacing: 0;
    text-transform: none;
    color: #fff;
    background: var(--blue);
    border-radius: 9px;
  }

  .minimal .pop {
    left: 6px;
    right: 6px;
    bottom: 6px;
  }

  .minimal .suggest {
    padding: 3px 4px 4px;
    background: #fdfeff;
    border-radius: 9px;
    box-shadow:
      0 0 0 1px rgba(24, 38, 62, 0.12),
      0 8px 20px rgba(14, 24, 44, 0.2);
  }

  .minimal .shead {
    height: 22px;
    padding: 0 8px;
    font-size: 11px;
    font-weight: 600;
    line-height: 22px;
    letter-spacing: 0;
    text-transform: none;
  }

  .minimal .srow {
    height: 26px;
    margin: 0 0 2px;
    font-size: 14px;
    color: var(--ink);
    background: none;
    border-radius: 6px;
  }

  .minimal .srow:last-child {
    margin-bottom: 0;
  }

  .minimal .srow.off {
    color: var(--muted);
  }

  .minimal .srow.sel {
    color: #0f3f78;
    background: rgba(31, 98, 173, 0.13);
    box-shadow: inset 0 0 0 1px rgba(31, 98, 173, 0.26);
  }

  .minimal .salt,
  .minimal .sgame {
    font-size: 12px;
    font-weight: 400;
  }

  .minimal .srow.sel .sgame {
    color: var(--blue);
  }

  .minimal .error {
    margin-top: 4px;
    font-size: 13px;
    font-weight: 600;
    color: #a3271f;
    background: #fdeceb;
    border-radius: 8px;
    box-shadow:
      0 0 0 1px rgba(184, 50, 42, 0.24),
      0 4px 12px rgba(14, 24, 44, 0.16);
  }

  .minimal .input {
    gap: 7px;
    height: 38px;
    padding: 0 13px 0 11px;
    background: var(--glass);
    border-radius: 12px;
    box-shadow: var(--glass-edge);
  }

  .minimal .label {
    font-size: 14px;
    line-height: 22px;
    text-shadow: none;
  }

  .minimal .text {
    font-size: 14px;
    font-weight: 400;
    line-height: 22px;
    color: var(--ink);
    text-shadow: none;
  }

  .minimal .hint {
    font-size: 14px;
  }

  .minimal .caret {
    height: 18px;
    background: var(--blue);
  }

  .minimal .send {
    font-size: 12px;
    letter-spacing: 0;
    text-transform: none;
    color: #a2acb9;
  }

  .minimal .send.ready {
    color: var(--blue);
  }

  /* The hints sit in small pills, so they read over any scene. */
  .minimal .foot {
    width: 520px;
    height: 24px;
    padding: 0 2px;
    font-size: 11px;
    font-weight: 400;
    line-height: 18px;
    color: #4b5767;
    text-shadow: none;
  }

  .minimal .steam,
  .minimal .keys {
    height: 18px;
    margin-top: 4px;
    padding: 0 9px;
    background: rgba(255, 255, 255, 0.84);
    border-radius: 9px;
    box-shadow:
      0 0 0 1px rgba(24, 38, 62, 0.08),
      0 2px 8px rgba(14, 24, 44, 0.16);
  }

  .minimal .keys {
    gap: 10px;
  }

  .minimal .keys b {
    font-weight: 700;
    color: var(--ink);
  }

  .minimal .steam .dot {
    box-shadow: none;
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
