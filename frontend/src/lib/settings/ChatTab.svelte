<script lang="ts">
  import type * as Cfg from "../../../bindings/mychat/internal/config/models";
  import { app, saveConfig } from "../store.svelte";
  import HotkeyInput from "../ui/HotkeyInput.svelte";
  import Slider from "../ui/Slider.svelte";
  import Toggle from "../ui/Toggle.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  // Settings save the moment they change; a rejected change shows its reason
  // in the row it belongs to.
  let errors = $state<Record<string, string>>({});

  async function set(field: keyof Cfg.Config, patch: Partial<Cfg.Config>): Promise<void> {
    errors[field] = "";
    errors[field] = await saveConfig(patch);
  }
</script>

<Section id="chat-box" title="Chat box">
  <Row
    label="Open chat"
    hint="Press this key in game to open the chat box. Escape cancels a change."
    error={errors.hotkey}
  >
    <HotkeyInput label={app.config.hotkeyLabel} onpick={(hotkey) => set("hotkey", { hotkey })} />
  </Row>
  <Row
    label="Close after sending"
    hint="The chat box closes as soon as your message is sent."
    error={errors.closeAfterSend}
  >
    <Toggle
      label="Close after sending"
      checked={app.config.closeAfterSend}
      onchange={(closeAfterSend) => set("closeAfterSend", { closeAfterSend })}
    />
  </Row>
</Section>

<Section id="pop-ups" title="Pop-ups">
  <Row
    label="Show new message pop-ups"
    hint="A small card where the chat box opens, when a friend writes to you."
    error={errors.showToasts}
  >
    <Toggle
      label="Show new message pop-ups"
      checked={app.config.showToasts}
      onchange={(showToasts) => set("showToasts", { showToasts })}
    />
  </Row>
  <Row
    label="Pop-up duration"
    hint="How long a pop-up stays on screen."
    error={errors.toastSeconds}
    dim={!app.config.showToasts}
  >
    <Slider
      label="Pop-up duration in seconds"
      value={app.config.toastSeconds}
      min={2}
      max={30}
      format={(v) => `${v} s`}
      onchange={(toastSeconds) => set("toastSeconds", { toastSeconds })}
    />
  </Row>
</Section>
