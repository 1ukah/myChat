<script lang="ts">
  // The login screen. Three ways in; none of them changes the Steam app's
  // status:
  //   1. Use Steam on this PC: the login the Steam app already saved.
  //   2. Account name + password, then Steam Guard if Steam asks for it.
  //   3. QR code scanned with the Steam mobile app. Collapsed until asked for.
  // If Steam is not installed, 1 is replaced by a short note.
  //
  // The password lives only in the field's own state until it is handed to
  // Go, and is cleared the moment it is submitted. It is never stored,
  // logged or shown.
  import { onMount } from "svelte";
  import {
    app,
    cancelLogin,
    loginWithPassword,
    openSteamLogin,
    refreshLocalLogins,
    startLogin,
  } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import QrCode from "../ui/QrCode.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import AccountCards from "./AccountCards.svelte";
  import GuardStep from "./GuardStep.svelte";

  const login = $derived(app.login);
  const local = $derived(app.local);
  const accounts = $derived(local.accounts ?? []);

  // ---- banner: why the login screen is showing (for example an expired login) ----

  // Remembered here, because starting a login replaces LoginState.
  let reason = $state("");
  $effect(() => {
    if (login.phase === "idle" && login.error) reason = login.error;
  });

  // A failed Steam login for an account that has no card to show it on.
  const stray = $derived(
    login.phase === "error" &&
      login.method === "steam" &&
      !accounts.some((a) => a.accountName.toLowerCase() === login.account.trim().toLowerCase())
      ? login.error
      : "",
  );
  const banner = $derived(stray || reason);

  // ---- 1. Steam on this PC ----

  let opening = $state(false);
  let openError = $state("");

  async function openSteam(): Promise<void> {
    opening = true;
    openError = "";
    openError = await openSteamLogin();
    opening = false;
  }

  // Steam's saved accounts may have changed since the app started.
  onMount(refreshLocalLogins);

  // ---- 2. account name + password ----

  let account = $state(app.login.account ?? "");
  let password = $state("");

  const busy = $derived(login.method === "password" && login.phase === "submitting");
  const formError = $derived(
    login.method === "password" && login.phase === "error" ? login.error || "Sign-in failed." : "",
  );
  const canSubmit = $derived(!busy && account.trim() !== "" && password !== "");

  function submit(): void {
    if (!canSubmit) return;
    const secret = password;
    password = "";
    loginWithPassword(account.trim(), secret);
  }

  function onkeydown(e: KeyboardEvent): void {
    if (e.key === "Enter") {
      e.preventDefault();
      submit();
    }
  }

  // ---- 3. QR ----

  // The panel follows the state: it is open exactly while a QR login is
  // running or has failed, so it can never disagree with Go.
  const qrOpen = $derived(
    login.method === "qr" && (login.phase === "waiting" || login.phase === "error"),
  );
  const qr = $derived(qrOpen && login.phase === "waiting" ? (login.qr ?? []) : []);
  const qrError = $derived(
    qrOpen && login.phase === "error" ? login.error || "Could not get a QR code from Steam." : "",
  );

  function toggleQr(): void {
    if (qrOpen) cancelLogin();
    else startLogin();
  }
</script>

