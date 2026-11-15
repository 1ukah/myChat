<script lang="ts">
  // Steam sign-in, laid out like the Steam client's own: account name and
  // password on the left, "or sign in with QR" on the right. If Steam Guard
  // is needed, the form is replaced by the code step.
  //
  // The password lives only in the field's own state until it is handed to
  // Go, and is cleared the moment it is submitted. It is never stored,
  // logged or shown.
  import { untrack } from "svelte";
  import { app, loginWithPassword, startLogin } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import QrCode from "../ui/QrCode.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import GuardStep from "./GuardStep.svelte";

  const login = $derived(app.login);
  const busy = $derived(login.phase === "submitting");
  const qr = $derived(login.phase === "waiting" ? (login.qr ?? []) : []);

  let account = $state(app.login.account ?? "");
  let password = $state("");
  /** Why the last password sign-in failed. Kept here: restarting the QR replaces LoginState. */
  let formError = $state("");
  let qrError = $state("");

  // The QR code runs whenever this screen is open: start it when idle, and
  // again after a failed password sign-in (which cancelled it). A QR failure
  // is not retried automatically, so a dead connection cannot loop. Nothing
  // starts until Go has said that a login is needed (not while it is still
  // loading a saved session at startup).
  $effect(() => {
    const l = login;
    const needed = app.status.state === "needs_login";
    untrack(() => {
      switch (l.phase) {
        case "idle":
          if (needed) startLogin();
          break;
        case "waiting":
          qrError = "";
          break;
        case "submitting":
          formError = "";
          break;
        case "error":
          if (l.method === "qr") {
            qrError = l.error || "Could not get a QR code from Steam.";
          } else {
            formError = l.error || "Sign-in failed.";
            if (l.account) account = l.account;
            if (needed) startLogin();
          }
          break;
      }
    });
  });

  function submit(e: SubmitEvent): void {
    e.preventDefault();
    if (busy || !account.trim() || !password) return;
    const secret = password;
    password = "";
    formError = "";
    loginWithPassword(account.trim(), secret);
  }
</script>

<div class="signin">
  {#if login.phase === "guard"}
    <GuardStep />
  {:else}
    <div class="columns">
      <form class="form" onsubmit={submit}>
        <label class="field">
          <span class="cap accent">Sign in with account name</span>
          <input
            type="text"
            name="username"
            bind:value={account}
            autocomplete="username"
            autocapitalize="off"
            spellcheck="false"
            disabled={busy}
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
          />
        </label>
        <div class="submit">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            wide
            disabled={busy || !account.trim() || !password}
          >
            {busy ? "Signing in…" : "Sign in"}
          </Button>
        </div>
        <div class="status" aria-live="polite">
          {#if busy}
            <Spinner size={14} label="Signing in" /><span>Contacting Steam…</span>
          {:else if formError}
            <span class="error" role="alert">{formError}</span>
          {/if}
        </div>
      </form>

      <div class="qr">
        <div class="cap accent">Or sign in with QR</div>
        <div class="code" class:dim={busy}>
          {#if qrError}
            <span class="wait bad">{qrError}</span>
          {:else if qr.length > 0}
            <QrCode rows={qr} size={200} label="Steam sign-in QR code" />
          {:else}
            <span class="wait">{busy ? "Paused" : "Getting a code…"}</span>
          {/if}
        </div>
        <p class="how">Steam mobile app → Steam Guard → scan</p>
        <button type="button" class="link" disabled={busy} onclick={startLogin}>
          <Icon name="retry" /><span>{qrError ? "Try again" : "New code"}</span>
        </button>
      </div>
    </div>
  {/if}

  <p class="trust">
    <Icon name="lock" />
    <span>
      Your password goes only to Steam, encrypted with Steam's own key. myChat never stores it.
    </span>
  </p>
</div>

<style>
  .signin {
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  .columns {
    display: flex;
  }

  /* ---------- left: account name and password ---------- */

  .form {
    flex: 1;
    min-width: 0;
    padding: 24px 28px 20px 16px;
  }

  .field {
    display: block;
  }

  .field + .field {
    margin-top: 16px;
  }

  .cap {
    display: block;
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 600;
    line-height: 20px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .cap.accent {
    color: var(--accent);
  }

  input {
    display: block;
    width: 100%;
    height: 40px;
    margin-top: 6px;
    padding: 0 12px;
    font-family: var(--font-body);
    font-size: 15px;
    color: #fff;
    background: rgba(0, 0, 0, 0.42);
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
    background: rgba(0, 0, 0, 0.58);
    border-color: var(--accent);
  }

  input:disabled {
    opacity: 0.5;
  }

  .submit {
    margin-top: 20px;
  }

  /* Room for one line, so progress and errors never move the form. */
  .status {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 20px;
    margin-top: 12px;
    font-size: 13px;
    line-height: 20px;
    color: var(--text-soft);
  }

  .error {
    color: var(--red-text);
  }

  /* ---------- right: QR ---------- */

  .qr {
    flex: none;
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 264px;
    padding: 24px 16px 20px;
    border-left: 1px solid var(--line);
  }

  /* Always white and always the same size, so nothing moves when the code
     arrives or Steam refreshes it. */
  .code {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 208px;
    height: 208px;
    margin-top: 10px;
    padding: 12px;
    text-align: center;
    background: #fff;
    border-radius: 2px;
  }

  .code.dim {
    opacity: 0.35;
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

  .how {
    margin: 10px 0 0;
    font-size: 12px;
    line-height: 20px;
    color: var(--muted);
  }

  .link {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 20px;
    margin-top: 4px;
    padding: 0 4px;
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--muted);
    background: none;
    border: 0;
    cursor: pointer;
  }

  .link:hover {
    color: var(--accent);
  }

  .link:disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .link:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  /* ---------- trust line ---------- */

  .trust {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
    padding: 12px 16px;
    font-size: 12px;
    line-height: 18px;
    color: var(--muted);
    border-top: 1px solid var(--line);
    box-shadow: inset 3px 0 0 var(--green);
  }
</style>
