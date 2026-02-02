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
  import { charsibot } from '$lib/charsibot.svelte';
  import { onMount } from 'svelte';

  interface Props {
    configs: BlindBoxOverlayConfig[];
  }

  let { configs }: Props = $props();

  onMount(() => {
    charsibot.connect();
  });

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
    const lastMsg = charsibot.lastMessage;
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

{#if !charsibot.isConnected}
  <div class="not-connected-overlay">
    <div class="not-connected-card">
      <div class="sparkles">✨</div>
      <div class="sleeping-icon">
        <svg viewBox="0 0 36 36" xmlns="http://www.w3.org/2000/svg" role="img"
          ><path
            fill="currentColor"
            d="M33 19c1.187 0 2 .786 2 2c0 1.073-.983 2-2 2H22c-1.496 0-2-.813-2-2c0-.565.632-1.492 1-2l8-12h-7c-1.128 0-2-.843-2-2c0-1.073.929-2 2-2h11c1.639 0 2 1.012 2 2c0 .621-.635 1.519-1 2l-8 12h7zm-16 5c.633 0 1 .353 1 1c0 .573-.458 1-1 1h-6c-.798 0-1-.367-1-1c0-.301.337-.729.533-1L15 18h-4c-.602 0-1-.384-1-1c0-.573.428-1 1-1h6c.874 0 1 .473 1 1c0 .331-.338.877-.533 1.133L13 24h4zm-9 7c.633 0 1 .353 1 1c0 .573-.458 1-1 1H2c-.798 0-1-.367-1-1c0-.301.337-.729.533-1L6 25H2c-.602 0-1-.384-1-1c0-.572.428-1 1-1h6c.874 0 1 .473 1 1c0 .331-.338.877-.533 1.133L4 31h4z"
          ></path></svg
        >
      </div>
      <div class="not-connected-title">Blind Box Overlay is Sleeping</div>
      <div class="not-connected-warning">Please wait before redeeming! ♡</div>
      <div class="sparkles sparkles-bottom">✨</div>
    </div>
  </div>
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

  .not-connected-overlay {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  }

  .not-connected-card {
    background: linear-gradient(
      135deg,
      rgba(255, 182, 193, 0.95) 0%,
      rgba(221, 160, 221, 0.95) 100%
    );
    border-radius: 24px;
    padding: 32px 56px;
    box-shadow:
      0 8px 32px rgba(221, 160, 221, 0.4),
      0 0 0 3px rgba(255, 255, 255, 0.6);
    text-align: center;
    animation: float-gentle 3s ease-in-out infinite;
    position: relative;
  }

  .sparkles {
    font-size: 24px;
    position: absolute;
    top: 16px;
    left: 20px;
    animation: twinkle 1.5s ease-in-out infinite;
  }

  .sparkles-bottom {
    top: auto;
    bottom: 16px;
    left: auto;
    right: 20px;
    animation-delay: 0.75s;
  }

  .sleeping-icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 12px;
    color: #c084c0;
    filter: drop-shadow(0 2px 8px rgba(138, 43, 226, 0.4));
    animation: gentle-bob 2s ease-in-out infinite;
  }

  .sleeping-icon svg {
    width: 100%;
    height: 100%;
  }

  .not-connected-title {
    font-size: 26px;
    font-weight: 700;
    color: #ffffff;
    margin: 0 0 10px 0;
    text-shadow: 0 2px 8px rgba(138, 43, 226, 0.4);
    letter-spacing: 0.5px;
  }

  .not-connected-warning {
    font-size: 17px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.98);
    margin: 0;
    text-shadow: 0 1px 4px rgba(138, 43, 226, 0.3);
  }

  @keyframes float-gentle {
    0%,
    100% {
      transform: translateY(0px);
    }
    50% {
      transform: translateY(-8px);
    }
  }

  @keyframes gentle-bob {
    0%,
    100% {
      transform: translateY(0px) scale(1);
    }
    50% {
      transform: translateY(-6px) scale(1.05);
    }
  }

  @keyframes twinkle {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.5;
      transform: scale(1.2);
    }
  }
</style>
