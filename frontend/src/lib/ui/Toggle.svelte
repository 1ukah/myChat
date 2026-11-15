<script lang="ts">
  // On/off setting, CS2 style: the value in bold condensed type (Yes / No)
  // and a small pill switch like "Look for Competitive" in the friends rail.
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
  <span class="state">{checked ? "Yes" : "No"}</span>
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
    font-family: var(--font-cond);
    font-size: 21px;
    font-weight: 700;
    letter-spacing: 0.01em;
    color: #fff;
  }

  .track {
    position: relative;
    width: 36px;
    height: 18px;
    background: rgba(0, 0, 0, 0.45);
    border-radius: 9px;
  }

  .on .track {
    background: var(--tab-fill);
  }

  .knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 12px;
    height: 12px;
    background: #8d949b;
    border-radius: 50%;
  }

  .on .knob {
    background: var(--accent);
    transform: translateX(18px);
  }

  .toggle:hover .state {
    color: var(--accent);
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
      background-color 120ms;
  }

  :global(.anim) .track,
  :global(.anim) .state {
    transition:
      background-color 120ms,
      color 120ms;
  }
</style>
