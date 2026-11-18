<script lang="ts">
  let { width = 84 }: { width?: number } = $props();
</script>

<span class="mask" role="img" aria-label="Account name hidden" style:width="{width}px">
  <span class="marks"></span>
</span>

<style>
  .mask {
    position: relative;
    flex: none;
    display: inline-block;
    height: 14px;
    overflow: hidden;
    vertical-align: -2px;
    background: rgba(255, 255, 255, 0.07);
    border-radius: 2px;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.07);
  }

  .marks {
    position: absolute;
    inset: 3px 6px;
    background: repeating-linear-gradient(
      90deg,
      rgba(255, 255, 255, 0.6) 0 6px,
      transparent 6px 10px
    );
    filter: blur(2.5px);
  }
</style>
