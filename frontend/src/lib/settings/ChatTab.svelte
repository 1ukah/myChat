<script lang="ts">
  import type * as Cfg from "../../../bindings/mychat/internal/config/models";
  import { app, saveConfig } from "../store.svelte";
  import HotkeyInput from "../ui/HotkeyInput.svelte";
  import Segmented from "../ui/Segmented.svelte";
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

  const LAYOUTS = [
    { value: "stack", label: "Stack", hint: "Every message gets its own box." },
    { value: "group", label: "Group", hint: "Messages in a row from one person share a box." },
  ] as const;

  const COUNTS = ["1", "2", "3", "4", "5"].map((value) => ({ value, label: value }));

  const layout = $derived(app.config.toastLayout === "stack" ? LAYOUTS[0] : LAYOUTS[1]);
  const count = $derived(String(Math.min(5, Math.max(1, Math.round(app.config.toastMax) || 3))));
  const off = $derived(!app.config.showToasts);
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
  <Row label="Multiple messages" hint={layout.hint} error={errors.toastLayout} dim={off}>
    <Segmented
      label="Multiple messages"
      value={layout.value}
      options={LAYOUTS}
      onchange={(toastLayout) => set("toastLayout", { toastLayout })}
    />
  </Row>
  <Row
    label="Messages shown"
    hint="How many recent messages stay on screen at once."
    error={errors.toastMax}
    dim={off}
  >
    <Segmented
      label="Messages shown"
      value={count}
      options={COUNTS}
      onchange={(v) => set("toastMax", { toastMax: Number(v) })}
    />
  </Row>
</Section>
