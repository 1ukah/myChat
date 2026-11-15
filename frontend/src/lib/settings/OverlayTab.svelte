<script lang="ts">
  import type * as Cfg from "../../../bindings/mychat/internal/config/models";
  import {
    asSpeed,
    asStyle,
    type AnimationSpeed,
    type AnimationStyle,
  } from "../overlay/motion";
  import { app, previewToast, saveConfig } from "../store.svelte";
  import Button from "../ui/Button.svelte";
  import Segmented from "../ui/Segmented.svelte";
  import Select from "../ui/Select.svelte";
  import Slider from "../ui/Slider.svelte";
  import Toggle from "../ui/Toggle.svelte";
  import AnimationPreview from "./AnimationPreview.svelte";
  import Row from "./Row.svelte";
  import Section from "./Section.svelte";

  let errors = $state<Record<string, string>>({});

  async function set(field: keyof Cfg.Config, patch: Partial<Cfg.Config>): Promise<void> {
    errors[field] = "";
    errors[field] = await saveConfig(patch);
  }

  const STYLES: readonly { value: AnimationStyle; label: string }[] = [
    { value: "fade", label: "Fade" },
    { value: "slide-up", label: "Slide up" },
    { value: "slide-down", label: "Slide down" },
    { value: "slide-left", label: "Slide left" },
    { value: "slide-right", label: "Slide right" },
    { value: "typewriter", label: "Typewriter" },
  ];

  const SPEEDS: readonly { value: AnimationSpeed; label: string }[] = [
    { value: "fast", label: "Fast" },
    { value: "normal", label: "Normal" },
    { value: "slow", label: "Slow" },
  ];

  const percent = $derived(Math.round(app.config.scale * 100));
  // The slider covers 0–600 px; a larger value set by hand still fits.
  const offsetMax = $derived(Math.max(600, app.config.bottomOffset));
  const still = $derived(!app.config.animations);
</script>

<Section id="appearance" title="Appearance">
  <Row
    label="Show avatars"
    hint="Friends' Steam pictures next to their messages and names."
    error={errors.showAvatars}
  >
    <Toggle
      label="Show avatars"
      checked={app.config.showAvatars}
      onchange={(showAvatars) => set("showAvatars", { showAvatars })}
    />
  </Row>
  <Row
    label="Blur behind (experimental, may look solid)"
    hint="Windows blur can turn the whole overlay opaque. Leave it off for a see-through chat box."
    error={errors.blur}
  >
    <Toggle
      label="Blur behind (experimental, may look solid)"
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

<Section id="animation" title="Animation">
  <Row
    label="Animations"
    hint="Motion when the chat box, pop-ups and new messages appear."
    error={errors.animations}
  >
    <Toggle
      label="Animations"
      checked={app.config.animations}
      onchange={(animations) => set("animations", { animations })}
    />
  </Row>
  <Row
    label="Animation style"
    hint="How the chat box, pop-ups and new lines come in and go out."
    error={errors.animationStyle}
    dim={still}
  >
    <Select
      label="Animation style"
      value={asStyle(app.config.animationStyle)}
      options={STYLES}
      disabled={still}
      onchange={(animationStyle) => set("animationStyle", { animationStyle })}
    />
  </Row>
  <Row
    label="Animation speed"
    hint="Fast 120 ms · Normal 200 ms · Slow 320 ms"
    error={errors.animationSpeed}
    dim={still}
  >
    <Segmented
      label="Animation speed"
      value={asSpeed(app.config.animationSpeed)}
      options={SPEEDS}
      disabled={still}
      onchange={(animationSpeed) => set("animationSpeed", { animationSpeed })}
    />
  </Row>
  <AnimationPreview />
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
    label="Preview in game position"
    hint={app.config.showToasts
      ? "Shows a sample pop-up where it will appear in game."
      : "Pop-ups are turned off in the Chat tab."}
  >
    <Button variant="accent" icon="eye" disabled={!app.config.showToasts} onclick={previewToast}>
      Preview pop-up
    </Button>
  </Row>
</Section>
