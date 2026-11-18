<script lang="ts">
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
    error?: string;
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
    border-bottom: 1px solid var(--line);
  }

  .info {
    min-width: 0;
  }

  .info.dim {
    opacity: 0.5;
  }

  .label {
    font-size: var(--fs-md);
    font-weight: 500;
    line-height: 20px;
    color: var(--text);
  }

  .note {
    margin-top: 2px;
    font-size: var(--fs-xs);
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
