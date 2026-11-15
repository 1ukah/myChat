<script lang="ts">
  // Steam Guard step of the password login (phase "guard"): a 5-character
  // code from the authenticator app or from email, and/or approval in the
  // Steam mobile app (which completes on its own).
  import { onMount, untrack } from "svelte";
  import { app, cancelLogin, submitGuardCode } from "../store.svelte";
  import Icon from "../ui/Icon.svelte";
  import Spinner from "../ui/Spinner.svelte";

  const LENGTH = 5;

  const login = $derived(app.login);
  const guards = $derived(login.guards ?? []);
  const byApp = $derived(guards.includes("device_code"));
  const byEmail = $derived(guards.includes("email_code"));
  const wantsCode = $derived(byApp || byEmail);
  const canApprove = $derived(guards.includes("device_confirmation"));
  const canUseLink = $derived(guards.includes("email_confirmation"));

  let code = $state("");
  let focused = $state(false);
  let input = $state<HTMLInputElement>();

  const boxes = Array.from({ length: LENGTH }, (_, i) => i);
  const cursor = $derived(Math.min(code.length, LENGTH - 1));

  function oninput(e: Event & { currentTarget: HTMLInputElement }): void {
    // Typing and pasting both land here: keep letters and digits only.
    const next = e.currentTarget.value
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, "")
      .slice(0, LENGTH);
    e.currentTarget.value = next;
    code = next;
    if (next.length === LENGTH && !login.checking) submitGuardCode(next);
  }

  // A wrong code: empty the boxes and put the cursor back for another try.
  $effect(() => {
    if (!login.codeError) return;
    untrack(() => (code = ""));
    input?.focus();
  });

  onMount(() => input?.focus());
</script>

<div class="guard">
  <div class="cap">Steam Guard</div>

  {#if wantsCode}
    <h3>
      {#if byApp}
        Enter the code from your Steam Guard authenticator app
      {:else}
        Enter the code Steam sent to your {login.emailHint ? `${login.emailHint} address` : "email address"}
      {/if}
    </h3>
  {:else}
    <h3>Approve the sign-in to continue</h3>
  {/if}
  {#if login.account}
    <p class="who">Signing in as <b>{login.account}</b></p>
  {/if}

  {#if wantsCode}
    <div class="code" class:focused class:bad={!!login.codeError} class:busy={login.checking}>
      <input
        bind:this={input}
        type="text"
        value={code}
        aria-label="Steam Guard code, {LENGTH} characters"
        aria-invalid={!!login.codeError}
        autocomplete="one-time-code"
        autocapitalize="characters"
        spellcheck="false"
        disabled={login.checking}
        {oninput}
        onfocus={() => (focused = true)}
        onblur={() => (focused = false)}
      />
      <div class="boxes" aria-hidden="true">
        {#each boxes as i (i)}
          <span class="box" class:at={focused && i === cursor && code.length < LENGTH}>
            {code[i] ?? ""}
          </span>
        {/each}
      </div>
    </div>

    <div class="status" aria-live="polite">
      {#if login.checking}
        <Spinner size={14} label="Checking the code" /><span>Checking the code…</span>
      {:else if login.codeError}
        <span class="error">{login.codeError}</span>
      {/if}
    </div>
  {:else}
    <div class="status waiting">
      <Spinner size={14} label="Waiting" /><span>Waiting for your approval…</span>
    </div>
  {/if}

  {#if canApprove}
    <p class="other">
      <Icon name="phone" />
      <span>{wantsCode ? "Or approve" : "Approve"} the sign-in in the Steam mobile app</span>
    </p>
  {/if}
  {#if canUseLink}
    <p class="other">
      <Icon name="mail" />
      <span>{wantsCode || canApprove ? "Or use" : "Use"} the link in the email Steam sent you</span>
    </p>
  {/if}

  <button type="button" class="back" onclick={cancelLogin}>
    <Icon name="back" />
    <span>Back</span>
  </button>
</div>

<style>
  .guard {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 0 20px;
    text-align: center;
    border-bottom: 1px solid var(--line);
  }

  .cap {
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.01em;
    text-transform: uppercase;
    color: var(--accent);
  }

  h3 {
    margin: 8px 0 0;
    font-size: 17px;
    font-weight: 600;
    line-height: 24px;
    color: #fff;
  }

  .who {
    margin: 4px 0 0;
    font-size: 13px;
    line-height: 20px;
    color: var(--muted);
  }

  .who b {
    font-weight: 600;
    color: var(--text-soft);
  }

  /* One real input, invisible, on top of five drawn boxes: typing, paste
     and the on-screen keyboard all behave as in any text field. */
  .code {
    position: relative;
    margin-top: 20px;
  }

  .code input {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    margin: 0;
    padding: 0;
    opacity: 0;
    border: 0;
    cursor: text;
  }

  .boxes {
    display: flex;
    gap: 8px;
  }

  .box {
    width: 48px;
    height: 60px;
    font-family: var(--font-cond);
    font-size: 38px;
    font-weight: 700;
    line-height: 58px;
    text-align: center;
    color: #fff;
    background: var(--box);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }

  .focused .box {
    border-color: rgba(255, 255, 255, 0.5);
  }

  /* The box the next character goes into. */
  .box.at {
    border-color: var(--accent);
    box-shadow: inset 0 -3px 0 var(--accent);
  }

  .bad .box {
    border-color: var(--red);
  }

  .busy .box {
    opacity: 0.5;
  }

  .status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 20px;
    margin-top: 12px;
    font-size: 13px;
    line-height: 20px;
    color: var(--muted);
  }

  .status.waiting {
    margin-top: 20px;
    color: var(--text-soft);
  }

  .error {
    color: var(--red-text);
  }

  .other {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 12px 0 0;
    font-size: 13px;
    line-height: 20px;
    color: var(--text-soft);
  }

  .back {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 28px;
    margin-top: 16px;
    padding: 0 8px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.01em;
    text-transform: uppercase;
    color: #fff;
    background: none;
    border: 0;
    cursor: pointer;
  }

  .back:hover {
    color: var(--accent);
  }

  .back:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(.anim) .box.at {
    animation: at 1.1s step-end infinite;
  }

  @keyframes at {
    50% {
      box-shadow: inset 0 -3px 0 transparent;
    }
  }
</style>
