<script lang="ts">
  // A small progress ring. It turns only when animations are on.
  let { size = 16, label = "Working" }: { size?: number; label?: string } = $props();
</script>

<span class="spinner" style:--size="{size}px" role="img" aria-label={label}></span>

<style>
  .spinner {
    flex: none;
    display: inline-block;
    width: var(--size);
    height: var(--size);
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-top-color: var(--accent);
    border-radius: 50%;
  }

  :global(.anim) .spinner {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
