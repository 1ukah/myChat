const two = (n: number): string => (n < 10 ? "0" + n : String(n));

/** Unix milliseconds → local "12:41". */
export function clock(ms: number): string {
  const d = new Date(ms);
  return two(d.getHours()) + ":" + two(d.getMinutes());
}

/** The five CS2 player colours (chat dots, avatar borders). */
const PLAYER_COLORS = ["#5cb3ea", "#18b28c", "#eedc3c", "#f08a2e", "#e667b8"] as const;

/** A steady colour for a person, picked from their id (or name). */
export function playerColor(key: string): string {
  let h = 0;
  for (let i = 0; i < key.length; i++) h = (h * 31 + key.charCodeAt(i)) >>> 0;
  return PLAYER_COLORS[h % PLAYER_COLORS.length];
}

/** First visible character of a name, for the square letter avatars. */
export function initial(name: string): string {
  const first = Array.from(name.trim())[0];
  return first ? first.toUpperCase() : "?";
}
