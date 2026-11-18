import * as M from "../../../bindings/mychat/models";
import * as C from "../../../bindings/mychat/internal/chat/models";
import { setSpots } from "../store.svelte";

export const pictureSpot = (m: C.Media): string => `picture:${m.id}`;
export const nameSpot = (peerId: string, at: number): string =>
  /^\d+$/.test(peerId) ? `name:${peerId}@${at}` : "";
export const linkSpot = (at: number, which: number | string): string => `link:${at}.${which}`;

interface Source {
  area: HTMLElement;
  media: Map<string, C.Media>;
  key: string;
  settle: number;
}

let source: (() => Source | null) | null = null;
let last = "";
let frame = 0;
let later = 0;

export function spotSource(get: () => Source | null): void {
  source = get;
}

export function scheduleSpots(): void {
  if (!source) return;
  cancelAnimationFrame(frame);
  clearTimeout(later);
  frame = requestAnimationFrame(measure);
  const s = source();
  if (s) later = window.setTimeout(measure, s.settle + 80);
}

interface Box {
  l: number;
  t: number;
  r: number;
  b: number;
}

function box(r: DOMRect): Box {
  return { l: r.left, t: r.top, r: r.right, b: r.bottom };
}

function clip(a: Box, b: Box): Box {
  return { l: Math.max(a.l, b.l), t: Math.max(a.t, b.t), r: Math.min(a.r, b.r), b: Math.min(a.b, b.b) };
}

function measure(): void {
  const s = source?.();
  if (!s) {
    last = "";
    return;
  }
  const area = box(s.area.getBoundingClientRect());
  const spots: M.Spot[] = [];
  for (const el of s.area.querySelectorAll<HTMLElement>("[data-spot]")) {
    const key = el.dataset.spot ?? "";
    const sep = key.indexOf(":");
    const kind = key.slice(0, sep);
    const id = key.slice(sep + 1);
    const media = kind === "picture" ? s.media.get(id) : undefined;
    const url = kind === "link" ? (el.dataset.url ?? "") : "";
    if (kind === "picture" ? !media : kind === "link" ? !url : kind !== "name") continue;
    if (el.closest("[hidden]")) continue;
    const holder = el.closest<HTMLElement>("[data-clip]");
    const held = holder ? box(holder.getBoundingClientRect()) : null;
    const rects = kind === "link" ? Array.from(el.getClientRects()) : [el.getBoundingClientRect()];
    for (const rect of rects) {
      let r = clip(box(rect), area);
      if (held) r = clip(r, held);
      const w = r.r - r.l;
      const h = r.b - r.t;
      if (w < 4 || h < 4) continue;
      spots.push(
        new M.Spot({
          key,
          kind,
          x: Math.round(r.l - area.l),
          y: Math.round(r.t - area.t),
          w: Math.round(w),
          h: Math.round(h),
          media: media ?? new C.Media(),
          peerId: kind === "name" ? id.split("@")[0] : "",
          url,
        }),
      );
    }
  }
  const sig = s.key + JSON.stringify(spots.map((p) => [p.key, p.x, p.y, p.w, p.h]));
  if (sig === last) return;
  last = sig;
  setSpots(spots);
}
