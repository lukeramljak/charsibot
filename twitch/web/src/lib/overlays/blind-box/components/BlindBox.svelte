<script lang="ts">
  import type {
    OverlayEvent,
    BlindBoxRedemptionEvent,
    CollectionDisplayEvent,
    BlindBoxOverlayConfig,
    PlushieData,
  } from '$lib/types';
  import Box3D from './Box3D.svelte';
  import PlushieReveal from './PlushieReveal.svelte';
  import DisplayBanner from './DisplayBanner.svelte';
  import CollectionDisplay from './CollectionDisplay.svelte';
  import BackgroundEffects from './BackgroundEffects.svelte';
  import { BlindBoxQueue } from '../queue.svelte';
  import { charsibot } from '$lib/charsibot.svelte';
  import { onMount } from 'svelte';
  import DisconnectedBanner from '$lib/overlays/components/DisconnectedBanner.svelte';

  type AnimationMode = 'idle' | 'reveal' | 'collection';

  let mode = $state<AnimationMode>('idle');
  let animationKey = $state(0);
  let currentPlushie = $state<PlushieData | null>(null);
  let currentConfig = $state<BlindBoxOverlayConfig>();
  let displayMessage = $state('');
  let collection = $state<string[]>([]);
  let audioElement: HTMLAudioElement | undefined = $state();

  async function playAudio(sound: string) {
    if (!audioElement) return;

    try {
      audioElement.src = sound;
      if (!audioElement.paused) {
        audioElement.pause();
      }
      audioElement.currentTime = 0;
      await audioElement.play();
    } catch (error) {
      console.warn('Audio playback failed:', error);
    }
  }

  async function playAnimation(animationMode: AnimationMode): Promise<void> {
    return new Promise((resolve) => {
      animationKey++;
      mode = animationMode;

      setTimeout(() => {
        mode = 'idle';
        currentConfig = undefined;
        currentPlushie = null;
        resolve();
      }, 6500);
    });
  }

  async function playReveal(item: BlindBoxRedemptionEvent) {
    currentConfig = item.config;
    collection = item.collection;
    currentPlushie = item.plushie;
    displayMessage = `${item.username} just got <strong>${item.plushie.name}</strong>${
      !item.isNew ? ' (duplicate)' : ''
    }`;
    await playAudio(item.config.revealSound);
    await playAnimation('reveal');
  }

  async function playCollection(item: CollectionDisplayEvent) {
    currentConfig = item.config;
    collection = item.collection;
    currentPlushie = null;
    displayMessage = `${item.username}'s ${item.config.name}`;
    await playAnimation('collection');
  }

  const handlers = {
    onRedemption: async (item: BlindBoxRedemptionEvent) => {
      await playReveal(item);
    },
    onDisplay: async (item: CollectionDisplayEvent) => {
      await playCollection(item);
    },
  };

  const queue = new BlindBoxQueue(handlers);

  function handleMessage(message: OverlayEvent) {
    if (message.type !== 'blindbox_display' && message.type !== 'blindbox_redemption') {
      return;
    }

    queue.add(message);
    queue.processNext();
  }

  onMount(() => {
    charsibot.connect();
    const unsubscribe = charsibot.onMessage(handleMessage);

    return () => {
      unsubscribe();
      queue.clear();
      audioElement?.pause();
    };
  });
</script>

<audio bind:this={audioElement}></audio>

{#if !charsibot.isConnected}
  <DisconnectedBanner
    title="Blind Box Overlay is Sleeping"
    subtitle="Please wait before redeeming! ♡"
  />
{/if}

{#if currentConfig && charsibot.isConnected}
  {#key animationKey}
    <div class="scene" class:reveal={mode === 'reveal'} class:collection={mode === 'collection'}>
      <BackgroundEffects show={mode !== 'idle'} />

      <div class="content-wrapper" class:with-plushie={currentPlushie !== null}>
        <Box3D
          boxFrontFace={currentConfig.boxFrontFace}
          boxSideFace={currentConfig.boxSideFace}
          isAnimating={mode === 'reveal'}
          visible={mode !== 'idle'}
        >
          <div class="plushie-container">
            <PlushieReveal
              plushie={currentPlushie}
              isAnimating={mode === 'reveal'}
              visible={mode !== 'idle'}
            />
          </div>
        </Box3D>

        <div class="text-collection-container">
          <DisplayBanner
            message={displayMessage}
            displayColor={currentConfig.displayColor}
            textColor={currentConfig.textColor}
            visible={mode !== 'idle'}
          />

          <CollectionDisplay
            plushies={currentConfig.plushies}
            {collection}
            visible={mode !== 'idle'}
          />
        </div>
      </div>
    </div>
  {/key}
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
    width: 100vw;
    height: 100vh;
    font-family:
      'Nunito',
      -apple-system,
      BlinkMacSystemFont,
      'Segoe UI',
      sans-serif;
  }

  .scene {
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0;
    padding: 0;
    width: 100vw;
    height: 100vh;
  }

  .content-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 60px;
    transform-style: preserve-3d;
    transform: translateY(150px);
  }

  .plushie-container {
    position: absolute;
    left: 50%;
    top: 0;
    transform: translateX(-50%) rotateY(30deg) rotateX(20deg);
    display: flex;
    flex-direction: column;
    align-items: center;
    transform-style: preserve-3d;
    pointer-events: none;
    z-index: 999;
    will-change: transform;
  }

  .text-collection-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    opacity: 0;
    transform: translateY(20px);
    pointer-events: none;
  }

  .scene.reveal .text-collection-container {
    animation:
      fadeInText 0.6s ease-out 1.5s forwards,
      fadeOutText 0.6s ease-out 6s forwards;
  }

  .scene.collection .text-collection-container {
    animation:
      fadeInText 0.6s ease-out 0.05s forwards,
      fadeOutText 0.6s ease-out 6s forwards;
  }

  @keyframes fadeInText {
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes fadeOutText {
    to {
      opacity: 0;
      transform: translateY(20px);
    }
  }
</style>
