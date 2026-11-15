<script lang="ts">
  // One settings row: label on the left, control on the right, thin divider.
  // The row is tall enough for a second line (hint or error), so showing an
  // error never moves the layout.
  import type { Snippet } from "svelte";

  let {
    label,
    hint = "",
    error = "",
    dim = false,
    children,
  }: {
    label: string;
    hint?: string;
    /** Shown in red instead of the hint (for example a rejected hotkey). */
    error?: string;
    /** The setting has no effect right now. */
    dim?: boolean;
    children?: Snippet;
  } = $props();
</script>

<div class="row">
  <div class="info" class:dim>
    <div class="label">{label}</div>
    {#if error}
      <div class="note error" role="alert">{error}</div>
    {:else if hint}
      <div class="note">{hint}</div>
    {/if}
  </div>
  {#if children}<div class="control">{@render children()}</div>{/if}
</div>

<style>
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    min-height: 56px;
    padding: 8px 0;
    /* No fill: only the thin light rule under each row, as in CS2. */
    border-bottom: 1px solid var(--line);
  }

  .info {
    min-width: 0;
  }

  .info.dim {
    opacity: 0.5;
  }

  .label {
    font-size: 15px;
    line-height: 20px;
    color: var(--text-soft);
  }

  .note {
    margin-top: 1px;
    font-size: 12px;
    line-height: 16px;
    color: var(--muted);
  }

  .note.error {
    color: var(--red-text);
  }

  .control {
    flex: none;
    display: flex;
    align-items: center;
  }
</style>
