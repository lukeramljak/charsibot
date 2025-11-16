<script lang="ts">
  import { charsibotWebSocket } from '$lib/charsibot.svelte';
  import type { BlindBoxRedemptionEvent, CollectionDisplayEvent } from '$lib/types';
  import type { BlindBoxOverlayConfig, PlushieData } from '../types';
  import Box3D from './Box3D.svelte';
  import PlushieReveal from './PlushieReveal.svelte';
  import DisplayBanner from './DisplayBanner.svelte';
  import CollectionDisplay from './CollectionDisplay.svelte';
  import BackgroundEffects from './BackgroundEffects.svelte';

  interface Props {
    config: BlindBoxOverlayConfig;
  }

  let { config }: Props = $props();

  let isAnimating = $state(false);
  let isShowingCollection = $state(false);
  let showEffects = $state(false);
  let boxVisible = $state(false);
  let plushieVisible = $state(false);
  let displayTextVisible = $state(false);

  let currentPlushie = $state<PlushieData | null>(null);
  let displayMessage = $state('');
  let userCollection = $state<string[]>([]);

  let audioElement: HTMLAudioElement;

  let activeTimers: number[] = [];
  let lastProcessedMessage: unknown = null;

  const clearTimers = () => {
    activeTimers.forEach((timer) => clearTimeout(timer));
    activeTimers = [];
  };

  const playAudio = () => {
    if (audioElement) {
      audioElement.currentTime = 0;
      audioElement.play().catch((error) => {
        console.warn('Audio playback failed:', error);
      });
    }
  };

  const playRevealAnimation = async (): Promise<void> => {
    return new Promise((resolve) => {
      clearTimers();

      // Initial state - box appears
      boxVisible = true;
      showEffects = true;
      isAnimating = true;
      plushieVisible = false;
      displayTextVisible = false;

      // Show plushie after box opens (1200ms)
      activeTimers.push(
        setTimeout(() => {
          plushieVisible = true;
        }, 1200) as unknown as number,
      );

      // Show display text (1500ms)
      activeTimers.push(
        setTimeout(() => {
          displayTextVisible = true;
        }, 1500) as unknown as number,
      );

      // Hide effects and display text (5000ms)
      activeTimers.push(
        setTimeout(() => {
          showEffects = false;
          displayTextVisible = false;
        }, 5000) as unknown as number,
      );

      // Hide everything and complete (6500ms)
      activeTimers.push(
        setTimeout(() => {
          boxVisible = false;
          plushieVisible = false;
          isAnimating = false;
          clearTimers();
          resolve();
        }, 6500) as unknown as number,
      );
    });
  };

  const playCollectionAnimation = async (): Promise<void> => {
    return new Promise((resolve) => {
      clearTimers();

      // Show collection
      boxVisible = true;
      showEffects = true;
      isShowingCollection = true;
      displayTextVisible = true;
      plushieVisible = false;

      // Start hiding effects and text (5000ms)
      activeTimers.push(
        setTimeout(() => {
          showEffects = false;
          displayTextVisible = false;
          isShowingCollection = false;
          boxVisible = false;
        }, 5000) as unknown as number,
      );

      // Complete (5500ms)
      activeTimers.push(
        setTimeout(() => {
          clearTimers();
          resolve();
        }, 5500) as unknown as number,
      );
    });
  };

  const forceCancel = () => {
    clearTimers();
    isShowingCollection = false;
    boxVisible = false;
    showEffects = false;
    displayTextVisible = false;
    plushieVisible = false;
    isAnimating = false;
  };

  const playReveal = async (username: string, plushie: PlushieData, isDuplicate: boolean) => {
    currentPlushie = plushie;
    displayMessage = `${username} just got <strong>${plushie.name}</strong>${
      isDuplicate ? ' (duplicate)' : ''
    }`;
    playAudio();
    await playRevealAnimation();
  };

  const playCollection = async (username: string) => {
    displayMessage = `${username}'s ${config.collectionName}`;
    currentPlushie = null;
    await playCollectionAnimation();
  };

  // Listen for WebSocket messages
  $effect(() => {
    if (audioElement) {
      audioElement.volume = config.audioVolume / 100;
    }

    const lastMsg = charsibotWebSocket.lastMessage;
    if (!lastMsg) return;

    if (lastMsg === lastProcessedMessage) return;

    if (lastMsg.type === 'blindbox_redemption') {
      const event = lastMsg as BlindBoxRedemptionEvent;

      // Only handle events for this collection type
      if (event.data.collectionType !== config.collectionType) return;

      // Don't interrupt another animation
      if (isAnimating) return;

      lastProcessedMessage = lastMsg;

      // Force cancel any collection display
      forceCancel();

      // Wait for cleanup
      setTimeout(async () => {
        const plushie: PlushieData = {
          key: event.data.plushie.key,
          name: event.data.plushie.name,
          image: config.plushies.find((p) => p.key === event.data.plushie.key)?.image || '',
        };

        userCollection = event.data.collection;
        await playReveal(event.data.username, plushie, !event.data.isNew);
      }, 500);
    }

    // Handle collection display events
    if (lastMsg.type === 'collection_display') {
      const event = lastMsg as CollectionDisplayEvent;

      // Only handle events for this collection type
      if (event.data.collectionType !== config.collectionType) return;

      // Don't start if already animating or showing collection
      if (isAnimating || isShowingCollection) return;

      lastProcessedMessage = lastMsg;

      userCollection = event.data.collection;
      playCollection(event.data.username);
    }
  });
</script>

<audio bind:this={audioElement} src={config.revealSound} preload="auto"></audio>

<div class="scene">
  <BackgroundEffects show={showEffects} />

  <div class="content-wrapper" class:with-plushie={currentPlushie !== null}>
    <Box3D
      boxFrontFace={config.boxFrontFace}
      boxSideFace={config.boxSideFace}
      {isAnimating}
      visible={boxVisible}
    >
      <div class="plushie-container">
        <PlushieReveal plushie={currentPlushie} {isAnimating} visible={plushieVisible} />
      </div>
    </Box3D>

    <div class="text-collection-container" class:visible={displayTextVisible}>
      <DisplayBanner
        message={displayMessage}
        displayColor={config.displayColor}
        textColor={config.textColor}
        visible={displayTextVisible}
      />

      <CollectionDisplay
        plushies={config.plushies}
        {userCollection}
        emptyPlushieImage={config.emptyPlushieImage}
        visible={isAnimating || isShowingCollection}
      />
    </div>
  </div>
</div>

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
    transition:
      opacity 0.6s,
      transform 0.6s;
    pointer-events: none;
  }

  .text-collection-container:global(.visible) {
    opacity: 1;
    transform: translateY(0);
    pointer-events: auto;
  }
</style>
