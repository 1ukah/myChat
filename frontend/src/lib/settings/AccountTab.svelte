<script lang="ts">
  import { initial } from "../format";
  import { app, cancelLogin, logout, startLogin } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Icon from "../ui/Icon.svelte";
  import QrCode from "../ui/QrCode.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const status = $derived(app.status);
  const login = $derived(app.login);

  const loggedOut = $derived(status.state === "needs_login" || !status.accountName);
  const name = $derived(
    loggedOut ? "Not logged in" : status.personaName || status.accountName || "Steam",
  );

  const state = $derived.by(() => {
    if (loggedOut) return { tone: "off", text: "Not logged in" };
    switch (status.state) {
      case "online":
        return { tone: "on", text: "Online" };
      case "connecting":
        return { tone: "warn", text: "Connecting" };
      default:
        return { tone: "bad", text: "Offline" };
    }
  });

  const qr = $derived(login.qr ?? []);
</script>

<Section id="steam" title="Steam">
  <!-- Status card, in the style of the CS2 friends panel. -->
  <div class="card">
    <div class="avatar" class:off={loggedOut} aria-hidden="true">
      {#if loggedOut}<Icon name="user" size={28} />{:else}{initial(name)}{/if}
    </div>
    <div class="who">
      <div class="name"><bdi>{name}</bdi></div>
      <div class="state">
        <span class="dot {state.tone}"></span>
        <span class="label">{state.text}</span>
        {#if !loggedOut && status.accountName}
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

  {#if login.phase === "waiting"}
    <div class="login">
      <div class="qr">
        {#if qr.length > 0}
          <QrCode rows={qr} size={232} label="Steam login QR code" />
        {:else}
          <span class="wait">Getting a code…</span>
        {/if}
      </div>
      <div class="guide">
        <h3>Scan with the Steam mobile app</h3>
        <ol>
          <li><span class="n">1</span><span>Open the <b>Steam</b> app on your phone.</span></li>
          <li>
            <span class="n">2</span>
            <span>Tap the <b>Steam Guard</b> icon <span class="shield"><Icon name="shield" /></span> at the bottom.</span>
          </li>
          <li><span class="n">3</span><span>Scan this code and approve the login.</span></li>
        </ol>
        <p class="fine">Your password never touches myChat.</p>
        <div class="actions">
          <Button onclick={cancelLogin}>Cancel</Button>
        </div>
      </div>
    </div>
  {:else if login.phase === "error"}
    <div class="failed" role="alert">
      <div class="mark"><Icon name="alert" size={20} /></div>
      <div class="what">
        <h3>Login failed</h3>
        <p>{login.error || "Steam did not complete the login."}</p>
      </div>
      <Button variant="accent" icon="retry" onclick={startLogin}>Try again</Button>
    </div>
  {:else if loggedOut}
    <div class="intro">
      <div class="pitch">
        <h3>Log in with the Steam mobile app.</h3>
        <p>Your password never touches myChat.</p>
      </div>
      <Button variant="go" size="lg" icon="qr" onclick={startLogin}>Log in with QR</Button>
    </div>
  {:else}
    <Row label="Log out" hint="Removes the saved Steam login from this PC.">
      <Button variant="danger" icon="logout" onclick={logout}>Log out</Button>
    </Row>
  {/if}
</Section>

<style>
  /* ---------- status card ---------- */

  .card {
    display: flex;
    align-items: center;
    gap: 16px;
    min-height: 96px;
    padding: 16px;
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  .avatar {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    font-family: var(--font-head);
    font-size: 36px;
    font-weight: 700;
    line-height: 1;
    color: var(--accent);
    background: var(--accent-fill);
    box-shadow: inset 0 0 0 1px var(--accent-line);
  }

  .avatar.off {
    color: var(--muted);
    background: rgba(255, 255, 255, 0.05);
    box-shadow: inset 0 0 0 1px var(--line);
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

  /* ---------- logged out: call to action ---------- */

  .intro {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 24px 16px;
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;
    color: var(--text);
  }

  .pitch p,
  .what p {
    margin: 4px 0 0;
    font-size: 14px;
    line-height: 20px;
    color: var(--muted);
  }

  /* ---------- QR login ---------- */

  .login {
    display: flex;
    align-items: center;
    gap: 32px;
    padding: 24px 16px;
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  /* Always white and always the same size, so nothing moves when the code
     arrives or Steam refreshes it. */
  .qr {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 248px;
    height: 248px;
    background: #fff;
    border-radius: 2px;
  }

  .wait {
    font-family: var(--font-head);
    font-size: 16px;
    font-weight: 600;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #6b7279;
  }

  .guide {
    flex: 1;
    min-width: 0;
  }

  ol {
    margin: 16px 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 6px 0;
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

  .shield {
    display: inline-block;
    vertical-align: -3px;
    color: var(--accent);
  }

  .fine {
    margin: 12px 0 0;
    font-size: 13px;
    line-height: 20px;
    color: var(--muted);
  }

  .actions {
    margin-top: 20px;
  }

  /* ---------- login error ---------- */

  .failed {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px 16px;
    background: rgba(200, 53, 47, 0.12);
    border-bottom: 1px solid var(--line);
    box-shadow: inset 3px 0 0 var(--red);
  }

  .mark {
    flex: none;
    color: var(--red-text);
  }

  .what {
    flex: 1;
    min-width: 0;
  }

  .what p {
    overflow-wrap: anywhere;
    color: #ffb1aa;
  }
</style>
