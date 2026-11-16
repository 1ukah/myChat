// Sizes for the pictures in messages (`Entry.media`), shared by the pop-ups
// and the chat box so that what is reserved is what is drawn.

import type * as C from "../../../bindings/mychat/internal/chat/models";

export interface Size {
  w: number;
  h: number;
}

/** The newest picture in the pop-ups. */
export const TOAST_FULL: Size = { w: 300, h: 150 };
/** Pictures in older pop-ups. */
export const TOAST_THUMB: Size = { w: 120, h: 40 };
/** Pictures in the chat box. */
export const CHAT_THUMB: Size = { w: 160, h: 90 };

/** Nothing is drawn smaller than this, however thin the picture is. */
const MIN = 16;

/** The picture of a message, if it has one (at most one for now). */
export function pictureOf(entry: C.Entry): C.Media | null {
  const m = entry.media?.[0];
  return m && m.url ? m : null;
}

/**
 * Pictures are only ever loaded from myChat itself (`/media/<id>`): a
 * relative address, or an absolute one on this page's own origin.
 */
export function isLocal(url: string): boolean {
  try {
    return new URL(url, location.href).origin === location.origin;
  } catch {
    return false;
  }
}

/**
 * The box for a `width` × `height` picture inside `maxW` × `maxH`: same
 * aspect ratio, never enlarged. With the size unknown (0) it is a 4:3 box
 * of full height, until the loaded image tells.
 */
export function fitSize(width: number, height: number, maxW: number, maxH: number): Size {
  if (!(width > 0 && height > 0)) {
    return { w: Math.min(maxW, Math.round((maxH * 4) / 3)), h: maxH };
  }
  const k = Math.min(1, maxW / width, maxH / height);
  return {
    w: Math.min(maxW, Math.max(MIN, Math.round(width * k))),
    h: Math.min(maxH, Math.max(MIN, Math.round(height * k))),
  };
}
