<script lang="ts">
  // The login screen. myChat never asks for a password. Three ways in:
  //   1. Use Steam on this PC: the login the Steam app already saved (default).
  //   2. Sign in with the Steam app: open Steam's own sign-in; Go watches for
  //      the new login and continues on its own.
  //   3. QR code: myChat's own session. Collapsed until asked for.
  // If Steam is not installed, only the QR is shown.
  import { onMount, untrack } from "svelte";
  import {
    app,
    cancelLogin,
    openSteamLogin,
    refreshLocalLogins,
    startLogin,
  } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import QrCode from "../ui/QrCode.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import AccountCards from "./AccountCards.svelte";

  const login = $derived(app.login);
  const local = $derived(app.local);
  const hasSteam = $derived(local.steamInstalled);
  const accounts = $derived(local.accounts ?? []);
  const anyUsable = $derived(accounts.some((a) => a.hasLogin));

  // ---- banner: why the login screen is showing (for example an expired login) ----

  // Remembered here, because opening the QR replaces LoginState.
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

  // ---- 2. Steam app sign-in ----

  let opening = $state(false);
  let openError = $state("");

  async function openSteam(): Promise<void> {
    opening = true;
    openError = "";
    openError = await openSteamLogin();
    opening = false;
  }

  // ---- 3. QR ----

  // The panel follows the state: it is open exactly while a QR login is
  // running or has failed, so it can never disagree with Go.
  const qrActive = $derived(
    login.method === "qr" && (login.phase === "waiting" || login.phase === "error"),
  );
  const qrOpen = $derived(!hasSteam || qrActive);
  const qr = $derived(qrActive && login.phase === "waiting" ? (login.qr ?? []) : []);
  const qrError = $derived(
    qrActive && login.phase === "error"
      ? login.error || "Could not get a QR code from Steam."
      : "",
  );

  function toggleQr(): void {
    if (qrActive) cancelLogin();
    else startLogin();
  }

  // Without Steam the QR is the only way, so it starts by itself. A failed QR
  // is not retried automatically: a dead connection cannot loop.
  $effect(() => {
    if (!hasSteam && login.phase === "idle" && app.status.state === "needs_login") {
      untrack(startLogin);
    }
  });

  // Steam's saved accounts may have changed since the app started.
  onMount(refreshLocalLogins);
</script>

<div class="login">
  {#if banner}
    <div class="banner" role="alert">
      <Icon name="alert" />
      <span>{banner}</span>
    </div>
  {/if}

  {#if hasSteam}
    <section class="way">
      <h3 class="main"><Icon name="pc" /><span>Use Steam on this PC</span></h3>
      {#if accounts.length > 0}
        <AccountCards />
      {:else}
        <p class="text empty">
          The Steam app has no saved account on this PC yet. Sign in with the Steam app below.
        </p>
      {/if}
      {#if local.error}
        <p class="text bad">{local.error}</p>
      {/if}
    </section>

    <section class="way">
      <h3><Icon name="open" /><span>Sign in with the Steam app</span></h3>
      <div class="line">
        {#if local.watching}
          <Spinner size={16} label="Waiting" />
          <p class="text" aria-live="polite">Waiting for you to sign in to Steam…</p>
          <Button onclick={cancelLogin}>Cancel</Button>
        {:else}
          <p class="text">
            Sign in to Steam with "Remember me" on. myChat continues on its own.
          </p>
          <Button
            variant={anyUsable ? "outline" : "primary"}
            icon="open"
            disabled={opening}
            onclick={openSteam}
          >
            Open Steam sign-in
          </Button>
        {/if}
      </div>
      {#if openError}
        <p class="text bad" role="alert">{openError}</p>
      {/if}
    </section>
  {:else}
    <p class="text nosteam">
      {local.error || "Steam is not installed on this PC."} You can still sign in with a QR code.
    </p>
  {/if}

  <section class="way">
    {#if hasSteam}
      <button type="button" class="fold" aria-expanded={qrOpen} onclick={toggleQr}>
        <Icon name="qr" />
        <span>Sign in with QR code</span>
        <span class="chevron" class:open={qrOpen}><Icon name="chevron" /></span>
      </button>
    {:else}
      <h3 class="main"><Icon name="qr" /><span>Sign in with QR code</span></h3>
    {/if}

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
          <p class="fine">myChat keeps its own session for this. Nothing is typed in here.</p>
          <button type="button" class="link" onclick={startLogin}>
            <Icon name="retry" /><span>{qrError ? "Try again" : "New code"}</span>
          </button>
        </div>
      </div>
    {/if}
  </section>
</div>

<style>
  /* ---------- banner ---------- */

  .banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    font-size: 13px;
    line-height: 20px;
    color: #ffd9a8;
    background: rgba(224, 178, 82, 0.12);
    border-bottom: 1px solid var(--line);
    box-shadow: inset 3px 0 0 var(--gold);
  }

  /* ---------- the three ways ---------- */

  .way {
    border-bottom: 1px solid var(--line);
  }

  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 36px;
    margin: 0;
    padding: 0 16px;
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--text-soft);
  }

  h3.main {
    color: var(--accent);
  }

  .text {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: 13px;
    line-height: 20px;
    color: var(--muted);
  }

  .text.empty,
  .text.nosteam {
    padding: 4px 16px 14px;
  }

  .text.nosteam {
    padding-top: 14px;
    border-bottom: 1px solid var(--line);
  }

  .text.bad {
    padding: 0 16px 10px;
    color: var(--red-text);
  }

  .line {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 0 16px 10px;
  }

  /* ---------- QR ---------- */

  .fold {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    height: 40px;
    padding: 0 16px;
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-align: left;
    text-transform: uppercase;
    color: var(--text-soft);
    background: none;
    border: 0;
    cursor: pointer;
  }

  .fold:hover {
    color: var(--accent);
    background: rgba(255, 255, 255, 0.04);
  }

  .fold:focus-visible,
  .link:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .chevron {
    margin-left: auto;
    color: var(--muted);
  }

  .chevron.open {
    transform: rotate(180deg);
  }

  .qr {
    display: flex;
    align-items: center;
    gap: 28px;
    padding: 8px 16px 20px;
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
    border-radius: 2px;
  }

  .wait {
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #6b7279;
  }

  .wait.bad {
    font-family: var(--font-body);
    font-size: 13px;
    font-weight: 400;
    line-height: 18px;
    letter-spacing: 0;
    text-transform: none;
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
    font-family: var(--font-head);
    font-size: 16px;
    font-weight: 700;
    line-height: 22px;
    text-align: center;
    color: var(--accent);
    border: 1px solid var(--accent-line);
  }

  .fine {
    margin: 8px 0 0;
    font-size: 12px;
    line-height: 18px;
    color: var(--muted);
  }

  .link {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 24px;
    margin-top: 8px;
    padding: 0 4px;
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--text-soft);
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
