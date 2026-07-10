<script>
  // You can pass a custom size prop if you want to scale it
  export let size = 200;
</script>

<div class="loader-container" style="--size: {size}px;">
  <div class="spinner-wrapper">
    <svg viewBox="0 0 100 100" class="spinner-svg">
      <circle 
        cx="50" 
        cy="50" 
        r="40" 
        class="track"
      />
      <circle 
        cx="50" 
        cy="50" 
        r="40" 
        class="indicator"
      />
    </svg>
    
    <div class="icon-container">
      <img src="/sandwich1.svg" alt="" class="burger-icon" />
    </div>
  </div>

  <p class="loading-text">Bread-y in a second</p>
</div>

<style>

  .loader-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    font-family: var(--font-family);
    width: var(--size);
    text-align: center;
  }

  .spinner-wrapper {
    position: relative;
    width: 100%;
    height: var(--size);
  }

  .spinner-svg {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg); /* Start from the top */
    animation: spin 2s linear infinite;
  }

  /* Base circle properties */
  circle {
    fill: none;
    stroke-width: 8;
    stroke-linecap: round;
  }

  /* Dark track covering roughly 75% of the circle */
  .track {
    stroke: var(--dark-navy-blue);
    stroke-dasharray: 251.2; /* 2 * PI * R (2 * 3.1415 * 40) */
    stroke-dashoffset: 62.8; /* Hides 25% of the loop, creating the gap */
  }

  /* Lime indicator covering roughly 20% of the circle */
  .indicator {
    stroke: var(--neon-yellow-lime);
    stroke-dasharray: 251.2;
    /* Shows 20% of the track, offsets it to sit in the upper right quadrant */
    stroke-dashoffset: 201deg; 
    transform: rotate(110deg);
    transform-origin: center;
  }

  /* Center Icon Styling */
  .icon-container {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 25%;
    height: 25%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-dark);
  }

  .burger-icon {
    width: 100%;
    height: 100%;
  }

  /* Text Styling */
  .loading-text {
    margin-top: 1rem;
    font-size: 1.25rem;
    font-weight: 500;
    color: var(--color-dark);
    line-height: 1.3;
  }

  /* CSS Animation for the spin effect */
  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style>