<script lang="ts">
  // The Steam accounts saved on this PC, as spectator-style player cards:
  // ringed circle avatar, persona name in condensed caps, account name muted,
  // gold highlight on the selected card. One primary button uses the selected
  // account. Accounts the Steam app is not signed in to are muted and offer
  // "Sign in in Steam" instead.
  //
  // Used on the login screen and for "Switch account" when logged in.
  import { app, openSteamLogin, useSteamAccount } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
  import Spinner from "../ui/Spinner.svelte";

  let {
    current = "",
    ondone,
    oncancel,
  }: {
    /** The account myChat is using now (switch mode): tagged, and not offered again. */
    current?: string;
    /** Called after a successful login. */
    ondone?: () => void;
    /** Switch mode: show a Cancel button. */
    oncancel?: () => void;
  } = $props();

  const same = (a: string, b: string): boolean => a.trim().toLowerCase() === b.trim().toLowerCase();

  const accounts = $derived(app.local.accounts ?? []);
  const usable = $derived(accounts.filter((a) => a.hasLogin));

  // The most recent usable account is preselected, until the user picks one.
  let picked = $state<string | null>(null);
  const selected = $derived.by(() => {
    if (picked !== null && usable.some((a) => a.accountName === picked)) return picked;
    const others = usable.filter((a) => !same(a.accountName, current));
    const pool = others.length > 0 ? others : usable;
    return (pool.find((a) => a.mostRecent) ?? pool[0])?.accountName ?? "";
  });

  /** Account being logged in right now. */
  let pending = $state("");
  /** A failure from this screen's own calls, shown on that account's card. */
  let failed = $state<{ account: string; text: string } | null>(null);

  // Go also reports a failed Steam login in LoginState (for example when the
  // sign-in watch found a new login but could not use it).
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
    {@const name = a.personaName || a.accountName}
    {@const error = errorFor(a.accountName)}
    {@const inUse = current !== "" && same(a.accountName, current)}
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
        <Avatar size={36} ring={a.accountName === selected ? "gold" : "offline"} {name} />
        <span class="who">
          <bdi class="persona">{name}</bdi>
          <span class="account">{a.accountName}</span>
          {#if error}<span class="error" role="alert">{error}</span>{/if}
        </span>
        {#if inUse}
          <span class="tag used">In use</span>
        {:else if a.mostRecent}
          <span class="tag">Most recent</span>
        {/if}
        <span class="radio"></span>
      </button>
    {:else}
      <div class="card">
        <span class="muted"><Avatar size={36} ring="offline" {name} /></span>
        <span class="who">
          <bdi class="persona muted">{name}</bdi>
          <span class="account">{a.accountName} · Not signed in to Steam</span>
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
      <!-- Switch mode has no "Steam app" section to show the watch in. -->
      <Spinner size={14} label="Waiting" />
      <span>Waiting for you to sign in to Steam…</span>
    {:else}
      <span>Uses the login the Steam app saved on this PC. It never leaves this PC.</span>
    {/if}
  </p>
</div>

<style>
  .card {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 100%;
    min-height: 56px;
    padding: 8px 16px;
    font: inherit;
    text-align: left;
    color: var(--text);
    background: var(--row);
    border: 0;
    border-bottom: 1px solid var(--line);
  }

  button.card {
    cursor: pointer;
  }

  button.card:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  /* The scoreboard's highlighted player. */
  button.card.selected {
    background: var(--gold-wash);
    box-shadow:
      inset 3px 0 0 var(--gold),
      inset 0 1px 0 rgba(224, 178, 82, 0.35),
      inset 0 -1px 0 rgba(224, 178, 82, 0.35);
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
    font-family: var(--font-head);
    font-size: 20px;
    font-weight: 700;
    line-height: 22px;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: #fff;
  }

  .account {
    overflow: hidden;
    font-size: 13px;
    line-height: 18px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .selected .account {
    color: #d9c08a;
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
    padding: 0 8px;
    font-family: var(--font-head);
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--text-soft);
    border: 1px solid var(--line-strong);
  }

  .selected .tag {
    color: #f3dca6;
    border-color: rgba(224, 178, 82, 0.6);
  }

  .tag.used {
    color: var(--green-text);
    border-color: var(--green);
  }

  .radio {
    flex: none;
    width: 16px;
    height: 16px;
    border: 1px solid var(--line-strong);
    border-radius: 50%;
  }

  .selected .radio {
    background: radial-gradient(circle, var(--gold) 0 4px, transparent 4.5px);
    border-color: var(--gold);
  }

  .foot {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
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
