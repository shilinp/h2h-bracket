<!-- BracketConnector.svelte -->
<script lang="ts">
    interface Props {
        type: 'incoming' | 'outgoing' | 'both';
        isEven?: boolean;
        lineHeight?: number; // The exact pixel distance to travel vertically
        gapWidth?: number;   // Dynamic horizontal gap width between columns
    }

    let { type, isEven = false, lineHeight = 0, gapWidth = 32 }: Props = $props();

    // The horizontal midpoint where lines shift vertically
    let midX = $derived(gapWidth / 2);
    
    // Ensure accurate signed vertical displacement
    let signedLineHeight = $derived(isEven ? lineHeight : -lineHeight);
</script>

<!-- 
  By omitting the viewBox and explicitly assigning layout dimensions, 
  SVG coordinates map 1:1 directly to screen pixels without scaling distortion.
-->
<svg 
    class="bracket-svg" 
    style="width: {gapWidth}px; height: 1px; right: -{gapWidth}px;"
>
    {#if type === 'incoming'}
        <!-- Pure incoming: line hits middle-left edge at the current card's center -->
        <path d="M 0 0 L {midX} 0" class="connector-path" />
    {/if}

    {#if type === 'outgoing' || type === 'both'}
        <!-- 
          Outgoing/Both: Always start from the card edge (0,0), move to the split point (midX,0),
          travel vertically to match the next round's height, then run to the edge of the next card.
        -->
        <path d="M 0 0 L {midX} 0 L {midX} {signedLineHeight} L {gapWidth} {signedLineHeight}" class="connector-path" />
    {/if}
</svg>

<style>
    .bracket-svg {
        position: absolute;
        top: 50%; /* Center alignment on the middle of the card container */
        overflow: visible; /* Allows paths to draw perfectly outside the 1px tall bounding box */
        pointer-events: none;
        z-index: 1;
    }

    .connector-path {
        stroke: #3b4358;
        stroke-width: 2px;
        fill: none;
    }
</style>