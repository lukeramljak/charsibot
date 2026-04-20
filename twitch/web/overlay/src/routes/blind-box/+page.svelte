<script lang="ts">
  import { onMount } from 'svelte';
  import BlindBox from '$lib/overlays/blind-box/components/BlindBox.svelte';
  import type { BlindBoxOverlayConfig } from '$lib/overlays/blind-box/types';
  import { env } from '$env/dynamic/public';

  let configs = $state<BlindBoxOverlayConfig[]>([]);

  onMount(async () => {
    const res = await fetch(`${env.PUBLIC_SERVER_BASE_URL}/api/blindbox`);
    if (res.ok) {
      configs = await res.json();
    }
  });
</script>

<BlindBox {configs} />
