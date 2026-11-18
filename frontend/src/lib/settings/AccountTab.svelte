<script lang="ts">
  import { app, cancelLogin, logout, refreshLocalLogins, saveConfig } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
  import NameEye from "../ui/NameEye.svelte";
  import NameMask from "../ui/NameMask.svelte";
  import AccountCards from "./AccountCards.svelte";
  import Login from "./Login.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const status = $derived(app.status);

  const loggedOut = $derived(
    !status.accountName || (status.state === "needs_login" && app.login.phase !== "done"),
  );
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

  let switching = $state(false);

  function startSwitch(): void {
    refreshLocalLogins();
    switching = true;
  }

  function cancelSwitch(): void {
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
    <div class="card">
      <Avatar round size={64} ring={conn.ring} src={status.avatar} {name} />
      <div class="who">
        <div class="name"><bdi>{name}</bdi></div>
        <div class="state">
          <span class="dot {conn.tone}"></span>
          <span class="label {conn.tone}">{conn.text}</span>
          {#if showAccount}
            <span class="account">· {status.accountName}</span>
          {:else}
            <span class="account veiled"><span aria-hidden="true">·</span><NameMask /></span>
          {/if}
          <span class="eye"><NameEye shown={showAccount} onclick={toggleAccount} /></span>
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
    font-size: var(--fs-xl);
    font-weight: 600;
    line-height: 28px;
    letter-spacing: -0.01em;
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
    font-weight: 600;
    color: #fff;
  }

  .state .label.in {
    color: var(--accent);
  }

  .account {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .account.veiled {
    flex: none;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .eye {
    flex: none;
    display: flex;
    margin-left: -2px;
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
    font-size: var(--fs-md);
    font-weight: 600;
    line-height: 32px;
    color: var(--accent);
  }

  :global(.anim) .account {
    animation: account-in 180ms ease-out both;
  }

  @keyframes account-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
