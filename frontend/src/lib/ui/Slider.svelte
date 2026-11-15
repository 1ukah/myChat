<script lang="ts">
  // Thin track, small white handle, boxed value on the right ("93%").
  // The value box follows the handle while dragging; `onchange` fires once,
  // when the handle is released, so each drag saves a single time.
  let {
    value,
    min,
    max,
    step = 1,
    label,
    disabled = false,
    format = (v: number) => String(v),
    onchange,
  }: {
    value: number;
    min: number;
    max: number;
    step?: number;
    /** Accessible name (the visible label lives in the row). */
    label: string;
    disabled?: boolean;
    format?: (value: number) => string;
    onchange: (value: number) => void;
  } = $props();

  let dragging = $state<number | null>(null);
  const shown = $derived(dragging ?? value);
  const fraction = $derived(
    max > min ? (Math.min(Math.max(shown, min), max) - min) / (max - min) : 0,
  );

  type RangeEvent = Event & { currentTarget: HTMLInputElement };

  function oninput(e: RangeEvent): void {
    dragging = e.currentTarget.valueAsNumber;
  }

  function commit(e: RangeEvent): void {
    const v = e.currentTarget.valueAsNumber;
    dragging = null;
    if (v !== value) onchange(v);
  }
</script>

<div class="slider" class:disabled>
  <input
    type="range"
    {min}
    {max}
    {step}
    value={shown}
    {disabled}
    aria-label={label}
    aria-valuetext={format(shown)}
    style:--p={fraction}
    {oninput}
    onchange={commit}
  />
  <output class="value">{format(shown)}</output>
</div>

<style>
  .slider {
    display: inline-flex;
    align-items: center;
    gap: 16px;
  }

  .slider.disabled {
    opacity: 0.4;
  }

  input {
    --thumb: 8px;
    -webkit-appearance: none;
    appearance: none;
    width: 220px;
    height: 28px;
    margin: 0;
    background: transparent;
    cursor: pointer;
  }

  /* The filled part ends at the centre of the handle. */
  input::-webkit-slider-runnable-track {
    height: 2px;
    background: linear-gradient(
      to right,
      var(--text) calc(var(--thumb) / 2 + (100% - var(--thumb)) * var(--p)),
      rgba(255, 255, 255, 0.24) 0
    );
  }

  input::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: var(--thumb);
    height: 18px;
    margin-top: -8px;
    background: #fff;
    border: 0;
    border-radius: 1px;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.45);
  }

  input:hover::-webkit-slider-thumb {
    background: var(--accent);
  }

  input:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .value {
    width: 72px;
    height: 28px;
    font-family: var(--font-head);
    font-size: 16px;
    font-weight: 700;
    line-height: 26px;
    letter-spacing: 0.06em;
    text-align: center;
    text-transform: uppercase;
    font-variant-numeric: tabular-nums;
    color: var(--text);
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }
</style>
