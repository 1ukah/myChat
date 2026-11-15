<script lang="ts">
  // Key-bind control. Click it, then press the new key combination.
  // Pure modifier presses are ignored (keep holding, then press the key);
  // Escape cancels.
  import * as Cfg from "../../../bindings/mychat/internal/config/models";

  let {
    label,
    onpick,
  }: {
    /** Current hotkey as text, e.g. "I" or "Ctrl + F6". */
    label: string;
    onpick: (hotkey: Cfg.Hotkey) => void;
  } = $props();

  let capturing = $state(false);

  const keys = $derived(label ? label.split(" + ") : ["–"]);
  const isModifier = (code: string): boolean => /^(Control|Shift|Alt|Meta|OS)(Left|Right)$/.test(code);

  $effect(() => {
    if (!capturing) return;
    const onkeydown = (e: KeyboardEvent): void => {
      e.preventDefault();
      e.stopPropagation();
      if (!e.code || isModifier(e.code)) return;
      capturing = false;
      if (e.code === "Escape") return;
      onpick(new Cfg.Hotkey({ code: e.code, ctrl: e.ctrlKey, shift: e.shiftKey, alt: e.altKey }));
    };
    // Capture phase: nothing else in the window reacts to the key.
    window.addEventListener("keydown", onkeydown, true);
    return () => window.removeEventListener("keydown", onkeydown, true);
  });
</script>

<button
  type="button"
  class="hotkey"
  class:capturing
  aria-label={capturing ? "Press a key, or Escape to cancel" : `Hotkey ${label}. Click to change.`}
  onclick={() => (capturing = !capturing)}
  onblur={() => (capturing = false)}
>
  {#if capturing}
    <span class="ask">Press a key…</span>
  {:else}
    <span class="keys">
      {#each keys as key, i (i)}
        {#if i > 0}<span class="plus">+</span>{/if}
        <kbd>{key}</kbd>
      {/each}
    </span>
    <span class="edit">Change</span>
  {/if}
</button>

<style>
  .hotkey {
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    gap: 16px;
    min-width: 200px;
    height: 36px;
    padding: 0 12px;
    font-family: var(--font-head);
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--text);
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
    cursor: pointer;
  }

  .hotkey:hover {
    border-color: rgba(255, 255, 255, 0.6);
  }

  .hotkey.capturing {
    justify-content: center;
    color: var(--accent);
    background: var(--accent-fill);
    border-color: var(--accent);
  }

  .hotkey:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .keys {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-right: auto;
  }

  kbd {
    min-width: 24px;
    height: 22px;
    padding: 0 7px;
    font: inherit;
    font-size: 16px;
    line-height: 22px;
    text-align: center;
    color: #14161a;
    background: var(--text);
    border-radius: 2px;
  }

  .plus {
    font-size: 14px;
    color: var(--muted);
  }

  .edit {
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.14em;
    color: var(--muted);
  }

  .hotkey:hover .edit {
    color: var(--accent);
  }

  .ask {
    font-size: 15px;
    letter-spacing: 0.14em;
  }

  :global(.anim) .hotkey.capturing .ask {
    animation: ask 1s ease-in-out infinite;
  }

  @keyframes ask {
    50% {
      opacity: 0.45;
    }
  }
</style>
