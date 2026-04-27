<script lang="ts">
  interface Props {
    title: string;
    subtitle?: string;
  }

  let { title, subtitle }: Props = $props();
</script>

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
    <div class="not-connected-title">{title}</div>
    <div class="not-connected-warning">{subtitle}</div>
    <div class="sparkles sparkles-bottom">✨</div>
  </div>
</div>

<style>
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
