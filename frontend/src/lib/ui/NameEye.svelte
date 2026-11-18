<script lang="ts">
  import Icon from "./Icon.svelte";

  let {
    shown,
    onclick,
  }: {
    shown: boolean;
    onclick: () => void;
  } = $props();

  const label = $derived(shown ? "Hide account name" : "Show account name");
</script>

<button type="button" class="eye" aria-label={label} title={label} aria-pressed={shown} {onclick}>
  <Icon name={shown ? "eye-off" : "eye"} size={14} />
</button>

<style>
  .eye {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 20px;
    padding: 0;
    color: var(--muted);
    background: none;
    border: 0;
    border-radius: 2px;
    cursor: pointer;
  }

  .eye:hover {
    color: var(--accent);
  }

  .eye:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  :global(.anim) .eye {
    transition: color 120ms;
  }
</style>
