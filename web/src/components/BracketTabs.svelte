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
        gap: 12px;
        padding: 20px 24px;
        overflow-x: auto;
        scrollbar-width: none;
        flex-shrink: 0;
        position: sticky;
        top: 0;
        background-color: #1a1f2e;
        z-index: 10;
    }
    .tabs-container::-webkit-scrollbar {
        display: none;
    }
    .tab {
        background: #2a3249;
        color: #94a3b8;
        border: none;
        padding: 10px 20px;
        border-radius: 20px;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
        white-space: nowrap;
        transition: all 0.2s ease;
    }
    .tab.active {
        background: #6366f1;
        color: #ffffff;
    }
</style>