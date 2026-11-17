// What can be clicked: pictures (open at full size) and people's names
// (whisper to them). The overlay window never sees the mouse; Go watches it
// with a mouse hook instead, so it has to know where these are drawn, and
// tells back which one the pointer is over (`app.hover`).
//
// Every such element carries `data-spot` (see `pictureSpot` / `nameSpot`).
// They are measured in CSS px of the 640×360 area and sent to Go whenever
// something may have moved them: once at the next frame, and again when the
// entry animations have ended. Unchanged lists are not sent again.

import * as M from "../../../bindings/mychat/models";
import * as C from "../../../bindings/mychat/internal/chat/models";
import { setSpots } from "../store.svelte";

/** `data-spot` of a picture. */
export const pictureSpot = (m: C.Media): string => `picture:${m.id}`;
/**
 * `data-spot` of a person's name; `at` tells apart the places the same name
 * shows (one link lights up at a time). "" (not clickable) without a Steam id.
 */
export const nameSpot = (peerId: string, at: number): string =>
  /^\d+$/.test(peerId) ? `name:${peerId}@${at}` : "";

interface Source {
  /** The 640×360 area everything is drawn in. */
  area: HTMLElement;
  /** The pictures on screen, by `Media.id`. */
  media: Map<string, C.Media>;
  /** Something that changes whenever Go starts over (mode, viewer). */
  key: string;
  /** How long entry animations run, in ms. */
  settle: number;
}

let source: (() => Source | null) | null = null;
let last = "";
let frame = 0;
let later = 0;

/** Called by the overlay once; `get` returns what to measure now. */
export function spotSource(get: () => Source | null): void {
  source = get;
}

/** Measures again soon (after layout changes, image loads, animations). */
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
    // Hidden: Go forgets the spots by itself; send them again next time.
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
    if ((kind === "picture" && !media) || (kind !== "picture" && kind !== "name")) continue;
    if (el.closest("[hidden]")) continue;
    let r = clip(box(el.getBoundingClientRect()), area);
    const holder = el.closest<HTMLElement>("[data-clip]");
    if (holder) r = clip(r, box(holder.getBoundingClientRect()));
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
      }),
    );
  }
  const sig = s.key + JSON.stringify(spots.map((p) => [p.key, p.x, p.y, p.w, p.h]));
  if (sig === last) return;
  last = sig;
  setSpots(spots);
}
