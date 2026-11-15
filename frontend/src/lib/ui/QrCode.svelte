<script lang="ts">
  // QR code from rows of "1" (dark) / "0" (light). Dark on white with the
  // standard 4-module quiet zone, drawn as one crisp path so phones scan it
  // at any DPI.
  let {
    rows,
    size = 232,
    label = "QR code",
  }: {
    rows: string[];
    /** Largest side in CSS px. The code uses the biggest whole-pixel module that fits. */
    size?: number;
    label?: string;
  } = $props();

  const QUIET = 4;

  const total = $derived(rows.length + QUIET * 2);
  const px = $derived(Math.max(1, Math.floor(size / total)) * total);

  // One rectangle per horizontal run of dark modules.
  const path = $derived.by(() => {
    let d = "";
    rows.forEach((row, y) => {
      let x = 0;
      while (x < row.length) {
        if (row[x] !== "1") {
          x++;
          continue;
        }
        let end = x;
        while (end < row.length && row[end] === "1") end++;
        d += `M${x + QUIET} ${y + QUIET}h${end - x}v1h${x - end}z`;
        x = end;
      }
    });
    return d;
  });
</script>

<svg
  class="qr"
  width={px}
  height={px}
  viewBox="0 0 {total} {total}"
  shape-rendering="crispEdges"
  role="img"
  aria-label={label}
>
  <rect width={total} height={total} fill="#fff" />
  <path d={path} fill="#000" />
</svg>

<style>
  .qr {
    display: block;
    flex: none;
  }
</style>
