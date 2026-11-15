const two = (n: number): string => (n < 10 ? "0" + n : String(n));

/** Unix milliseconds → local "12:41". */
export function clock(ms: number): string {
  const d = new Date(ms);
  return two(d.getHours()) + ":" + two(d.getMinutes());
}

/** First visible character of a name, for the square letter avatars. */
export function initial(name: string): string {
  const first = Array.from(name.trim())[0];
  return first ? first.toUpperCase() : "?";
}
