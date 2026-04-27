<script lang="ts">
  import type { BlindBoxOverlayConfig, PlushieData } from '$lib/types';
  import Box3D from './Box3D.svelte';
  import PlushieReveal from './PlushieReveal.svelte';
  import DisplayBanner from './DisplayBanner.svelte';
  import CollectionDisplay from './CollectionDisplay.svelte';
  import BackgroundEffects from './BackgroundEffects.svelte';
  import { BlindBoxQueue } from '../queue.svelte';
  import { charsibot } from '$lib/charsibot.svelte';
  import { onMount } from 'svelte';

  type AnimationMode = 'idle' | 'reveal' | 'collection';

  interface CurrentItem {
    plushie: PlushieData | null;
    config: BlindBoxOverlayConfig;
    message: string;
    collection: string[];
  }

  let mode = $state<AnimationMode>('idle');
  let animationKey = $state(0);
  let currentItem = $state<CurrentItem | null>(null);
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
        currentItem = null;
        resolve();
      }, 6500);
    });
  }

  const queue = new BlindBoxQueue({
    onRedemption: async (item) => {
      currentItem = {
        config: item.config,
        collection: item.collection,
        plushie: item.plushie,
        message: `${item.username} just got <strong>${item.plushie.name}</strong>${
          !item.isNew ? ' (duplicate)' : ''
        }`,
      };
      await playAudio(item.config.revealSound);
      await playAnimation('reveal');
    },
    onDisplay: async (item) => {
      currentItem = {
        config: item.config,
        collection: item.collection,
        plushie: null,
        message: `${item.username}'s ${item.config.name}`,
      };
      await playAnimation('collection');
    },
  });

  onMount(() => {
    charsibot.connect();
    const unsubscribe = charsibot.onMessage((message) => {
      if (message.type !== 'blindbox_display' && message.type !== 'blindbox_redemption') return;
      queue.add(message);
    });

    return () => {
      unsubscribe();
      queue.clear();
      audioElement?.pause();
    };
  });
</script>

<audio bind:this={audioElement}></audio>

{#if currentItem && charsibot.isConnected}
  {#key animationKey}
    <div class="scene" class:reveal={mode === 'reveal'} class:collection={mode === 'collection'}>
      <BackgroundEffects show={mode !== 'idle'} />

      <div class="content-wrapper" class:with-plushie={currentItem.plushie !== null}>
        <Box3D
          boxFrontFace={currentItem.config.boxFrontFace}
          boxSideFace={currentItem.config.boxSideFace}
          isAnimating={mode === 'reveal'}
          visible={mode !== 'idle'}
        >
          <div class="plushie-container">
            <PlushieReveal
              plushie={currentItem.plushie}
              isAnimating={mode === 'reveal'}
              visible={mode !== 'idle'}
            />
          </div>
        </Box3D>

        <div class="text-collection-container">
          <DisplayBanner
            message={currentItem.message}
            displayColor={currentItem.config.displayColor}
            textColor={currentItem.config.textColor}
            visible={mode !== 'idle'}
          />

          <CollectionDisplay
            plushies={currentItem.config.plushies}
            collection={currentItem.collection}
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
