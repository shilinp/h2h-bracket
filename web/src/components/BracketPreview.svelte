<!-- BracketPreview.svelte -->
<script lang="ts">
    import { tick } from "svelte";
    import type { Match, MatchPosition } from "../lib/proto/bracket";
    import BracketTabs from "./BracketTabs.svelte";
    import BracketMatchGroup from "./BracketMatchGroup.svelte";

    interface Props {
        rounds: { round: number; matches: Match[] }[];
        predictions: Record<number, number>;
        teamNames: Record<number, string>;
        masterPredictions: Record<number, number>;
        matchPositions?: Record<number, MatchPosition>;
        isLocked: boolean;
        activeMatchId: number | null;
    }

    let {
        rounds = [],
        predictions = {},
        teamNames = {},
        masterPredictions = {},
        isLocked = false,
        activeMatchId = null,
    }: Props = $props();

    const BASE_MATCH_HEIGHT = 140;
    const GAP_WIDTH = 40;

    let activeRoundIndex = $state(0);
    let matchElements = $state<Record<number, HTMLElement>>({});
    let scrollContainer = $state<HTMLElement | null>(null);
    let isUserNavigating = $state(false);

    let matchesInView = $state(rounds[0]?.matches.length ?? 1);
    let pendingMatchesInView = $state<number | null>(null);

    function calculateConnectorHeight(
        roundIndex: number,
        currentMatchesInView: number,
    ) {
        const totalCanvasHeight = currentMatchesInView * BASE_MATCH_HEIGHT;
        const matchesInRound = rounds[roundIndex]?.matches.length ?? 1;
        if (matchesInRound <= 1) return 0;
        return totalCanvasHeight / matchesInRound / 2;
    }

    function registerMatch(node: HTMLElement, matchId: number) {
        matchElements[matchId] = node;
        return {
            destroy() {
                delete matchElements[matchId];
            },
        };
    }

    function focusActiveMatch() {
        if (
            activeMatchId !== null &&
            matchElements[activeMatchId] &&
            scrollContainer
        ) {
            const roundIdx = rounds.findIndex((r) =>
                r.matches.some((m) => m.matchId === activeMatchId),
            );

            if (roundIdx !== -1) {
                const targetMatches = rounds[roundIdx]?.matches.length ?? 1;
                const isFinalRound = roundIdx === rounds.length - 1;

                if (targetMatches > matchesInView) {
                    matchesInView = targetMatches;
                    pendingMatchesInView = null;
                } else if (targetMatches < matchesInView) {
                    pendingMatchesInView = isFinalRound ? null : targetMatches;
                } else {
                    pendingMatchesInView = null;
                }
                activeRoundIndex = roundIdx;
            }

            tick().then(() => {
                if (!scrollContainer) return;
                const el = matchElements[activeMatchId!];
                if (el) {
                    const bracketContent =
                        scrollContainer.querySelector(".bracket-content");
                    const columnEl = bracketContent?.children[
                        roundIdx
                    ] as HTMLElement;
                    const elRect = el.getBoundingClientRect();
                    const viewRect = scrollContainer.getBoundingClientRect();

                    const scrollTop =
                        scrollContainer.scrollTop +
                        (elRect.top - viewRect.top) -
                        viewRect.height / 2 +
                        elRect.height / 2;

                    scrollContainer.scrollTo({
                        left: columnEl?.offsetLeft ?? 0,
                        top: scrollTop,
                        behavior: "smooth",
                    });
                }
            });
        }
    }

    $effect(() => {
        if (activeMatchId !== null && !isUserNavigating) {
            focusActiveMatch();
        }
    });

    function selectRound(index: number) {
        isUserNavigating = true;
        activeRoundIndex = index;

        const targetMatches = rounds[index]?.matches.length ?? 1;
        const isFinalRound = index === rounds.length - 1;

        if (targetMatches > matchesInView) {
            matchesInView = targetMatches;
            pendingMatchesInView = null;
        } else if (targetMatches < matchesInView) {
            pendingMatchesInView = isFinalRound ? null : targetMatches;
        } else {
            pendingMatchesInView = null;
        }

        if (scrollContainer) {
            const bracketContent =
                scrollContainer.querySelector(".bracket-content");
            const columnEl = bracketContent?.children[index] as HTMLElement;
            if (columnEl) {
                scrollContainer.scrollTo({
                    left: columnEl.offsetLeft,
                    top: 0,
                    behavior: "smooth",
                });
            }
        }
    }

    function handleScrollEnd() {
        if (pendingMatchesInView !== null) {
            matchesInView = pendingMatchesInView;
            pendingMatchesInView = null;
        }
    }

    $effect(() => {
        if (activeMatchId !== null) {
            isUserNavigating = false;
        }
    });
</script>

<div class="bracket-app">
    <BracketTabs {rounds} {activeRoundIndex} onSelectRound={selectRound} />

    <div
        class="bracket-viewport"
        bind:this={scrollContainer}
        onscrollend={handleScrollEnd}
    >
        <div
            class="bracket-content"
            style="--max-matches: {matchesInView}; gap: {GAP_WIDTH}px;"
        >
            {#each rounds as group, i}
                <div class="round-column">
                    {#each group.matches as match, matchIdx}
                        <BracketMatchGroup
                            {match}
                            {matchIdx}
                            roundIndex={i}
                            totalRounds={rounds.length}
                            {activeRoundIndex}
                            {matchesInView}
                            selected={predictions[match.matchId]}
                            masterWinner={masterPredictions[match.matchId]}
                            computedLineHeight={calculateConnectorHeight(
                                i,
                                matchesInView,
                            )}
                            isActiveMatch={activeMatchId === match.matchId}
                            {isLocked}
                            gapWidth={GAP_WIDTH}
                            {teamNames}
                            {registerMatch}
                        />
                    {/each}
                </div>
            {/each}
        </div>
    </div>
</div>

<style>
    .bracket-app {
        position: relative;
        display: flex;
        flex-direction: column;
        background-color: #faf9f6;
        border-radius: 1.5rem;
        border: 0.0625rem solid #eeeeef;
        padding-top: 1rem;
        padding-left: 1rem;
        width: 100%;
        height: 100%;
        min-height: 80vh;
        overflow: hidden;
    }
    .bracket-viewport {
        flex: 1;
        overflow-x: hidden;
        overflow-y: auto;

        margin: 1rem 0px 0px 0px;

        position: relative;
        scroll-behavior: smooth;
    }
    .bracket-content {
        display: flex;
        height: calc(var(--max-matches) * 140px);
        transition: height 0.4s cubic-bezier(0.25, 1, 0.5, 1);
    }
    .round-column {
        display: flex;
        flex-direction: column;
        min-width: 280px;
        height: 100%;
        flex-shrink: 0;
        overflow: visible;
    }
</style>
