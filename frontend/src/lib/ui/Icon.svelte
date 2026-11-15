<script lang="ts" module>
  // Inline line icons on a 16 px grid, drawn square like the Panorama UI.
  const paths = {
    close: "M3.5 3.5l9 9 M12.5 3.5l-9 9",
    chevron: "M4 6l4 4 4-4",
    back: "M7 3.5l-4.5 4.5 4.5 4.5 M3 8h10.5",
    lock: "M3.5 7.5h9v6h-9z M5.5 7.5v-2.5a2.5 2.5 0 0 1 5 0v2.5",
    phone: "M4.5 1.5h7v13h-7z M7 12.5h2",
    mail: "M1.5 3.5h13v9h-13z M1.5 4l6.5 5 6.5-5",
    minimise: "M3.5 8.5h9",
    reset: "M14.5 2.5v4h-4 M13 10a5.5 5.5 0 1 1-1.3-5.7l2.8 2.2",
    logout: "M6 2.5h-3.5v11h3.5 M10 5l3 3-3 3 M13 8h-7",
    power: "M8 1.5v6 M4.5 4a5.5 5.5 0 1 0 7 0",
    eye: "M1.5 8s2.5-4.5 6.5-4.5 6.5 4.5 6.5 4.5-2.5 4.5-6.5 4.5-6.5-4.5-6.5-4.5z M10 8a2 2 0 1 1-4 0 2 2 0 0 1 4 0z",
    user: "M8 8a2.75 2.75 0 1 0 0-5.5 2.75 2.75 0 0 0 0 5.5z M2.5 14c.5-2.8 2.7-4 5.5-4s5 1.2 5.5 4",
    retry: "M1.5 2.5v4h4 M3 10a5.5 5.5 0 1 0 1.3-5.7l-2.8 2.2",
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