<div class="login">
  {#if banner}
    <div class="banner" role="alert">
      <Icon name="alert" />
      <span>{banner}</span>
    </div>
  {/if}

  {#if login.phase === "guard"}
    <GuardStep />
  {:else}
    {#if local.steamInstalled}
      <section class="way">
        <h3 class="main"><Icon name="pc" /><span>Use Steam on this PC</span></h3>
        {#if accounts.length > 0}
          <AccountCards />
        {/if}
        {#if local.watching}
          <div class="line">
            <Spinner size={16} label="Waiting" />
            <p class="text" aria-live="polite">Waiting for you to sign in to Steam…</p>
            <Button onclick={cancelLogin}>Cancel</Button>
          </div>
        {:else if accounts.length === 0}
          <div class="line">
            <p class="text">
              No saved Steam account on this PC. Sign in to Steam with "Remember me" on.
            </p>
            <Button icon="open" disabled={opening} onclick={openSteam}>Open Steam sign-in</Button>
          </div>
        {/if}
        {#if openError || local.error}
          <p class="text bad" role="alert">{openError || local.error}</p>
        {/if}
      </section>
    {:else}
      <p class="text nosteam">{local.error || "Steam is not installed on this PC."}</p>
    {/if}

    <section class="way">
      <h3 class:main={!local.steamInstalled}>
        <Icon name="user" /><span>Sign in with account name</span>
      </h3>
      <div class="form">
        <label class="field">
          <span class="cap">Account name</span>
          <input
            type="text"
            name="username"
            bind:value={account}
            autocomplete="username"
            autocapitalize="off"
            spellcheck="false"
            disabled={busy}
            {onkeydown}
          />
        </label>
        <label class="field">
          <span class="cap">Password</span>
          <input
            type="password"
            name="password"
            bind:value={password}
            autocomplete="current-password"
            disabled={busy}
            {onkeydown}
          />
        </label>
        <div class="submit">
          <Button variant="primary" disabled={!canSubmit} onclick={submit}>
            {busy ? "Signing in…" : "Sign in"}
          </Button>
          <div class="status" aria-live="polite">
            {#if busy}
              <Spinner size={14} label="Signing in" /><span>Contacting Steam…</span>
            {:else if formError}
              <span class="error" role="alert">{formError}</span>
            {/if}
          </div>
        </div>
      </div>
      <p class="trust">
        <Icon name="lock" />
        <span>Your password goes only to Steam and is not stored.</span>
      </p>
    </section>

    <section class="way">
      <button type="button" class="fold" aria-expanded={qrOpen} onclick={toggleQr}>
        <Icon name="qr" />
        <span>Sign in with QR code</span>
        <span class="chevron" class:open={qrOpen}><Icon name="chevron" /></span>
      </button>

      {#if qrOpen}
        <div class="qr">
          <!-- Always white and the same size, so nothing moves when the code
               arrives or Steam refreshes it. -->
          <div class="code">
            {#if qrError}
              <span class="wait bad">{qrError}</span>
            {:else if qr.length > 0}
              <QrCode rows={qr} size={184} label="Steam sign-in QR code" />
            {:else}
              <span class="wait">Getting a code…</span>
            {/if}
          </div>
          <div class="guide">
            <ol>
              <li><span class="n">1</span><span>Open the <b>Steam</b> app on your phone.</span></li>
              <li><span class="n">2</span><span>Tap the <b>Steam Guard</b> shield icon.</span></li>
              <li><span class="n">3</span><span>Scan this code and approve the sign-in.</span></li>
            </ol>
            <button type="button" class="link" onclick={startLogin}>
              <Icon name="retry" /><span>{qrError ? "Try again" : "New code"}</span>
            </button>
          </div>
        </div>
      {/if}
    </section>
  {/if}
</div>

<style>
  /* ---------- banner ---------- */

  .banner {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 6px;
    padding: 10px 14px;
    font-size: 13px;
    line-height: 20px;
    color: #ffe2b0;
    background: rgba(240, 164, 40, 0.16);
    box-shadow: inset 3px 0 0 var(--gold);
  }

  /* ---------- the three ways ---------- */

  .way {
    border-bottom: 1px solid var(--line);
  }

  .text {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: 13px;
    line-height: 20px;
    color: var(--muted);
  }

  .text.nosteam {
    padding: 14px 0;
    border-bottom: 1px solid var(--line);
  }

  .text.bad {
    padding: 0 0 10px;
    color: var(--red-text);
  }

  .line {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 0 0 12px;
  }

  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 44px;
    margin: 0;
    padding: 6px 0 0;
    font-size: var(--fs-md);
    font-weight: 600;
    color: #fff;
  }

  h3.main {
    color: var(--accent);
  }

  /* ---------- account name + password ---------- */

  .form {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px 16px;
    padding: 4px 0 4px;
  }

  .field {
    display: block;
    min-width: 0;
  }

  .cap {
    display: block;
    font-size: var(--fs-xs);
    font-weight: 500;
    line-height: 20px;
    color: var(--muted);
  }

  input {
    display: block;
    width: 100%;
    height: 36px;
    margin-top: 4px;
    padding: 0 12px;
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    color: #fff;
    background: var(--box);
    border: 1px solid var(--line);
    border-radius: 2px;
    outline: none;
    user-select: text;
    cursor: text;
  }

  input:hover {
    border-color: var(--line-strong);
  }

  input:focus {
    background: rgba(0, 0, 0, 0.4);
    border-color: var(--accent);
  }

  input:disabled {
    opacity: 0.5;
  }

  .submit {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    gap: 14px;
  }

  /* Room for one line, so progress and errors never move the form. */
  .status {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    min-height: 20px;
    font-size: 13px;
    line-height: 20px;
    color: var(--text-soft);
  }

  .error {
    color: var(--red-text);
  }

  .trust {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 12px 0 14px;
    font-size: 12px;
    line-height: 18px;
    color: var(--muted);
  }

  /* ---------- QR ---------- */

  .fold {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    height: 50px;
    padding: 0;
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    font-weight: 600;
    text-align: left;
    color: #fff;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .fold:hover {
    color: var(--accent);
  }

  .fold:focus-visible,
  .link:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .chevron {
    margin-left: auto;
  }

  .chevron.open {
    transform: rotate(180deg);
  }

  .qr {
    display: flex;
    align-items: center;
    gap: 28px;
    padding: 4px 0 20px;
  }

  .code {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 192px;
    height: 192px;
    padding: 10px;
    text-align: center;
    background: #fff;
  }

  .wait {
    font-size: var(--fs-sm);
    font-weight: 500;
    line-height: 18px;
    color: #6b7279;
  }

  .wait.bad {
    font-weight: 400;
    color: #a3221d;
  }

  .guide {
    flex: 1;
    min-width: 0;
  }

  ol {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 5px 0;
    font-size: 14px;
    line-height: 24px;
    color: var(--text-soft);
  }

  li b {
    font-weight: 600;
    color: #fff;
  }

  .n {
    flex: none;
    width: 24px;
    height: 24px;
    font-size: var(--fs-sm);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 24px;
    text-align: center;
    color: var(--accent);
    background: var(--tab-fill);
    border-radius: 2px;
  }

  .link {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 24px;
    margin-top: 8px;
    padding: 0;
    font-family: var(--font-ui);
    font-size: var(--fs-sm);
    font-weight: 600;
    color: #fff;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .link:hover {
    color: var(--accent);
  }

  :global(.anim) .qr {
    animation: unfold 160ms var(--ease-out) both;
  }

  @keyframes unfold {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
</style>
