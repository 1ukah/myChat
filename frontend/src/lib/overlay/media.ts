import type * as C from "../../../bindings/mychat/internal/chat/models";

export interface Size {
  w: number;
  h: number;
}

export const TOAST_FULL: Size = { w: 300, h: 150 };
export const TOAST_THUMB: Size = { w: 120, h: 40 };
export const CHAT_THUMB: Size = { w: 160, h: 90 };

const MIN = 16;

export function pictureOf(entry: C.Entry): C.Media | null {
  const m = entry.media?.[0];
  return m && m.url ? m : null;
}

export function isLocal(url: string): boolean {
  try {
    return new URL(url, location.href).origin === location.origin;
  } catch {
    return false;
  }
}

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
