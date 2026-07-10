<!-- BracketConnector.svelte -->
<script lang="ts">
    interface Props {
        type: 'incoming' | 'outgoing' | 'both';
        isEven?: boolean;
        lineHeight?: number; 
        gapWidth?: number;   
    }

    let { type, isEven = false, lineHeight = 0, gapWidth = 40 }: Props = $props();

    let midX = $derived(gapWidth / 2);
    let signedLineHeight = $derived(isEven ? lineHeight : -lineHeight);
</script>

<svg 
    class="bracket-svg" 
    style="width: {gapWidth}px; height: 1px; right: -{gapWidth}px;"
>
    {#if type === 'incoming'}
        <path d="M 0 0 L {midX} 0" class="connector-path" />
    {/if}

    {#if type === 'outgoing' || type === 'both'}
        <!-- Smooth S-Curve blending into the next round match -->
        <path 
            d="M 0 0 C {midX} 0, {midX} {signedLineHeight}, {gapWidth} {signedLineHeight}" 
            class="connector-path" 
        />
    {/if}
</svg>

<style>
    .bracket-svg {
        position: absolute;
        top: 50%; 
        overflow: visible; 
        pointer-events: none;
        z-index: 1;
    }

    .connector-path {
        stroke: rgba(255, 255, 255, 0.2);
        stroke-width: 2px;
        fill: none;
    }
</style>