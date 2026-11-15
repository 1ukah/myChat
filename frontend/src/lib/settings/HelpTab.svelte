<script lang="ts">
  import { app, quit } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const commands: [string, string][] = [
    ["/r", "Reply to the last person who whispered you. Type /r and a space, then your message."],
    ["/w name", "Whisper a friend. Start typing the name and pick from the list."],
    ["/p", "Message everyone in your CS2 party."],
    ["/help", "List the commands in the chat box."],
    ["/clear", "Empty the chat log."],
  ];

  const keys = $derived<[string[], string][]>([
    [app.config.hotkeyLabel ? app.config.hotkeyLabel.split(" + ") : ["I"], "Open the chat box in game."],
    [["Enter"], "Send. On an empty line it closes the chat box."],
    [["Esc"], "Close the chat box."],
    [["Tab"], "Next recent person. Shift + Tab goes back."],
    [["↑", "↓"], "Your sent messages, or move in the name list."],
    [["PgUp", "PgDn"], "Scroll the chat log."],
    [["Ctrl", "V"], "Paste."],
  ]);
</script>

<Section id="commands" title="Commands">
  <table>
    <tbody>
      {#each commands as [command, text] (command)}
        <tr>
          <th scope="row"><code>{command}</code></th>
          <td>{text}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</Section>

<Section id="keys" title="Keys">
  <table>
    <tbody>
      {#each keys as [caps, text] (text)}
        <tr>
          <th scope="row">
            <span class="caps">
              {#each caps as cap (cap)}<kbd>{cap}</kbd>{/each}
            </span>
          </th>
          <td>{text}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</Section>

<Section id="about" title="About">
  <p class="note">
    myChat never opens, reads or writes the game process. It uses a normal Windows hotkey and a
    keyboard hook only while the chat box is open.
  </p>
  <Row label="Quit myChat" hint="Closing this window keeps myChat running in the tray.">
    <Button variant="danger" icon="power" onclick={quit}>Quit myChat</Button>
  </Row>
</Section>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
  }

  tr {
    background: var(--row);
    border-bottom: 1px solid var(--line);
  }

  th,
  td {
    height: 44px;
    padding: 8px 16px;
    font-size: 14px;
    font-weight: 400;
    line-height: 20px;
    text-align: left;
    vertical-align: middle;
  }

  th {
    width: 184px;
    white-space: nowrap;
  }

  td {
    color: var(--text-soft);
  }

  code {
    font-family: var(--font-head);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: var(--accent);
  }

  .caps {
    display: inline-flex;
    gap: 6px;
  }

  kbd {
    min-width: 28px;
    height: 24px;
    padding: 0 8px;
    font-family: var(--font-head);
    font-size: 15px;
    font-weight: 700;
    line-height: 22px;
    letter-spacing: 0.06em;
    text-align: center;
    text-transform: uppercase;
    color: var(--text);
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }

  .note {
    margin: 0;
    padding: 16px;
    font-size: 14px;
    line-height: 22px;
    color: var(--text-soft);
    background: var(--row);
    border-bottom: 1px solid var(--line);
    box-shadow: inset 3px 0 0 var(--green);
  }
</style>
