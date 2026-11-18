export interface Part {
  text: string;
  url?: string;
}

const LINK = /https?:\/\/[^\s[\]<>"']+/g;
const TRAIL = /[.,!?)]+$/;

function isWeb(url: string): boolean {
  try {
    const u = new URL(url);
    return (u.protocol === "https:" || u.protocol === "http:") && u.hostname !== "";
  } catch {
    return false;
  }
}

export function splitLinks(text: string): Part[] {
  const parts: Part[] = [];
  let at = 0;
  for (const m of text.matchAll(LINK)) {
    const url = m[0].replace(TRAIL, "");
    if (!isWeb(url)) continue;
    const start = m.index ?? 0;
    if (start > at) parts.push({ text: text.slice(at, start) });
    parts.push({ text: url, url });
    at = start + url.length;
  }
  if (at < text.length) parts.push({ text: text.slice(at) });
  return parts;
}
