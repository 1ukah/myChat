<script lang="ts" module>
  const paths = {
    close: "M3.5 3.5l9 9 M12.5 3.5l-9 9",
    chevron: "M4 6l4 4 4-4",
    open: "M9.5 2.5h4v4 M13.5 2.5l-6 6 M11.5 9.5v4h-9v-9h4",
    swap: "M2.5 5.5h10 M9.5 2.5l3 3-3 3 M13.5 10.5h-10 M6.5 7.5l-3 3 3 3",
    alert: "M8 2l6.5 11.5h-13z M8 6.5v3.5 M8 12v.5",
    pc: "M2.5 3.5h11v7h-11z M6 13.5h4 M8 10.5v3",
    qr: "M2.5 2.5h4v4h-4z M9.5 2.5h4v4h-4z M2.5 9.5h4v4h-4z M9.5 9.5h1.5v1.5 M13.5 9.5v4h-4",
    minimise: "M3.5 8.5h9",
    reset: "M14.5 2.5v4h-4 M13 10a5.5 5.5 0 1 1-1.3-5.7l2.8 2.2",
    logout: "M6 2.5h-3.5v11h3.5 M10 5l3 3-3 3 M13 8h-7",
    power: "M8 1.5v6 M4.5 4a5.5 5.5 0 1 0 7 0",
    eye: "M1.5 8s2.5-4.5 6.5-4.5 6.5 4.5 6.5 4.5-2.5 4.5-6.5 4.5-6.5-4.5-6.5-4.5z M10 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0z",
    "eye-off":
      "M1.5 8s2.5-4.5 6.5-4.5 6.5 4.5 6.5 4.5-2.5 4.5-6.5 4.5-6.5-4.5-6.5-4.5z M10 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0z M2.5 13.5l11-11",
    user: "M8 8a2.75 2.75 0 1 0 0-5.5 2.75 2.75 0 0 0 0 5.5z M2.5 14c.5-2.8 2.7-4 5.5-4s5 1.2 5.5 4",
    retry: "M1.5 2.5v4h4 M3 10a5.5 5.5 0 1 0 1.3-5.7l-2.8 2.2",
    search: "M6.75 11a4.25 4.25 0 1 0 0-8.5 4.25 4.25 0 0 0 0 8.5z M10 10l4 4",
    history: "M1.5 3v3.5h3.5 M2.6 10.2a5.75 5.75 0 1 0 .3-5.2l-1.4 1.5 M8 5v3.2l2.2 1.4",
    check: "M3.5 8.5l3 3 6-7",
    lock: "M3.5 7.5h9v6.5h-9z M5.5 7.5v-2a2.5 2.5 0 0 1 5 0v2",
    phone: "M4.5 1.5h7v13h-7z M7 12.5h2",
    mail: "M1.5 3.5h13v9h-13z M1.5 3.5l6.5 5.5 6.5-5.5",
    back: "M10 3l-5 5 5 5",
  } as const;

  export type IconName = keyof typeof paths;
</script>

<script lang="ts">
  let { name, size = 16 }: { name: IconName; size?: number } = $props();
</script>

<svg
  class="icon"
  width={size}
  height={size}
  viewBox="0 0 16 16"
  fill="none"
  stroke="currentColor"
  stroke-width="1.5"
  stroke-linecap="square"
  stroke-linejoin="miter"
  aria-hidden="true"
>
  <path d={paths[name]} />
</svg>

<style>
  .icon {
    display: block;
    flex: none;
  }
</style>
