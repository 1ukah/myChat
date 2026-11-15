<script lang="ts" generics="T extends string">
  // CS2-style dropdown: the value in condensed uppercase with a chevron, and
  // a flat list that opens below. The option under the pointer or the
  // keyboard gets the gold highlight.
  import { tick } from "svelte";
  import Icon from "./Icon.svelte";

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

  let open = $state(false);
  let root = $state<HTMLElement>();
  let trigger = $state<HTMLButtonElement>();

  const current = $derived(options.find((o) => o.value === value));

  const items = (): HTMLElement[] =>
    Array.from(root?.querySelectorAll<HTMLElement>('[role="option"]') ?? []);

  async function show(): Promise<void> {
    open = true;
    await tick();
    (items().find((el) => el.getAttribute("aria-selected") === "true") ?? items()[0])?.focus();
  }

  function hide(): void {
    open = false;
    trigger?.focus();
  }

  function pick(next: T): void {
    hide();
    if (next !== value) onchange(next);
  }

  function onkeydown(e: KeyboardEvent): void {
    if (e.key === "Escape" && open) {
      e.preventDefault();
      hide();
      return;
    }
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    e.preventDefault();
    if (!open) {
      void show();
      return;
    }
    const all = items();
    const at = all.indexOf(document.activeElement as HTMLElement);
    const step = e.key === "ArrowDown" ? 1 : -1;
    all[(at + step + all.length) % all.length]?.focus();
  }

  // Focus left the control (click elsewhere, Tab away): close.
  function onfocusout(e: FocusEvent): void {
    if (!(e.relatedTarget instanceof Node) || !root?.contains(e.relatedTarget)) open = false;
  }
</script>

<div class="select" class:open bind:this={root} {onfocusout}>
  <button
    type="button"
    class="value"
    bind:this={trigger}
    aria-haspopup="listbox"
    aria-expanded={open}
    aria-label="{label}: {current?.label ?? value}"
    {disabled}
    onclick={() => (open ? hide() : show())}
    {onkeydown}
  >
    <span>{current?.label ?? value}</span>
    <span class="chevron"><Icon name="chevron" /></span>
  </button>

  {#if open}
    <div class="list" role="listbox" aria-label={label}>
      {#each options as option (option.value)}
        <button
          type="button"
          role="option"
          class="option"
          class:selected={option.value === value}
          aria-selected={option.value === value}
          onclick={() => pick(option.value)}
          {onkeydown}
        >
          {option.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .select {
    position: relative;
    width: 220px;
    font-family: var(--font-head);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .value {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    height: 32px;
    padding: 0 8px 0 12px;
    font: inherit;
    letter-spacing: inherit;
    text-transform: inherit;
    color: var(--text);
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
    cursor: pointer;
  }

  .value:hover,
  .open .value {
    border-color: rgba(255, 255, 255, 0.6);
  }

  .value:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .chevron {
    color: var(--muted);
  }

  .open .chevron {
    transform: rotate(180deg);
  }

  .list {
    position: absolute;
    top: calc(100% + 2px);
    right: 0;
    left: 0;
    z-index: 20;
    padding: 2px 0;
    background: #171a1e;
    border: 1px solid var(--line-strong);
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.55);
  }

  .option {
    display: block;
    width: 100%;
    height: 30px;
    padding: 0 12px;
    font: inherit;
    font-weight: 600;
    letter-spacing: inherit;
    text-align: left;
    text-transform: inherit;
    color: var(--text-soft);
    background: none;
    border: 0;
    cursor: pointer;
  }

  .option.selected {
    color: var(--accent);
  }

  .option:hover,
  .option:focus-visible {
    color: #fff;
    background: var(--gold-wash);
    box-shadow: inset 3px 0 0 var(--gold);
    outline: none;
  }

  .value:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(.anim) .list {
    animation: list-in 100ms var(--ease-out) both;
  }

  @keyframes list-in {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
</style>
