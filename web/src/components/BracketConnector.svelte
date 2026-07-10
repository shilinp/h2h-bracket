<script lang="ts">
    interface Props {
        type: 'incoming' | 'outgoing' | 'both';
        isEven?: boolean;
        lineHeight?: number; 
        gapWidth?: number;
        borderRadius?: number;
        isActiveMatch?: boolean;
        roundIndex: number;
        activeRoundIndex: number;
    }

    let { 
        type, 
        isEven = false, 
        lineHeight = 0, 
        gapWidth = 40,
        borderRadius = 12,
        isActiveMatch = false,
        roundIndex,
        activeRoundIndex
    }: Props = $props();

    let midX = $derived(gapWidth / 2);
    let signedLineHeight = $derived(isEven ? lineHeight : -lineHeight);
    
    let dir = $derived(isEven ? 1 : -1);
    let r = $derived(Math.min(borderRadius, Math.abs(lineHeight) / 2, midX));

    let roundedPath = $derived(
        `M 0 0 ` +
        `L ${midX - r} 0 ` +
        `Q ${midX} 0, ${midX} ${dir * r} ` +
        `L ${midX} ${signedLineHeight - (dir * r)} ` +
        `Q ${midX} ${signedLineHeight}, ${midX + r} ${signedLineHeight} ` +
        `L ${gapWidth} ${signedLineHeight}`
    );

    let shouldShowDot = $derived(roundIndex === activeRoundIndex);
</script>

<svg 
    class="bracket-svg" 
    class:active={isActiveMatch}
    style="width: {gapWidth}px; height: 1px; right: -{gapWidth}px;"
>
    {#if type === 'incoming'}
        <path 
            d="M 0 0 L {midX} 0" 
            class="connector-path" 
            class:active={isActiveMatch} 
        />
    {/if}

    {#if type === 'outgoing' || type === 'both'}
        <path 
            d={roundedPath} 
            class="connector-path" 
            class:active={isActiveMatch} 
        />
        
        {#if shouldShowDot}
            <circle 
                cx={gapWidth} 
                cy={signedLineHeight} 
                r="3" 
                class="connector-dot" 
                class:active={isActiveMatch} 
            />
        {/if}
    {/if}
</svg>

<svg 
    class="bracket-svg" 
    class:active={isActiveMatch}
    style="width: {gapWidth}px; height: 1px; right: -{gapWidth}px;"
>
    {#if type === 'incoming'}
        <path 
            d="M 0 0 L {midX} 0" 
            class="connector-path" 
            class:active={isActiveMatch} 
        />
    {/if}

    {#if type === 'outgoing' || type === 'both'}
        <path 
            d={roundedPath} 
            class="connector-path" 
            class:active={isActiveMatch} 
        />
        
        {#if shouldShowDot}
            <circle 
                cx={gapWidth} 
                cy={signedLineHeight} 
                r="3" 
                class="connector-dot" 
                class:active={isActiveMatch} 
            />
        {/if}
    {/if}
</svg>

<style>
    .bracket-svg {
        position: absolute;
        top: 50%; 
        overflow: visible; 
        pointer-events: none;
        z-index: 3;
    }

    .bracket-svg.active {
        z-index: 5;
    }

    .connector-path {
        stroke: var(--bracket-connector-inactive);
        stroke-width: 2px;
        fill: none;
        stroke-linecap: round;
        stroke-linejoin: round;
        transition: stroke 0.2s ease, stroke-width 0.2s ease;
    }
    
    .connector-dot {
        fill: var(--bracket-connector-inactive); 
        transition: fill 0.2s ease, r 0.2s ease;
        r: 3;
    }

    .connector-path.active {
        stroke: var(--dark-navy-blue);
        stroke-width: 2px;
    }

    .connector-dot.active {
        fill: var(--dark-navy-blue);
    }
</style>