<script lang="ts">
  import { app, cancelLogin, logout, refreshLocalLogins } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
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
  const name = $derived(status.personaName || status.accountName || "Steam");
  const viaSteam = $derived(status.source === "steam");

  const conn = $derived.by(() => {
    switch (status.state) {
      case "online":
        return { tone: "on", ring: "green", text: "Online" } as const;
      case "connecting":
      case "needs_login":
        return { tone: "warn", ring: "accent", text: "Connecting" } as const;
      default:
        return { tone: "bad", ring: "offline", text: "Offline" } as const;
    }
  });

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
      <Avatar round size={64} ring={conn.ring} {name} />
      <div class="who">
        <div class="name"><bdi>{name}</bdi></div>
        <div class="state">
          <span class="dot {conn.tone}"></span>
          <span class="label">{conn.text}</span>
          {#if status.accountName}
            <span class="account">· {status.accountName}</span>
          {/if}
        </div>
        {#if status.detail && status.state !== "online"}
          <div class="detail">{status.detail}</div>
        {/if}
      </div>
      <div class="stats" class:off={status.state !== "online"}>
        <div class="stat">
          <div class="num">{status.state === "online" ? status.friendsOnline : "–"}</div>
          <div class="cap">Friends online</div>
        </div>
        <div class="stat">
          <div class="num">
            {status.state === "online" && status.partySize > 0 ? status.partySize : "–"}
          </div>
          <div class="cap">In your party</div>
        </div>
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
      <Row
        label={viaSteam ? "Using Steam on this PC" : "QR login"}
        hint={viaSteam
          ? "myChat uses the login the Steam app saved on this PC. It never leaves this PC."
          : "myChat keeps its own session from the QR code you scanned."}
      >
        {#if viaSteam}
          <Button icon="swap" onclick={startSwitch}>Switch account</Button>
        {/if}
      </Row>
      <Row
        label="Log out"
        hint={viaSteam
          ? "Ends myChat's session. The Steam app stays signed in."
          : "Removes myChat's saved login from this PC."}
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
    padding: 16px 0 16px 20px;
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
    text-transform: uppercase;
    color: #fff;
  }

  .account {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .detail {
    overflow: hidden;
    font-size: 12px;
    line-height: 16px;
    white-space: nowrap;
    text-overflow: ellipsis;
    color: var(--muted);
  }

  .dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--dim);
  }

  .dot.on {
    background: var(--green);
  }

  .dot.warn {
    background: var(--gold);
  }

  .dot.bad {
    background: var(--red);
  }

  .stats {
    flex: none;
    display: flex;
  }

  .stat {
    min-width: 112px;
    padding: 0 20px;
    text-align: center;
    border-left: 1px solid var(--line);
  }

  .num {
    font-family: var(--font-cond);
    font-size: 38px;
    font-weight: 700;
    line-height: 38px;
    font-variant-numeric: tabular-nums;
    color: #fff;
  }

  .cap {
    font-family: var(--font-cond);
    font-size: 17px;
    font-weight: 500;
    line-height: 18px;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--muted);
  }

  .stats.off .num {
    color: var(--dim);
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
</style>
