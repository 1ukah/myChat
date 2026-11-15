<script lang="ts" generics="T extends string">
  // A row of small tabs; the selected one is the filled steel-blue box
  // (Fast / Normal / Slow).
  let {
    value,
    options,
    label,
    disabled = false,
    onchange,
  }: {
    value: T;
    options: readonly { value: T; label: string }[];
    /** Accessible name (the visible label lives in the row). */
    label: string;
    disabled?: boolean;
    onchange: (value: T) => void;
  } = $props();
</script>

<div class="segmented" role="radiogroup" aria-label={label}>
  {#each options as option (option.value)}
    <button
      type="button"
      role="radio"
      class:on={option.value === value}
      aria-checked={option.value === value}
      {disabled}
      onclick={() => option.value !== value && onchange(option.value)}
    >
      {option.label}
    </button>
  {/each}
</div>

<style>
  .segmented {
    display: inline-flex;
    gap: 4px;
  }

  /* Small tabs, like the CS2 sub-tab row. */
  button {
    height: 28px;
    padding: 0 9px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.01em;
    color: #fff;
    background: none;
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  button:hover {
    background: rgba(255, 255, 255, 0.12);
  }

  button.on {
    color: var(--accent);
    background: var(--tab-fill);
  }

  button:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  button:focus-visible {
    position: relative;
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(.anim) button {
    transition:
      color 120ms,
      background-color 120ms;
  }
</style>
