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
        computedLineHeight: number;
        isActiveMatch: boolean;
        isLocked: boolean;
        gapWidth: number;
        teamNames: Record<number, string>;
        registerMatch: (
            node: HTMLElement,
            matchId: number,
        ) => { destroy: () => void };
    }

    let {
        match,
        matchIdx,
        roundIndex,
        totalRounds,
        activeRoundIndex,
        matchesInView,
        selected,
        computedLineHeight,
        isActiveMatch,
        isLocked,
        gapWidth,
        teamNames,
        registerMatch,
    }: Props = $props();
</script>

<div
    class="match-wrapper"
    class:out-of-scroll-bounds={roundIndex !== activeRoundIndex &&
        matchIdx >= matchesInView}
    class:active-match={isActiveMatch}
    use:registerMatch={match.matchId}
>
    <div class="match-container">
        <BracketMatchCard
            {match}
            {selected}
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
            {gapWidth}
            {isActiveMatch}
            {roundIndex}
            {activeRoundIndex}
        />
    {:else if roundIndex > 0}
        <BracketConnector
            type="incoming"
            {gapWidth}
            {isActiveMatch}
            {roundIndex}
            {activeRoundIndex}
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
        flex-direction: column;
        justify-content: center;
        height: 100%; 
    }

    .match-container {
        display: flex;
        flex-direction: column;
        width: 100%;
        position: relative;
        z-index: 2;
    }
</style>
