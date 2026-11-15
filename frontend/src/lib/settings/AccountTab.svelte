<script lang="ts">
  import { app, cancelLogin, logout, refreshLocalLogins, saveConfig } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import AccountCards from "./AccountCards.svelte";
  import Login from "./Login.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const status = $derived(app.status);

  // A finished login sets the account a moment before Steam reports a
  // connection state; "done" bridges that gap so the login screen does not
  // flash back.
  const loggedOut = $derived(
    !status.accountName || (status.state === "needs_login" && app.login.phase !== "done"),
  );
  /** The Steam login name is private: shown only when the user asks for it. */
  const showAccount = $derived(app.config.showAccountName);
  const name = $derived(
    status.personaName || (showAccount ? status.accountName : "") || "Steam",
  );
  const viaSteam = $derived(status.source === "steam");
  const method = $derived(
    status.source === "steam"
      ? "Using Steam on this PC"
      : status.source === "password"
        ? "Password login"
        : "QR login",
  );

  const conn = $derived.by(() => {
    switch (status.state) {
      case "online":
        // Logged in to Steam; it says nothing about the user's Steam status.
        return { tone: "in", ring: "accent", text: "Logged in" } as const;
      case "connecting":
      case "needs_login":
        return { tone: "warn", ring: "accent", text: "Connecting" } as const;
      default:
        return { tone: "bad", ring: "offline", text: "Offline" } as const;
    }
  });

  let eyeError = $state("");

  async function toggleAccount(): Promise<void> {
    eyeError = await saveConfig({ showAccountName: !showAccount });
  }

  /** Showing Steam's saved accounts again, to change to another one. */
  let switching = $state(false);

  function startSwitch(): void {
    refreshLocalLogins();
    switching = true;
  }

  function cancelSwitch(): void {
    // "Sign in in Steam" on a card may have started the sign-in watch.
    if (app.local.watching) cancelLogin();
    switching = false;
  }
</script>

{#if loggedOut}
  <Section id="steam" title="Sign in to Steam">
    <Login />
  </Section>
{:else}
  <Section id="steam" title="Steam">
    <!-- Status card, in the style of the CS2 spectator bar: the player in
         focus gets the circle avatar with a coloured ring. -->
    <div class="card">
      <Avatar round size={64} ring={conn.ring} src={status.avatar} {name} />
      <div class="who">
        <div class="name"><bdi>{name}</bdi></div>
        <div class="state">
          <span class="dot {conn.tone}"></span>
          <span class="label {conn.tone}">{conn.text}</span>
          {#if showAccount}
            <span class="account">· {status.accountName}</span>
          {/if}
          <button
            type="button"
            class="eye"
            aria-label={showAccount ? "Hide account name" : "Show account name"}
            title={showAccount ? "Hide account name" : "Show account name"}
            aria-pressed={showAccount}
            onclick={toggleAccount}
          >
            <Icon name={showAccount ? "eye-off" : "eye"} size={14} />
          </button>
        </div>
        {#if eyeError}
          <div class="detail bad" role="alert">{eyeError}</div>
        {:else if status.detail && status.state !== "online"}
          <div class="detail">{status.detail}</div>
        {/if}
      </div>
    </div>

    {#if switching && viaSteam}
      <h3>Switch account</h3>
      <AccountCards
        current={status.accountName}
        ondone={() => (switching = false)}
        oncancel={cancelSwitch}
      />
    {:else}
      {#if viaSteam}
        <Row label={method}>
          <Button icon="swap" onclick={startSwitch}>Switch account</Button>
        </Row>
      {/if}
      <Row
        label={viaSteam ? "Log out" : method}
        hint={viaSteam
          ? "The Steam app stays signed in."
          : "Logging out removes the saved login from this PC. The Steam app stays signed in."}
      >
        <Button variant="danger" icon="logout" onclick={logout}>Log out</Button>
      </Row>
    {/if}
  </Section>
{/if}

<style>
  .card {
    display: flex;
    align-items: center;
    gap: 18px;
    min-height: 104px;
    margin-top: 6px;
    padding: 16px 20px;
    background: var(--box);
  }

  .who {
    flex: 1;
    min-width: 0;
  }

  .name {
    overflow: hidden;
    font-family: var(--font-head);
    font-size: 22px;
    font-weight: 700;
    line-height: 28px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: #fff;
  }

  .state {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    line-height: 20px;
    white-space: nowrap;
    color: var(--muted);
  }

  .state .label {
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 700;
    letter-spacing: 0.01em;
    color: #fff;
  }

  /* Logged in: the app's blue, not Steam's green "online". */
  .state .label.in {
    color: var(--accent);
  }

  .account {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Shows or hides the Steam login name. */
  .eye {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 20px;
    margin-left: -2px;
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

  .detail {
    overflow: hidden;
    font-size: 12px;
    line-height: 16px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .detail.bad {
    color: var(--red-text);
  }

  .dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--dim);
  }

  .dot.in {
    background: var(--accent);
  }

  .dot.warn {
    background: var(--gold);
  }

  .dot.bad {
    background: var(--red);
  }

  h3 {
    height: 40px;
    margin: 0;
    padding-top: 8px;
    font-family: var(--font-head);
    font-size: 14px;
    font-weight: 700;
    line-height: 32px;
    letter-spacing: 0.01em;
    text-transform: uppercase;
    color: var(--accent);
  }

  :global(.anim) .eye {
    transition: color 120ms;
  }
</style>
