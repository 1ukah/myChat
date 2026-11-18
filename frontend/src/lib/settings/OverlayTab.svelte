<script lang="ts">
  import type * as Cfg from "../../../bindings/mychat/internal/config/models";
  import {
    asLook,
    asSpeed,
    asStyle,
    type AnimationSpeed,
    type AnimationStyle,
  } from "../overlay/motion";
  import { app, previewToast, saveConfig } from "../store.svelte";
  import Button from "../ui/Button.svelte";
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

  const LOOKS = [
    { value: "legacy", label: "Legacy", hint: "The dark CS-style chat." },
    { value: "minimal", label: "Minimal", hint: "Dark blue-grey glass with light text, square and compact; a soft glow, fading out at the edges." },
  ] as const;

  const look = $derived(asLook(app.config.overlayStyle) === "minimal" ? LOOKS[1] : LOOKS[0]);

  const percent = $derived(Math.round(app.config.scale * 100));
  const offsetMax = $derived(Math.max(600, app.config.bottomOffset));
  const still = $derived(!app.config.animations);
</script>

<Section id="appearance" title="Appearance">
  <Row label="Chat theme" hint={look.hint} error={errors.overlayStyle}>
    <Select
      label="Chat theme"
      value={look.value}
      options={LOOKS}
      onchange={(overlayStyle) => set("overlayStyle", { overlayStyle })}
    />
  </Row>
  <Row
    label="Show avatars"
    hint="Friends' pictures next to their names."
    error={errors.showAvatars}
  >
    <Toggle
      label="Show avatars"
      checked={app.config.showAvatars}
      onchange={(showAvatars) => set("showAvatars", { showAvatars })}
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
    hint="Motion when the chat box, pop-ups and new messages appear or go."
    error={errors.animations}
  >
    <Toggle
      label="Animations"
      checked={app.config.animations}
      onchange={(animations) => set("animations", { animations })}
    />
  </Row>
  <Row label="Animation style" error={errors.animationStyle} dim={still}>
    <Select
      label="Animation style"
      value={asStyle(app.config.animationStyle)}
      options={STYLES}
      disabled={still}
      onchange={(animationStyle) => set("animationStyle", { animationStyle })}
    />
  </Row>
  <Row label="Animation speed" error={errors.animationSpeed} dim={still}>
    <Select
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
      ? "Shows a sample pop-up where it appears in game."
      : "Pop-ups are turned off in the Chat tab."}
  >
    <Button variant="accent" icon="eye" disabled={!app.config.showToasts} onclick={previewToast}>
      Preview pop-up
    </Button>
  </Row>
</Section>
