<script lang="ts">
  interface Props {
    message: string;
    displayColor: string;
    textColor: string;
    visible: boolean;
  }

  let { message, displayColor, textColor, visible }: Props = $props();
</script>

<div
  class="displayText"
  class:visible
  style="background: radial-gradient(108.43% 127.63% at 83.12% 0%, rgba(0, 0, 0, 0) 77.37%, rgba(0, 0, 0, 0.06) 77.38%), radial-gradient(92.01% 101.56% at 38.39% 87.58%, rgba(0, 0, 0, 0) 77.39%, rgba(255, 255, 255, 0.08) 77.4%), linear-gradient({displayColor}, {displayColor})"
>
  <div class="star" style="background-color: {textColor}"></div>
  <div class="nameText" style="color: {textColor}">
    {@html message}
  </div>
  <div class="star" style="background-color: {textColor}"></div>
</div>

<style>
  .displayText {
    display: flex;
    gap: 6px;
    align-items: center;
    transform: rotate(-1deg);
    padding: 8px 16px;
    border-radius: 101px;
    border: 3px solid #fff;
    color: #fff;
    background:
      radial-gradient(
        108.43% 127.63% at 83.12% 0,
        rgba(0, 0, 0, 0) 77.37%,
        rgba(0, 0, 0, 0.06) 77.38%
      ),
      radial-gradient(
        92.01% 101.56% at 38.39% 87.58%,
        rgba(0, 0, 0, 0) 77.39%,
        rgba(255, 255, 255, 0.08) 77.4%
      ),
      #ff9b9d;
    white-space: nowrap;
    text-align: center;
    opacity: 0;
    transition: opacity 0.3s;
  }

  .displayText.visible {
    opacity: 1;
  }

  .nameText {
    font-size: 35px;
    font-weight: 500;
    white-space: nowrap;
    text-align: center;
  }

  .star {
    width: 30px;
    height: 30px;
    background: #fff;
    clip-path: polygon(50% 0%, 60% 40%, 100% 50%, 60% 60%, 50% 100%, 40% 60%, 0% 50%, 40% 40%);
  }
</style>
