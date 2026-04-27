<script lang="ts">
  import BlindBox from '$lib/overlays/blind-box/components/BlindBox.svelte';
  import DisconnectedBanner from '$lib/overlays/components/DisconnectedBanner.svelte';
  import { charsibot } from '$lib/charsibot.svelte';
</script>

{#if !charsibot.isConnected}
  <DisconnectedBanner
    title="Blind Box Overlay is Sleeping"
    subtitle="Please wait before redeeming! ♡"
  />
{/if}

<BlindBox />
