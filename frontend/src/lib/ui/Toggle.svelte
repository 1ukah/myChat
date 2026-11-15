<script lang="ts">
  // On/off switch: CS2 lobby green with a white knob, value text beside it.
  let {
    checked,
    label,
    disabled = false,
    onchange,
  }: {
    checked: boolean;
    /** Accessible name (the visible label lives in the row). */
    label: string;
    disabled?: boolean;
    onchange: (checked: boolean) => void;
  } = $props();
</script>

<button
  type="button"
  class="toggle"
  class:on={checked}
  role="switch"
  aria-checked={checked}
  aria-label={label}
  {disabled}
  onclick={() => onchange(!checked)}
>
  <span class="state">{checked ? "Enabled" : "Disabled"}</span>
  <span class="track"><span class="knob"></span></span>
</button>

<style>
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    height: 32px;
    padding: 0;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .state {
    font-family: var(--font-head);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .on .state {
    color: var(--text);
  }

  .track {
    position: relative;
    width: 44px;
    height: 22px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }

  .on .track {
    background: var(--green);
    border-color: #6dcc76;
  }

  .knob {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    background: #fff;
    border-radius: 1px;
    opacity: 0.7;
  }

  .on .knob {
    opacity: 1;
    transform: translateX(22px);
  }

  .toggle:hover .track {
    border-color: rgba(255, 255, 255, 0.6);
  }

  .toggle:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .toggle:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 4px;
  }

  :global(.anim) .knob {
    transition:
      transform 120ms var(--ease-out),
      opacity 120ms;
  }

  :global(.anim) .track {
    transition:
      background-color 120ms,
      border-color 120ms;
  }
</style>
