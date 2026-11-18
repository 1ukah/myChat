<script lang="ts">
  let {
    checked,
    label,
    disabled = false,
    onchange,
  }: {
    checked: boolean;
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
  <span class="track"><span class="knob"></span></span>
</button>

<style>
  .toggle {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0;
    background: none;
    border: 0;
    border-radius: 11px;
    cursor: pointer;
  }

  .track {
    position: relative;
    width: 40px;
    height: 22px;
    background: rgba(0, 0, 0, 0.45);
    border-radius: 11px;
    box-shadow: inset 0 0 0 1px var(--line-soft);
  }

  .on .track {
    background: var(--tab-fill);
    box-shadow: inset 0 0 0 1px var(--accent-line);
  }

  .knob {
    position: absolute;
    top: 4px;
    left: 4px;
    width: 14px;
    height: 14px;
    background: #8d949b;
    border-radius: 50%;
  }

  .on .knob {
    background: var(--accent);
    transform: translateX(18px);
  }

  .toggle:hover .knob {
    background: #b4bac0;
  }

  .toggle.on:hover .knob {
    background: #b9e3ff;
  }

  .toggle.on:hover .track {
    background: var(--tab-fill-hover);
  }

  .toggle:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .toggle:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 0;
  }

  :global(.anim) .knob {
    transition:
      transform 120ms var(--ease-out),
      background-color 120ms;
  }

  :global(.anim) .track {
    transition:
      background-color 120ms,
      box-shadow 120ms;
  }
</style>
