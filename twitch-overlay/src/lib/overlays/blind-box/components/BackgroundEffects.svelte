<script lang="ts">
  interface Props {
    show: boolean;
  }

  let { show }: Props = $props();
</script>

<div class="background-effects" class:animate={show}>
  <div class="light-beam"></div>
  <div class="stars star-1"></div>
  <div class="stars star-2"></div>
  <div class="stars star-3"></div>
</div>

<style>
  .background-effects {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    overflow: visible;
  }

  .background-effects .light-beam,
  .background-effects .stars {
    opacity: 0;
    transition: opacity 0.5s ease-in-out;
    pointer-events: none;
  }

  .background-effects:global(.animate) .light-beam,
  .background-effects:global(.animate) .stars {
    opacity: 1;
  }

  .light-beam {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 600px;
    height: 600px;
    transform: translate(-50%, -50%);
    background: conic-gradient(
      #fff 0deg 15deg,
      transparent 15deg 45deg,
      #fff 45deg 60deg,
      transparent 60deg 90deg,
      #fff 90deg 105deg,
      transparent 105deg 135deg,
      #fff 135deg 150deg,
      transparent 150deg 180deg,
      #fff 180deg 195deg,
      transparent 195deg 225deg,
      #fff 225deg 240deg,
      transparent 240deg 270deg,
      #fff 270deg 285deg,
      transparent 285deg 315deg,
      #fff 315deg 330deg,
      transparent 330deg 360deg
    );
    mask-image: radial-gradient(circle, rgba(0, 0, 0, 1) 0%, rgba(0, 0, 0, 0) 70%);
    -webkit-mask-image: radial-gradient(circle, rgba(0, 0, 0, 1) 0%, rgba(0, 0, 0, 0) 70%);
    border-radius: 50%;
    animation: 40s linear infinite rotateRays;
  }

  .stars {
    position: absolute;
    width: 30px;
    height: 30px;
    background: #fff;
    clip-path: polygon(50% 0%, 60% 40%, 100% 50%, 60% 60%, 50% 100%, 40% 60%, 0% 50%, 40% 40%);
    animation: 4s ease-in-out infinite twinkle;
  }

  .star-1 {
    top: 50%;
    left: 40%;
    animation-delay: 0s;
  }

  .star-2 {
    top: 30%;
    right: 40%;
    animation-delay: 1s;
  }

  .star-3 {
    bottom: 34%;
    left: 40%;
    animation-delay: 2s;
  }

  @keyframes rotateRays {
    0% {
      transform: translate(-50%, -50%) rotate(0);
    }
    100% {
      transform: translate(-50%, -50%) rotate(360deg);
    }
  }

  @keyframes twinkle {
    0%,
    100% {
      transform: scale(1);
    }
    50% {
      transform: scale(0.7);
    }
  }
</style>
