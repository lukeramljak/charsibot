<script lang="ts">
  import BlindBox from '$lib/overlays/blind-box/components/BlindBox.svelte';
  import { coobubuConfig } from '$lib/overlays/blind-box/configs/coobubu';
  import { charsibotWebSocket } from '$lib/charsibot.svelte';
  import { onMount } from 'svelte';

  onMount(() => {
    charsibotWebSocket.connect();
  });
</script>

<BlindBox config={coobubuConfig} />
