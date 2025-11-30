<script lang="ts">
  import { type BlindBoxRedemptionEvent, type CollectionDisplayEvent } from '$lib/types';
  import type { BlindBoxOverlayConfig, PlushieData } from '../types';
  import Box3D from './Box3D.svelte';
  import PlushieReveal from './PlushieReveal.svelte';
  import DisplayBanner from './DisplayBanner.svelte';
  import CollectionDisplay from './CollectionDisplay.svelte';
  import BackgroundEffects from './BackgroundEffects.svelte';
  import {
    BlindBoxQueue,
    type PlushieDisplayQueueItem,
    type PlushieRedemptionQueueItem,
  } from '../queue.svelte';
  import { CharsibotWebSocket } from '$lib/charsibot.svelte';

  interface Props {
    configs: BlindBoxOverlayConfig[];
  }

  let { configs }: Props = $props();

  const charsibotWebSocket = new CharsibotWebSocket();
  charsibotWebSocket.connect();

  type AnimationMode = 'idle' | 'reveal' | 'collection';

  let mode = $state<AnimationMode>('idle');
  let animationKey = $state(0);
  let currentPlushie = $state<PlushieData | null>(null);
  let currentConfig = $state<BlindBoxOverlayConfig>();
  let displayMessage = $state('');
  let userCollection = $state<string[]>([]);
  let audioElement: HTMLAudioElement | undefined = $state();
  let lastProcessedMessage: unknown = null;

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

  async function playReveal(item: PlushieRedemptionQueueItem) {
    currentConfig = item.config;
    userCollection = item.collection;
    currentPlushie = item.plushie;
    displayMessage = `${item.username} just got <strong>${item.plushie.name}</strong>${
      item.isDuplicate ? ' (duplicate)' : ''
    }`;
    await playAudio(item.config.revealSound);
    await playAnimation('reveal');
  }

  async function playCollection(item: PlushieDisplayQueueItem) {
    currentConfig = item.config;
    userCollection = item.collection;
    currentPlushie = null;
    displayMessage = `${item.username}'s ${item.config.collectionName}`;
    await playAnimation('collection');
  }

  const handlers = {
    onRedemption: async (item: PlushieRedemptionQueueItem) => {
      await playReveal(item);
    },
    onDisplay: async (item: PlushieDisplayQueueItem) => {
      await playCollection(item);
    },
  };

  const queue = new BlindBoxQueue(handlers);

  function handleRedemptionEvent(event: BlindBoxRedemptionEvent) {
    const config = configs.find((c) => c.collectionType === event.data.collectionType);
    if (!config) {
      console.warn('Config not found for collection type:', event.data.collectionType);
      return;
    }

    const plushie = config.plushies.find((p) => p.key === event.data.plushie);
    if (!plushie) {
      console.warn('Plushie not found for reward key:', event.data.plushie);
      return;
    }

    queue.addRedemption({
      type: 'redemption',
      username: event.data.username,
      plushie,
      isDuplicate: !event.data.isNew,
      collection: event.data.collection,
      config: config,
    });

    queue.processNext();
  }

  function handleDisplayEvent(event: CollectionDisplayEvent) {
    const config = configs.find((c) => c.collectionType === event.data.collectionType);
    if (!config) {
      console.warn('Config not found for collection type:', event.data.collectionType);
      return;
    }

    queue.addDisplay({
      type: 'display',
      username: event.data.username,
      collection: event.data.collection,
      config: config,
    });

    queue.processNext();
  }

  $effect(() => {
    const lastMsg = charsibotWebSocket.lastMessage;
    if (!lastMsg || lastMsg === lastProcessedMessage) return;

    lastProcessedMessage = lastMsg;

    if (lastMsg.type === 'blindbox_redemption') {
      handleRedemptionEvent(lastMsg as BlindBoxRedemptionEvent);
    } else if (lastMsg.type === 'collection_display') {
      handleDisplayEvent(lastMsg as CollectionDisplayEvent);
    }
  });

  $effect(() => {
    return () => {
      queue.clear();
      audioElement?.pause();
    };
  });
</script>

<audio bind:this={audioElement}></audio>

{#if currentConfig}
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
            {userCollection}
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
      'Inter',
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
