<script lang="ts">
    interface Props {
        accuracy: number | null;
    }

    let { accuracy = null }: Props = $props();
</script>

<div class="accuracy-panel">
    <span class="title">The results are in...</span>

    {#if accuracy !== null}
        <span class="subtitle"
            >You got a {accuracy.toFixed(1)}% match with Miranda. Thanks for
            playing :&rpar;</span
        >
    {:else}
        <span class="subtitle">You didn't submit any guesses!</span>
    {/if}
    <img src="/fish-sandwich.svg" alt="" />
</div>

<style>
    .accuracy-panel {
        text-align: center;
        color: var(--dark-navy-blue);
        font-size: 1rem;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        box-sizing: border-box;
        width: 100%;
        height: 100%;

        aspect-ratio: 1.5 / 1;

        border: 1px solid var(--match-picker-container-border);
        border-radius: 20px;
        padding: 1.5rem 2rem 1.5rem 2rem;
    }

    .title {
        color: var(--dark-navy-blue);
        font-weight: 500;
        font-size: 1rem;
    }

    .subtitle {
        color: var(--dark-navy-blue);
        font-weight: 700;
        font-size: 1rem;
    }
</style>
