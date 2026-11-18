<script lang="ts">
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
    --thumb: 12px;
    -webkit-appearance: none;
    appearance: none;
    width: 200px;
    height: 28px;
    margin: 0;
    background: transparent;
    cursor: pointer;
  }

  input::-webkit-slider-runnable-track {
    height: 4px;
    background: linear-gradient(
      to right,
      #c4c9ce calc(var(--thumb) / 2 + (100% - var(--thumb)) * var(--p)),
      rgba(255, 255, 255, 0.3) 0
    );
  }

  input::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: var(--thumb);
    height: 10px;
    margin-top: -3px;
    background: #e3e6e9;
    border: 0;
    border-radius: 1px;
  }

  input:hover::-webkit-slider-thumb {
    background: var(--accent);
  }

  input:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .value {
    width: 68px;
    height: 32px;
    font-family: var(--font-ui);
    font-size: var(--fs-sm);
    font-weight: 600;
    line-height: 30px;
    text-align: center;
    font-variant-numeric: tabular-nums;
    color: #fff;
    border: 1px solid var(--line-strong);
    border-radius: 2px;
  }
</style>
