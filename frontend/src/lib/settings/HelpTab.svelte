<script lang="ts">
  import { app, quit } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  const commands: [string, string][] = [
    ["/r", "Reply to the last person who whispered you."],
    ["/w name", "Whisper a friend. Type a name or a nickname and pick from the list."],
    ["/help", "List the commands in the chat box."],
    ["/clear", "Empty the chat log."],
  ];

  const keys = $derived<[string[], string][]>([
    [app.config.hotkeyLabel ? app.config.hotkeyLabel.split(" + ") : ["I"], "Open the chat box in game."],
    [
      ["Enter"],
      !app.config.closeAfterSend && app.config.doubleEnterClose
        ? "Send. On an empty line, press twice quickly to close the chat box."
        : "Send. On an empty line it closes the chat box.",
    ],
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
    The game process is never opened, read or written. Keys are read only while the chat box is
    open.
  </p>
  <Row label="Quit myChat" hint="Closing this window only hides it to the tray.">
    <Button variant="danger" icon="power" onclick={quit}>Quit myChat</Button>
  </Row>
</Section>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
  }

  tr {
    border-bottom: 1px solid var(--line);
  }

  th,
  td {
    height: 48px;
    padding: 8px 16px 8px 0;
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

  td {
    padding-right: 0;
  }

  code {
    font-family: var(--font-ui);
    font-size: var(--fs-md);
    font-weight: 600;
    color: var(--accent);
  }

  .caps {
    display: inline-flex;
    gap: 6px;
  }

  kbd {
    min-width: 32px;
    height: 28px;
    padding: 0 8px;
    font-family: var(--font-ui);
    font-size: var(--fs-sm);
    font-weight: 600;
    line-height: 26px;
    text-align: center;
    color: #fff;
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }

  .note {
    margin: 0;
    padding: 12px 0 14px;
    font-size: 14px;
    line-height: 22px;
    color: var(--text-soft);
    border-bottom: 1px solid var(--line);
  }
</style>
