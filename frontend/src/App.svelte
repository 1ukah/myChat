<script lang="ts">
  // TEMPORARY test page. Replaced by the real UI.
  import {onMount} from 'svelte';
  import {Events} from "@wailsio/runtime";
  import {ChatService} from "../bindings/mychat";

  const view = new URLSearchParams(location.search).get('view') ?? 'settings';
  let mode = $state('hidden');
  let text = $state('');
  let channel = $state('');
  let log: string[] = $state([]);

  onMount(async () => {
    Events.On('mychat:mode', (e: any) => { mode = e.data.mode; });
    Events.On('mychat:composer', (e: any) => { text = e.data.text; channel = e.data.channel.kind + ' ' + e.data.channel.targetName; });
    Events.On('mychat:entry', (e: any) => { log = [...log, e.data.kind + ': ' + e.data.text]; });
    if (view === 'overlay') await ChatService.OverlayReady();
  });
</script>

{#if view === 'overlay'}
  <div class="panel">
    <div>MODE {mode} · {channel}</div>
    {#each log.slice(-6) as l}<div>{l}</div>{/each}
    <div class="input">&gt; {text}<span class="caret">|</span></div>
  </div>
{:else}
  <div style="color:white;padding:20px">settings test</div>
{/if}

<style>
  :global(html), :global(body) { margin:0; background: transparent; overflow:hidden; font-family: sans-serif; }
  .panel { position:fixed; inset:0; background: rgba(20,24,28,0.45); color:#fff; padding:12px; font-size:16px; border-top:2px solid #fff; }
  .input { position:absolute; bottom:12px; left:12px; right:12px; background: rgba(0,0,0,0.35); padding:6px; }
</style>
