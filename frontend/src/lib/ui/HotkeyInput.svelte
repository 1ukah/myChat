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
  /* The bound key in a thin bordered box, like a CS2 value box. */
  .hotkey {
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    gap: 14px;
    height: 34px;
    padding: 0;
    font-family: var(--font-cond);
    font-weight: 700;
    letter-spacing: 0.01em;
    white-space: nowrap;
    color: #fff;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .hotkey.capturing {
    min-width: 150px;
    justify-content: center;
    color: var(--accent);
    background: var(--tab-fill);
    border-radius: 2px;
  }

  .hotkey:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .keys {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  kbd {
    min-width: 40px;
    height: 34px;
    padding: 0 10px;
    font: inherit;
    font-size: 20px;
    line-height: 32px;
    text-align: center;
    color: #fff;
    border: 1px solid var(--line-strong);
  }

  .hotkey:hover kbd {
    border-color: var(--accent);
  }

  .plus {
    font-size: 18px;
    color: var(--muted);
  }

  .edit {
    order: -1;
    font-family: var(--font-head);
    font-size: 13px;
    color: var(--muted);
  }

  .hotkey:hover .edit {
    color: var(--accent);
  }

  .ask {
    font-size: 20px;
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
