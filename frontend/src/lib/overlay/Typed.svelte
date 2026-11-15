<script lang="ts" module>
  // Whole grapheme clusters, so an emoji or a letter with accents never splits.
  const segmenter =
    typeof Intl !== "undefined" && "Segmenter" in Intl
      ? new Intl.Segmenter(undefined, { granularity: "grapheme" })
      : null;

  function graphemes(text: string): string[] {
    return segmenter ? Array.from(segmenter.segment(text), (s) => s.segment) : Array.from(text);
  }

  /** The whole reveal never takes longer than this. */
  const TOTAL_MS = 600;
  /** Pace for short texts. */
  const PER_CHARACTER_MS = 22;
  /** No point stepping faster than the screen refreshes. */
  const FRAME_MS = 16;
</script>

<script lang="ts">
  // Typewriter reveal: the text appears a few characters at a time.
  //
  // There are no timers. The text is laid out in full from the start (so
  // nothing reflows), cut into at most ~37 pieces, and each piece is switched
  // from invisible to visible by a one-step CSS animation with its own delay.
  // Long texts get more characters per step, so the reveal is always ≤ 600 ms.
  // Without `.anim` (animations off, reduced motion) the text is just there.
  let {
    text,
    delay = 0,
  }: {
    text: string;
    /** Wait before the first character, in ms (for example a panel fading in). */
    delay?: number;
  } = $props();

  const plan = $derived.by(() => {
    const all = graphemes(text);
    const total = Math.min(TOTAL_MS, all.length * PER_CHARACTER_MS);
    const steps = Math.max(1, Math.min(all.length, Math.floor(total / FRAME_MS)));
    const size = Math.max(1, Math.ceil(all.length / steps));
    const pieces: string[] = [];
    for (let i = 0; i < all.length; i += size) pieces.push(all.slice(i, i + size).join(""));
    return { pieces, stepMs: pieces.length > 0 ? total / pieces.length : 0 };
  });
</script>

<span class="typed"
  >{#each plan.pieces as piece, i (i)}<span
      style:animation-delay="{Math.round(delay + i * plan.stepMs)}ms">{piece}</span
    >{/each}</span
>

<style>
  :global(.anim) .typed > span {
    animation: type-in 1ms step-end both;
  }

  /* Closing: show the rest at once and stop. */
  :global(.anim .leave) .typed > span {
    animation: none;
  }

  @keyframes type-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
