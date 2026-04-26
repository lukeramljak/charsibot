<script lang="ts">
  import type { PlushieData } from '../types';

  interface Props {
    plushie: PlushieData | null;
    isAnimating: boolean;
    visible: boolean;
  }

  let { plushie, isAnimating, visible }: Props = $props();
</script>

{#if plushie}
  <img
    class="plushie"
    class:animate-plushie={isAnimating}
    src={plushie.image}
    alt={plushie.name}
    style="opacity: {visible ? 1 : 0}"
  />
{/if}

<style>
  .plushie {
    height: 150px;
    display: block;
    opacity: 0;
    transform-style: preserve-3d;
    will-change: transform;
    transition: opacity 0.3s;
  }

  /* ANIMATIONS - PLUSHIE */
  @keyframes plushieReveal {
    0% {
      transform: translateY(100px) scale(0.5);
      opacity: 0;
    }
    25% {
      transform: translateY(-180%) scale(2);
      opacity: 1;
    }
    100% {
      transform: translateY(0) scale(3) translateZ(800px);
      opacity: 1;
    }
  }

  @keyframes plushieSwing {
    0%,
    100% {
      transform: translateY(0) scale(3) translateZ(800px) rotateZ(0);
    }
    20% {
      transform: translateY(0) scale(3) translateZ(800px) rotateZ(4deg);
    }
    40% {
      transform: translateY(0) scale(3) translateZ(800px) rotateZ(-4deg);
    }
    60% {
      transform: translateY(0) scale(3) translateZ(800px) rotateZ(3deg);
    }
    80% {
      transform: translateY(0) scale(3) translateZ(800px) rotateZ(-2deg);
    }
  }

  @keyframes plushieExit {
    0% {
      transform: translateY(0) translateZ(800px) scale(3);
      opacity: 1;
    }
    80% {
      transform: translateY(-180%) scale(2);
      opacity: 1;
    }
    100% {
      transform: translateY(100px) scale(0);
      opacity: 0.5;
    }
  }

  .plushie:global(.animate-plushie) {
    animation:
      1.4s cubic-bezier(0.5, -0.2, 0.3, 1.3) 1.2s forwards plushieReveal,
      1.2s ease-in-out 3s 2 alternate plushieSwing,
      1s cubic-bezier(0.5, -0.2, 0.3, 1.3) 5s forwards plushieExit;
  }
</style>
