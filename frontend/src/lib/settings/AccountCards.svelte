<script lang="ts">
  import { app, openSteamLogin, useSteamAccount } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import NameMask from "../ui/NameMask.svelte";
  import Spinner from "../ui/Spinner.svelte";

  let {
    current = "",
    ondone,
    oncancel,
  }: {
    current?: string;
    ondone?: () => void;
    oncancel?: () => void;
  } = $props();

  const same = (a: string, b: string): boolean => a.trim().toLowerCase() === b.trim().toLowerCase();

  const accounts = $derived(app.local.accounts ?? []);
  const usable = $derived(accounts.filter((a) => a.hasLogin));

  let picked = $state<string | null>(null);
  const selected = $derived.by(() => {
    if (picked !== null && usable.some((a) => a.accountName === picked)) return picked;
    const others = usable.filter((a) => !same(a.accountName, current));
    const pool = others.length > 0 ? others : usable;
    return (pool.find((a) => a.mostRecent) ?? pool[0])?.accountName ?? "";
  });

  let pending = $state("");
  let failed = $state<{ account: string; text: string } | null>(null);

  function errorFor(account: string): string {
    if (failed && same(failed.account, account)) return failed.text;
    const l = app.login;
    return l.phase === "error" && l.method === "steam" && same(l.account, account) ? l.error : "";
  }

  async function use(): Promise<void> {
    const account = selected;
    if (!account || pending) return;
    pending = account;
    failed = null;
    const text = await useSteamAccount(account);
    pending = "";
    if (text) failed = { account, text };
    else ondone?.();
  }

  async function signInInSteam(account: string): Promise<void> {
    failed = null;
    const text = await openSteamLogin();
    if (text) failed = { account, text };
  }
</script>

<div class="cards" role="radiogroup" aria-label="Steam accounts on this PC">
  {#each accounts as a (a.accountName)}
    {@const inUse = current !== "" && same(a.accountName, current)}
    {@const veiled = !app.config.showAccountName}
    {@const name = a.personaName || (veiled ? "Steam account" : a.accountName)}
    {@const error = errorFor(a.accountName)}
    {#if a.hasLogin}
      <button
        type="button"
        role="radio"
        class="card"
        class:selected={a.accountName === selected}
        aria-checked={a.accountName === selected}
        disabled={pending !== ""}
        onclick={() => (picked = a.accountName)}
      >
        <Avatar
          size={40}
          ring={a.accountName === selected ? "accent" : "none"}
          src={inUse ? app.status.avatar : ""}
          name=""
        />
        <span class="who">
          <bdi class="persona">{name}</bdi>
          <span class="account">{#if veiled}<NameMask />{:else}{a.accountName}{/if}</span>
          {#if error}<span class="error" role="alert">{error}</span>{/if}
        </span>
        {#if inUse}
          <span class="tag used">In use</span>
        {:else if a.mostRecent}
          <span class="tag">Most recent</span>
        {/if}
        <span class="radio"><Icon name="check" size={14} /></span>
      </button>
    {:else}
      <div class="card">
        <span class="muted"><Avatar size={40} name="" /></span>
        <span class="who">
          <bdi class="persona muted">{name}</bdi>
          <span class="account">{#if veiled}<NameMask />{:else}{a.accountName}{/if} · Not signed in to Steam</span>
          {#if error}<span class="error" role="alert">{error}</span>{/if}
        </span>
        <Button disabled={pending !== ""} onclick={() => signInInSteam(a.accountName)}>
          Sign in in Steam
        </Button>
      </div>
    {/if}
  {/each}
</div>

<div class="foot">
  <Button
    variant="primary"
    size="lg"
    disabled={!selected || pending !== "" || (current !== "" && same(selected, current))}
    onclick={use}
  >
    {pending ? "Connecting…" : "Use this account"}
  </Button>
  {#if oncancel}
    <Button size="lg" disabled={pending !== ""} onclick={oncancel}>Cancel</Button>
  {/if}
  <p class="trust">
    {#if pending}
      <Spinner size={14} label="Connecting" />
    {/if}
    {#if oncancel && app.local.watching}
      <Spinner size={14} label="Waiting" />
      <span>Waiting for you to sign in to Steam…</span>
    {/if}
  </p>
</div>

<style>
  .card {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    min-height: 58px;
    margin-bottom: 2px;
    padding: 8px 14px 8px 10px;
    font: inherit;
    text-align: left;
    color: var(--text);
    background: var(--box);
    border: 0;
  }

  button.card {
    cursor: pointer;
  }

  button.card:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  button.card.selected {
    background: var(--row-pick);
    box-shadow: inset 0 0 0 1px var(--row-pick-line);
  }

  button.card:disabled {
    cursor: default;
  }

  button.card:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .muted {
    opacity: 0.5;
  }

  .who {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .persona {
    overflow: hidden;
    font-size: var(--fs-md);
    font-weight: 600;
    line-height: 22px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: #fff;
  }

  .account {
    overflow: hidden;
    font-size: 12px;
    line-height: 17px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .selected .account {
    color: #c5d6e6;
  }

  .error {
    margin-top: 2px;
    font-size: 12px;
    line-height: 16px;
    color: var(--red-text);
  }

  .tag {
    flex: none;
    height: 20px;
    padding: 0 9px;
    font-size: var(--fs-xs);
    font-weight: 600;
    line-height: 20px;
    color: #fff;
    background: #4f7fb8;
    border-radius: 10px;
  }

  .tag.used {
    background: #3f9a45;
  }

  .radio {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    color: transparent;
    border: 1px solid var(--line-strong);
  }

  .selected .radio {
    color: #10202c;
    background: var(--accent);
    border-color: var(--accent);
  }

  .foot {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 0;
  }

  .trust {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    margin: 0 0 0 4px;
    font-size: 12px;
    line-height: 18px;
    color: var(--muted);
  }

  :global(.anim) button.card {
    transition: background-color 120ms;
  }
</style>
