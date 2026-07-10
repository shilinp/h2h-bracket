<!-- BracketMatchGroup.svelte -->
<script lang="ts">
    import type { Match } from "../lib/proto/bracket";
    import BracketMatchCard from "./BracketMatchCard.svelte";
    import BracketConnector from "./BracketConnector.svelte";

    interface Props {
        match: Match;
        matchIdx: number;
        roundIndex: number;
        totalRounds: number;
        activeRoundIndex: number;
        matchesInView: number;
        selected: number | undefined;
        masterWinner: number | undefined;
        computedLineHeight: number;
        isActiveMatch: boolean;
        isLocked: boolean;
        gapWidth: number;
        teamNames: Record<number, string>;
        registerMatch: (node: HTMLElement, matchId: number) => { destroy: () => void };
    }

    let {
        match,
        matchIdx,
        roundIndex,
        totalRounds,
        activeRoundIndex,
        matchesInView,
        selected,
        masterWinner,
        computedLineHeight,
        isActiveMatch,
        isLocked,
        gapWidth,
        teamNames,
        registerMatch
    }: Props = $props();
</script>

<div
    class="match-wrapper"
    class:out-of-scroll-bounds={roundIndex !== activeRoundIndex && matchIdx >= matchesInView}
    use:registerMatch={match.matchId}
>
    <div class="match-container">
        <BracketMatchCard
            {match}
            {selected}
            {masterWinner}
            {teamNames}
            {isLocked}
            {isActiveMatch}
        />
    </div>

    {#if roundIndex < totalRounds - 1}
        <BracketConnector
            type={roundIndex > 0 ? "both" : "outgoing"}
            isEven={matchIdx % 2 === 0}
            lineHeight={computedLineHeight}
            gapWidth={gapWidth}
        />
    {:else if roundIndex > 0}
        <BracketConnector
            type="incoming"
            gapWidth={gapWidth}
        />
    {/if}
</div>

<style>
    .match-wrapper.out-of-scroll-bounds {
        display: none !important;
    }
    .match-wrapper {
        position: relative;
        width: 100%;
        display: flex;
        flex: 1;
        flex-direction: column;
        justify-content: center;
        min-height: 115px;
    }
    .match-container {
        display: flex;
        flex-direction: column;
        gap: 8px;
        width: 100%;
        background-color: #1a1f2e;
        position: relative;
        z-index: 2;
    }
</style>