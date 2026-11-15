<script lang="ts" generics="T extends string">
  // A row of joined boxes; one is selected (FAST / NORMAL / SLOW).
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
  }

  button {
    min-width: 72px;
    height: 32px;
    padding: 0 14px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--muted);
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid var(--line-strong);
    cursor: pointer;
  }

  /* Neighbours share one border line. */
  button + button {
    margin-left: -1px;
  }

  button:first-child {
    border-radius: 2px 0 0 2px;
  }

  button:last-child {
    border-radius: 0 2px 2px 0;
  }

  button:hover {
    color: var(--text);
  }

  button.on {
    position: relative;
    color: #fff;
    background: var(--accent-fill);
    border-color: var(--accent-line);
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
      background-color 120ms,
      border-color 120ms;
  }
</style>
