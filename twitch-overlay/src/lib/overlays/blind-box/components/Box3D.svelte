<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    boxFrontFace: string;
    boxSideFace: string;
    isAnimating: boolean;
    visible: boolean;
    children?: Snippet;
  }

  let { boxFrontFace, boxSideFace, isAnimating, visible, children }: Props = $props();
</script>

<div class="box" class:animate-box={isAnimating} style="opacity: {visible ? 1 : 0}">
  <div class="box-shadow"></div>
  <div
    class="face front"
    style="background-image: url('{boxFrontFace}'); background-size: cover; background-position: center;"
  ></div>
  <div class="face back"></div>
  <div class="face left"></div>
  <div
    class="face right"
    style="background-image: url('{boxSideFace}'); background-size: cover; background-position: center;"
  ></div>
  <div class="flap front-flap" class:animate-front-flap={isAnimating}></div>
  <div class="flap back-flap" class:animate-back-flap={isAnimating}></div>

  {@render children?.()}
</div>

<style>
  :root {
    --bw: 220px;
    --bh: 260px;
    --bd: 220px;
  }

  .box {
    position: relative;
    width: var(--bw);
    height: var(--bh);
    transform-style: preserve-3d;
    transform: rotateX(-20deg) rotateY(-30deg);
    will-change: transform;
    opacity: 0;
    pointer-events: none;
  }

  .face {
    position: absolute;
    width: var(--bw);
    height: var(--bh);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    border-radius: 4px;
    background: center/cover #ededed;
    pointer-events: none;
  }

  .face::after {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.2), rgba(0, 0, 0, 0.1));
  }

  /* Face positioning */
  .front {
    transform: translateZ(calc(var(--bd) / 2));
  }

  .back {
    transform: rotateY(180deg) translateZ(calc(var(--bd) / 2));
    background-color: #e1e1e5;
  }

  .left {
    width: var(--bd);
    transform: rotateY(-90deg) translateZ(calc(var(--bw) / 2));
  }

  .right {
    width: var(--bd);
    transform: rotateY(90deg) translateZ(calc(var(--bw) / 2));
  }

  /* Box flaps */
  .flap {
    position: absolute;
    width: var(--bw);
    height: calc(var(--bd) / 2);
    transform-style: preserve-3d;
    background-size: cover;
    background-position: center;
    border-radius: 2px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .front-flap {
    top: 0;
    left: 0;
    transform-origin: top center;
    background-color: #ededed;
    transform: translateZ(calc(var(--bd) / 2)) rotateX(270deg);
  }

  .back-flap {
    background: #ededed;
    top: 0;
    left: 0;
    transform-origin: top center;
    transform: rotateY(180deg) translateZ(calc(var(--bd) / 2)) rotateX(-90deg);
  }

  .box-shadow {
    width: 225px;
    height: 225px;
    background: rgba(0, 0, 0, 0.2);
    position: absolute;
    transform: rotateX(90deg) translateZ(-150px);
    filter: blur(20px);
    border-radius: 10px;
  }

  /* ANIMATIONS - BOX */
  @keyframes boxEnter {
    from {
      transform: scale(0.2) rotateX(-20deg) rotateY(-180deg);
    }
    to {
      transform: scale(1) rotateX(-20deg) rotateY(-30deg);
    }
  }

  @keyframes boxExit {
    from {
      transform: scale(1) rotateX(-20deg) rotateY(-30deg);
    }
    to {
      transform: scale(0) rotateX(-20deg) rotateY(-180deg);
    }
  }

  .box:global(.animate-box) {
    animation:
      0.8s cubic-bezier(0.34, 1.56, 0.64, 1) forwards boxEnter,
      0.8s ease-in 6s forwards boxExit;
  }

  /* ANIMATIONS - FLAPS */
  @keyframes frontFlapOpen {
    to {
      transform: translateZ(calc(var(--bd) / 2)) rotateX(70deg);
    }
  }

  @keyframes frontFlapClose {
    from {
      transform: translateZ(calc(var(--bd) / 2)) rotateX(70deg);
    }
    to {
      transform: translateZ(calc(var(--bd) / 2)) rotateX(270deg);
    }
  }

  .front-flap:global(.animate-front-flap) {
    animation:
      0.8s ease-out 1s forwards frontFlapOpen,
      0.6s ease-in 5.8s forwards frontFlapClose;
  }

  @keyframes backFlapOpen {
    to {
      transform: rotateY(180deg) translateZ(calc(var(--bd) / 2)) rotateX(-280deg);
    }
  }

  @keyframes backFlapClose {
    from {
      transform: rotateY(180deg) translateZ(calc(var(--bd) / 2)) rotateX(-280deg);
    }
    to {
      transform: rotateY(180deg) translateZ(calc(var(--bd) / 2)) rotateX(-90deg);
    }
  }

  .back-flap:global(.animate-back-flap) {
    animation:
      0.8s ease-out 1s forwards backFlapOpen,
      0.6s ease-in 5.8s forwards backFlapClose;
  }
</style>
