<script lang="ts">
  import { app, logout } from "../store.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Button from "../ui/Button.svelte";
  import Login from "./Login.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const status = $derived(app.status);

  const loggedOut = $derived(status.state === "needs_login" || !status.accountName);
  const name = $derived(status.personaName || status.accountName || "Steam");

  const state = $derived.by(() => {
    switch (status.state) {
      case "online":
        return { tone: "on", ring: "green", text: "Online" } as const;
      case "connecting":
        return { tone: "warn", ring: "accent", text: "Connecting" } as const;
      default:
        return { tone: "bad", ring: "offline", text: "Offline" } as const;
    }
  });
</script>

{#if loggedOut}
  <Section id="steam" title="Sign in to Steam">
    <Login />
  </Section>
{:else}
  <Section id="steam" title="Steam">
    <!-- Status card, in the style of a CS2 spectator player card. -->
    <div class="card">
      <Avatar size={64} ring={state.ring} {name} />
      <div class="who">
        <div class="name"><bdi>{name}</bdi></div>
        <div class="state">
          <span class="dot {state.tone}"></span>
          <span class="label">{state.text}</span>
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

    <Row label="Log out" hint="Removes the saved Steam login from this PC.">
      <Button variant="danger" icon="logout" onclick={logout}>Log out</Button>
    </Row>
  </Section>
{/if}

<style>
  .card {
    display: flex;
    align-items: center;
    gap: 18px;
    min-height: 104px;
    padding: 16px 16px 16px 20px;
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  .who {
    flex: 1;
    min-width: 0;
  }

  .name {
    overflow: hidden;
    font-size: 20px;
    font-weight: 600;
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
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-soft);
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
    font-family: var(--font-head);
    font-size: 32px;
    font-weight: 700;
    line-height: 36px;
    font-variant-numeric: tabular-nums;
    color: var(--text);
  }

  .cap {
    font-family: var(--font-head);
    font-size: 13px;
    font-weight: 600;
    line-height: 16px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    white-space: nowrap;
    color: var(--muted);
  }

  .stats.off .num {
    color: var(--dim);
  }
</style>
