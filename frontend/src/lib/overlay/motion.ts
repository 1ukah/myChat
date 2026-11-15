// Animation style and speed, shared by the overlay and by the preview tile
// in the settings window.

export const ANIMATION_STYLES = [
  "fade",
  "slide-up",
  "slide-down",
  "slide-left",
  "slide-right",
  "typewriter",
] as const;

export type AnimationStyle = (typeof ANIMATION_STYLES)[number];

export const ANIMATION_SPEEDS = ["fast", "normal", "slow"] as const;

export type AnimationSpeed = (typeof ANIMATION_SPEEDS)[number];

/** The two chat looks (`overlayStyle`). Anything unknown counts as legacy. */
export type Look = "legacy" | "minimal";

export function asLook(value: string): Look {
  return value === "minimal" ? "minimal" : "legacy";
}

/** Entry time per speed, and the time Go waits before it hides the window. */
export const SPEED_MS: Record<AnimationSpeed, number> = { fast: 120, normal: 200, slow: 320 };

/** The exit animation must be over this long before Go hides the window. */
const EXIT_MARGIN_MS = 20;

export interface Motion {
  /** Animations are enabled at all. */
  on: boolean;
  style: AnimationStyle;
  enterMs: number;
  exitMs: number;
  /** New message text is revealed character by character. */
  typewriter: boolean;
}

export function asStyle(value: string): AnimationStyle {
  return (ANIMATION_STYLES as readonly string[]).includes(value)
    ? (value as AnimationStyle)
    : "slide-up";
}

export function asSpeed(value: string): AnimationSpeed {
  return (ANIMATION_SPEEDS as readonly string[]).includes(value)
    ? (value as AnimationSpeed)
    : "normal";
}

/**
 * @param hideDelayMs  from ModeInfo: how long Go keeps the window after
 *                     `hidden`. Leave out where nothing hides (the preview).
 */
export function motionOf(
  m: { animations: boolean; animationStyle: string; animationSpeed: string },
  hideDelayMs = 0,
): Motion {
  const style = asStyle(m.animationStyle);
  const enterMs = SPEED_MS[asSpeed(m.animationSpeed)];
  const hide = hideDelayMs > 0 ? Math.min(hideDelayMs, enterMs) : enterMs;
  return {
    on: m.animations,
    style,
    enterMs,
    exitMs: Math.max(0, hide - EXIT_MARGIN_MS),
    typewriter: m.animations && style === "typewriter",
  };
}
