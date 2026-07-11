<script lang="ts">
    import { tick, untrack } from "svelte";
    import type { Match, MatchPosition } from "../lib/proto/bracket";
    import BracketTabs from "./BracketTabs.svelte";
    import BracketMatchGroup from "./BracketMatchGroup.svelte";

    interface Props {
        rounds: { round: number; matches: Match[] }[];
        predictions: Record<number, number>;
        teamNames: Record<number, string>;
        matchPositions?: Record<number, MatchPosition>;
        isLocked: boolean;
        activeMatchId: number | null;
    }

    let {
        rounds = [],
        predictions = {},
        teamNames = {},
        isLocked = false,
        activeMatchId = null,
    }: Props = $props();

    // 1. Horizontal gap between rounds: 2rem = 32px
    const GAP_WIDTH = 32;

    // 2. Decreasing this from 140px brings the matches vertically closer together
    // while keeping the canvas connector logic perfectly aligned.
    const BASE_MATCH_HEIGHT = 140;

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
                // 1. Check if we are actually advancing to a new column
                const isChangingRounds = activeRoundIndex !== roundIdx;

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
                        const viewRect =
                            scrollContainer.getBoundingClientRect();

                        const scrollTop =
                            scrollContainer.scrollTop +
                            (elRect.top - viewRect.top) -
                            viewRect.height / 2 +
                            elRect.height / 2;

                        const targetLeft = columnEl?.offsetLeft ?? 0;

                        if (isChangingRounds) {
                            // Advancing to a new round: Smoothly sweep horizontally and vertically
                            scrollContainer.scrollTo({
                                left: targetLeft,
                                top: scrollTop,
                                behavior: "smooth",
                            });
                        } else {
                            // Same round: Force the X-axis instantly to counter Safari's layout reset,
                            // and only apply smooth scrolling to the Y-axis.
                            scrollContainer.scrollLeft = targetLeft;
                            scrollContainer.scrollTo({
                                top: scrollTop,
                                behavior: "smooth",
                            });
                        }
                    }
                });
            }
        }
    }

    $effect(() => {
        if (activeMatchId !== null && !isUserNavigating) {
            untrack(() => {
                focusActiveMatch();
            });
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

    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
        class="bracket-viewport"
        bind:this={scrollContainer}
        onscrollend={handleScrollEnd}
        ontouchstart={() => {
            isUserNavigating = true;
        }}
    >
        <div class="bracket-wrapper">
            <div
                class="bracket-content"
                style="--max-matches: {matchesInView}; --match-height: {BASE_MATCH_HEIGHT}px; gap: 2rem;"
            >
                {#each rounds as group, i}
                    <div class="round-column">
                        {#each group.matches as match, matchIdx}
                            <div class="match-wrapper">
                                <BracketMatchGroup
                                    {match}
                                    {matchIdx}
                                    roundIndex={i}
                                    totalRounds={rounds.length}
                                    {activeRoundIndex}
                                    {matchesInView}
                                    selected={predictions[match.matchId]}
                                    computedLineHeight={calculateConnectorHeight(
                                        i,
                                        matchesInView,
                                    )}
                                    isActiveMatch={activeMatchId ===
                                        match.matchId}
                                    {isLocked}
                                    gapWidth={GAP_WIDTH}
                                    {teamNames}
                                    {registerMatch}
                                />
                            </div>
                        {/each}
                    </div>
                {/each}
            </div>
        </div>
    </div>
</div>

<style>
    .bracket-app {
        position: relative;
        display: flex;
        flex-direction: column;
        background-color: var(--background);
        border-radius: 20px;
        border: 1px solid var(--bracket-container-border);
        padding-top: 1rem;
        padding-left: 1rem;
        width: 100%;
        height: 100%;
        overflow: hidden;
        box-sizing: border-box;
        gap: 1rem;
    }
   .bracket-viewport {
        flex: 1;
        overflow: auto;
        position: relative;
        /* Force hardware acceleration to keep scroll position stable */
        -webkit-overflow-scrolling: touch; 
        transform: translateZ(0); 
    }
    
    .bracket-wrapper {
        /* Ensure this container is exactly as large as the content */
        display: inline-block;
        min-width: 100%;
        padding-bottom: 20px; /* Prevent bottom bounce clipping */
    }
    .bracket-content {
        display: flex;
        height: calc(var(--max-matches) * var(--match-height));
        transition: height 0.4s cubic-bezier(0.25, 1, 0.5, 1);
    }
    .round-column {
        display: flex;
        flex-direction: column;
        justify-content: space-around;
        width: 17.5rem;
        min-width: 280px;
        height: 100%;
        flex-shrink: 0;
        overflow: visible;
    }
    .match-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        height: var(--match-height);
        width: 100%;
    }
</style>
