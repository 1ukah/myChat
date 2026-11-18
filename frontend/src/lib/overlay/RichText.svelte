<script lang="ts" module>
  const segmenter =
    typeof Intl !== "undefined" && "Segmenter" in Intl
      ? new Intl.Segmenter(undefined, { granularity: "grapheme" })
      : null;

  function graphemes(text: string): string[] {
    return segmenter ? Array.from(segmenter.segment(text), (s) => s.segment) : Array.from(text);
  }

  const TOTAL_MS = 600;
  const PER_CHARACTER_MS = 22;
  const FRAME_MS = 16;
</script>

<script lang="ts">
  import { app, clickSpot } from "../store.svelte";
  import { splitLinks } from "./links";
  import { linkSpot } from "./spots";

  let {
    text,
    at,
    links = true,
    typed = false,
    delay = 0,
  }: {
    text: string;
    at: number;
    links?: boolean;
    typed?: boolean;
    delay?: number;
  } = $props();

  interface Run {
    text: string;
    url?: string;
    spot: string;
    pieces: { text: string; ms: number }[];
  }

  const runs = $derived.by((): Run[] => {
    const parts = splitLinks(text);
    let n = 0;
    const chars = parts.map((p) => {
      const g = graphemes(p.text);
      n += g.length;
      return g;
    });
    const total = Math.min(TOTAL_MS, n * PER_CHARACTER_MS);
    const steps = Math.max(1, Math.min(n, Math.floor(total / FRAME_MS)));
    const size = Math.max(1, Math.ceil(n / steps));
    const stepMs = n > 0 ? total / Math.ceil(n / size) : 0;

    let link = 0;
    let offset = 0;
    return parts.map((p, i) => {
      const run: Run = { ...p, spot: p.url && links ? linkSpot(at, link++) : "", pieces: [] };
      if (typed) {
        const g = chars[i];
        for (let k = 0; k < g.length; ) {
          const step = Math.floor((offset + k) / size);
          const end = Math.min(g.length, (step + 1) * size - offset);
          run.pieces.push({ text: g.slice(k, end).join(""), ms: Math.round(delay + step * stepMs) });
          k = end;
        }
      }
      offset += chars[i].length;
      return run;
    });
  });
</script>

{#snippet body(run: Run)}{#if typed}{#each run.pieces as piece, j (j)}<span class="piece" style:animation-delay="{piece.ms}ms">{piece.text}</span>{/each}{:else}{run.text}{/if}{/snippet}

<span class="rich" class:typed>{#each runs as run, i (i)}{#if run.url}<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions --><span class="url" class:hover={run.spot !== "" && app.hover === run.spot} data-spot={run.spot || undefined} data-url={run.spot ? run.url : undefined} onclick={run.spot ? () => clickSpot(run.spot) : undefined}>{@render body(run)}</span>{:else}{@render body(run)}{/if}{/each}</span>

<style>
  .url {
    color: var(--link, #9ed2ff);
    text-decoration: underline;
    text-decoration-color: color-mix(in srgb, currentColor 45%, transparent);
    text-decoration-thickness: 1px;
    text-underline-offset: 2px;
    cursor: pointer;
  }

  .url.hover,
  .url:hover {
    color: var(--link-hover, #d4ecff);
    text-decoration-color: currentColor;
  }

  :global(.anim) .typed .piece {
    animation: type-in 1ms step-end both;
  }

  :global(.anim .leave) .typed .piece {
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
