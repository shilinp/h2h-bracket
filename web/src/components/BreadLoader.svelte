<script lang="ts">
    interface Props {
        showText: boolean;
        size: number
    }

    let { 
        showText = true,
        size = 200
    }: Props = $props();
</script>

<div class="loader-container" style="--size: {size}px;">
    <div class="spinner-wrapper">
        <svg viewBox="0 0 100 100" class="spinner-svg">
            <circle cx="50" cy="50" r="40" class="track" />
            <circle cx="50" cy="50" r="40" class="indicator" />
        </svg>

        <div class="icon-container">
            <img src="/sandwich1.svg" alt="" class="burger-icon" />
        </div>
    </div>

    {#if showText}
        <p class="loading-text">Bread-y in a second..</p>
    {/if}
</div>

<style>
    .loader-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        width: var(--size);
        text-align: center;
    }

    .spinner-wrapper {
        position: relative;
        width: 100%;
        height: var(--size);
    }

    /* The entire SVG rotates, keeping the gaps perfectly intact */
    .spinner-svg {
        width: 100%;
        height: 100%;
        animation: spin 1.5s linear infinite;
    }

    /* Base properties for both lines */
    circle {
        fill: none;
        stroke-width: 8;
        stroke-linecap: round;
    }

    /* * Circumference = 2 * pi * 40 = ~251.3 
   * We draw a solid line for 175, and leave the remaining 76.3 as a gap.
  */
    .track {
        stroke: var(--dark-navy-blue);
        stroke-dasharray: 175 76.3;
        stroke-dashoffset: 0;
    }

    /* * We draw a solid green line for 45, and leave the remaining 206.3 as a gap.
   * Using a negative offset pushes the green stroke perfectly into the middle 
   * of the dark track's gap without needing any finicky transform rotations.
  */
    .indicator {
        stroke: var(--neon-yellow-lime);
        stroke-dasharray: 45 206.3;
        /* Track length (175) + half of the remaining space for the first gap (15.6) = 190.6 */
        stroke-dashoffset: -190.6;
    }

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
    }

    .burger-icon {
        width: 100%;
        height: 100%;
    }

    .loading-text {
        margin-top: 1rem;
        font-size: 1.25rem;
        font-weight: 500;
        color: var(--dark-navy-blue);
    }

    @keyframes spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }
</style>
