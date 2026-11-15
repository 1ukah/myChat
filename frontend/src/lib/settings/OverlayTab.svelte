<script lang="ts">
  import type * as Cfg from "../../../bindings/mychat/internal/config/models";
  import { app, previewToast, saveConfig } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Slider from "../ui/Slider.svelte";
  import Toggle from "../ui/Toggle.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  let errors = $state<Record<string, string>>({});

  async function set(field: keyof Cfg.Config, patch: Partial<Cfg.Config>): Promise<void> {
    errors[field] = "";
    errors[field] = await saveConfig(patch);
  }

  const percent = $derived(Math.round(app.config.scale * 100));
  // The slider covers 0–600 px; a larger value set by hand still fits.
  const offsetMax = $derived(Math.max(600, app.config.bottomOffset));
</script>

<Section id="appearance" title="Appearance">
  <Row
    label="Animations"
    hint="Slide and fade when the chat box and pop-ups appear."
    error={errors.animations}
  >
    <Toggle
      label="Animations"
      checked={app.config.animations}
      onchange={(animations) => set("animations", { animations })}
    />
  </Row>
  <Row label="Background blur" hint="Turn off on low-end PCs" error={errors.blur}>
    <Toggle
      label="Background blur"
      checked={app.config.blur}
      onchange={(blur) => set("blur", { blur })}
    />
  </Row>
  <Row label="Size" hint="Scale of the chat box and pop-ups." error={errors.scale}>
    <Slider
      label="Overlay size in percent"
      value={percent}
      min={75}
      max={150}
      step={5}
      format={(v) => `${v}%`}
      onchange={(v) => set("scale", { scale: v / 100 })}
    />
  </Row>
</Section>

<Section id="position" title="Position">
  <Row
    label="Distance from bottom"
    hint="Gap between the chat box and the bottom edge of the screen."
    error={errors.bottomOffset}
  >
    <Slider
      label="Distance from the bottom of the screen in pixels"
      value={app.config.bottomOffset}
      min={0}
      max={offsetMax}
      step={10}
      format={(v) => `${v} px`}
      onchange={(bottomOffset) => set("bottomOffset", { bottomOffset })}
    />
  </Row>
  <Row
    label="Preview"
    hint={app.config.showToasts
      ? "Shows a sample pop-up where it will appear in game."
      : "Pop-ups are turned off in the Chat tab."}
  >
    <Button variant="accent" icon="eye" disabled={!app.config.showToasts} onclick={previewToast}>
      Preview pop-up
    </Button>
  </Row>
</Section>
