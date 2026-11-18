const TAU = Math.PI * 2;

function rng(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function topoTile(size = 560, seed = 5): string {
  const N = 64;
  const W = N + 1;
  const rand = rng(seed);

  const waves: { kx: number; ky: number; phase: number; amp: number }[] = [];
  while (waves.length < 6) {
    const kx = Math.floor(rand() * 7) - 3;
    const ky = Math.floor(rand() * 7) - 3;
    if (kx === 0 && ky === 0) continue;
    waves.push({ kx, ky, phase: rand() * TAU, amp: 1 / Math.hypot(kx, ky) });
  }

  const h = new Float32Array(W * W);
  for (let j = 0; j <= N; j++) {
    for (let i = 0; i <= N; i++) {
      let s = 0;
      for (const w of waves) s += w.amp * Math.sin((TAU * (w.kx * i + w.ky * j)) / N + w.phase);
      h[j * W + i] = s;
    }
  }

  const step = size / N;
  const p = (n: number): number => Math.round(n * step * 10) / 10;
  let d = "";

  for (let l = -7; l <= 7; l++) {
    const level = l * 0.34 + 0.013;
    for (let j = 0; j < N; j++) {
      for (let i = 0; i < N; i++) {
        const a = h[j * W + i] - level;
        const b = h[j * W + i + 1] - level;
        const c = h[(j + 1) * W + i + 1] - level;
        const e = h[(j + 1) * W + i] - level;

        const pts: number[] = [];
        if (a < 0 !== b < 0) pts.push(i + a / (a - b), j);
        if (b < 0 !== c < 0) pts.push(i + 1, j + b / (b - c));
        if (e < 0 !== c < 0) pts.push(i + e / (e - c), j + 1);
        if (a < 0 !== e < 0) pts.push(i, j + a / (a - e));

        for (let k = 0; k + 3 < pts.length; k += 4) {
          d += `M${p(pts[k])} ${p(pts[k + 1])}L${p(pts[k + 2])} ${p(pts[k + 3])}`;
        }
      }
    }
  }

  return (
    `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">` +
    `<path d="${d}" fill="none" stroke="#fff" stroke-opacity="0.1" stroke-width="1" stroke-linecap="round"/>` +
    `</svg>`
  );
}
