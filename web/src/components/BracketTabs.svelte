<!-- BracketTabs.svelte -->
<script lang="ts">
    import type { Match } from "../lib/proto/bracket";

    interface Props {
        rounds: { round: number; matches: Match[] }[];
        activeRoundIndex: number;
        onSelectRound: (index: number) => void;
    }

    let { rounds, activeRoundIndex, onSelectRound }: Props = $props();

    function getRoundName(index: number, totalRounds: number) {
        if (index === totalRounds - 1) return "Finals";
        if (index === totalRounds - 2) return "Semifinals";
        if (index === totalRounds - 3) return "Quarterfinals";
        const remainingRounds = totalRounds - index;
        const roundOf = 1 << remainingRounds;
        return `Round of ${roundOf}`;
    }
</script>

<div class="tabs-container">
    {#each rounds as _, i}
        <button
            class="tab"
            class:active={activeRoundIndex === i}
            onclick={() => onSelectRound(i)}
        >
            {getRoundName(i, rounds.length)}
        </button>
    {/each}
</div>

<style>
    .tabs-container {
        display: flex;
        gap: 8px;
        overflow-x: auto;
        scrollbar-width: none;
        flex-shrink: 0;
        position: sticky;
        top: 0;
        z-index: 10;
    }
    .tabs-container::-webkit-scrollbar {
        display: none;
    }
    .tab {
        background: #D4D4D4;
        color: #74777E;
        border: none;
        padding: 10px 20px;
        border-radius: 100px;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        white-space: nowrap;
        transition: all 0.2s ease;
    }
    .tab.active {
        background: #0E172B;
        color: #ffffff;
    }
</style>