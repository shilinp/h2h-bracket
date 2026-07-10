<!-- BracketPreview.svelte -->
<script lang="ts">
    import { tick } from "svelte";
    import type { Match, MatchPosition } from "../lib/proto/bracket";
    import BracketConnector from "./BracketConnector.svelte";

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
        matchPositions = {},
        isLocked = false,
        activeMatchId = null,
    }: Props = $props();

    const BASE_MATCH_HEIGHT = 115;
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

        const cardSpacing = totalCanvasHeight / matchesInRound;
        return cardSpacing / 2;
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
                    if (isFinalRound) {
                        pendingMatchesInView = null;
                    } else {
                        pendingMatchesInView = targetMatches;
                    }
                } else {
                    pendingMatchesInView = null;
                }

                activeRoundIndex = roundIdx;
            }

            tick().then(() => {
                if (!scrollContainer) return;

                const el = matchElements[activeMatchId];
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
            if (isFinalRound) {
                pendingMatchesInView = null;
            } else {
                pendingMatchesInView = targetMatches;
            }
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

    function getRoundName(index: number, totalRounds: number) {
        if (index === totalRounds - 1) return "Finals";
        if (index === totalRounds - 2) return "Semifinals";
        if (index === totalRounds - 3) return "Quarterfinals";
        const remainingRounds = totalRounds - index;
        const roundOf = 1 << remainingRounds;

        return `Round of ${roundOf}`;
    }
</script>

<div class="bracket-app">
    <div class="tabs-container">
        {#each rounds as group, i}
            <button
                class="tab"
                class:active={activeRoundIndex === i}
                onclick={() => selectRound(i)}
            >
                {getRoundName(i, rounds.length)}
            </button>
        {/each}
    </div>

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
                        {@const selected = predictions[match.matchId]}
                        {@const masterWinner = masterPredictions[match.matchId]}
                        {@const computedLineHeight = calculateConnectorHeight(
                            i,
                            matchesInView,
                        )}
                        {@const isActiveMatch = activeMatchId === match.matchId}

                        <div
                            class="match-wrapper"
                            class:out-of-scroll-bounds={i !==
                                activeRoundIndex && matchIdx >= matchesInView}
                            use:registerMatch={match.matchId}
                        >
                            <div class="match-container">
                                <div
                                    class="match-card"
                                    class:locked={isLocked}
                                    class:active-match={isActiveMatch}
                                >
                                    <!-- Team 1 Row -->
                                    <div
                                        class="team-row"
                                        class:winner={match.team1Id != null &&
                                            selected === match.team1Id}
                                    >
                                        <div class="team-identity">
                                            <span
                                                class="team-name"
                                                class:strikethrough={isLocked &&
                                                    masterWinner != null &&
                                                    selected ===
                                                        match.team1Id &&
                                                    selected !== masterWinner}
                                            >
                                                {match.team1Id != null
                                                    ? (teamNames[
                                                          match.team1Id
                                                      ] ?? "TBD")
                                                    : "TBD"}
                                            </span>
                                        </div>
                                        <div
                                            class="status-indicator"
                                            class:winner={match.team1Id !=
                                                null &&
                                                selected === match.team1Id}
                                        >
                                            {#if match.team1Id != null && selected != null}
                                                {#if selected === match.team1Id}
                                                    <!-- Checkmark for Win -->
                                                    <svg
                                                        viewBox="0 0 24 24"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        stroke-width="3"
                                                        width="16"
                                                        height="16"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            d="M5 13l4 4L19 7"
                                                        />
                                                    </svg>
                                                {:else}
                                                    <!-- X for Loss (Only if the other team was selected) -->
                                                    <svg
                                                        viewBox="0 0 24 24"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        stroke-width="3"
                                                        width="14"
                                                        height="14"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            d="M6 18L18 6M6 6l12 12"
                                                        />
                                                    </svg>
                                                {/if}
                                            {/if}
                                        </div>
                                    </div>

                                    <!-- Team 2 Row -->
                                    <div
                                        class="team-row"
                                        class:winner={match.team2Id != null &&
                                            selected === match.team2Id}
                                    >
                                        <div class="team-identity">
                                            <span
                                                class="team-name"
                                                class:strikethrough={isLocked &&
                                                    masterWinner != null &&
                                                    selected ===
                                                        match.team2Id &&
                                                    selected !== masterWinner}
                                            >
                                                {match.team2Id != null
                                                    ? (teamNames[
                                                          match.team2Id
                                                      ] ?? "TBD")
                                                    : "TBD"}
                                            </span>
                                        </div>
                                        <div
                                            class="status-indicator"
                                            class:winner={match.team2Id !=
                                                null &&
                                                selected === match.team2Id}
                                        >
                                            {#if match.team2Id != null && selected != null}
                                                {#if selected === match.team2Id}
                                                    <!-- Checkmark for Win -->
                                                    <svg
                                                        viewBox="0 0 24 24"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        stroke-width="3"
                                                        width="16"
                                                        height="16"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            d="M5 13l4 4L19 7"
                                                        />
                                                    </svg>
                                                {:else}
                                                    <!-- X for Loss (Only if the other team was selected) -->
                                                    <svg
                                                        viewBox="0 0 24 24"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        stroke-width="3"
                                                        width="14"
                                                        height="14"
                                                    >
                                                        <path
                                                            stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            d="M6 18L18 6M6 6l12 12"
                                                        />
                                                    </svg>
                                                {/if}
                                            {/if}
                                        </div>
                                    </div>

                                    {#if isLocked && masterWinner != null && selected !== masterWinner}
                                        <div class="master-override">
                                            Correct: {teamNames[masterWinner] ??
                                                "TBD"}
                                        </div>
                                    {/if}
                                </div>
                            </div>

                            {#if i < rounds.length - 1}
                                <BracketConnector
                                    type={i > 0 ? "both" : "outgoing"}
                                    isEven={matchIdx % 2 === 0}
                                    lineHeight={computedLineHeight}
                                    gapWidth={GAP_WIDTH}
                                />
                            {:else if i > 0}
                                <BracketConnector
                                    type="incoming"
                                    gapWidth={GAP_WIDTH}
                                />
                            {/if}
                        </div>
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
        background-color: #1a1f2e;
        border-radius: 24px;
        width: 100%;
        height: 100%;
        min-height: 80vh;
        overflow: hidden;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            Helvetica, Arial, sans-serif;
    }

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

    .bracket-viewport {
        flex: 1;
        overflow-x: hidden;
        overflow-y: auto;
        padding: 0 16px;
        position: relative;
        scroll-behavior: smooth;
    }

    .bracket-content {
        display: flex;
        height: calc(var(--max-matches) * 115px);
        transition: height 0.4s cubic-bezier(0.25, 1, 0.5, 1);
        padding-top: 20px;
        padding-bottom: 80px;
    }

    .round-column {
        display: flex;
        flex-direction: column;
        min-width: 260px;
        height: 100%;
        flex-shrink: 0;
        overflow: visible;
    }

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

    .match-card {
        background: #2a3249;
        border-radius: 12px;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        border: 2px solid transparent;
        transition: border-color 0.2s ease;
    }

    .match-card.active-match {
        border-color: #6366f1;
        box-shadow: 0 0 12px rgba(99, 102, 241, 0.25);
    }

    .team-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: #232a3f;
        transition: background 0.2s;
    }

    .team-row:first-child {
        background: #181b28;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    }

    .team-row.winner {
        background: #1e2436;
    }

    .team-identity {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .team-name {
        color: #f8fafc;
        font-size: 0.95rem;
        font-weight: 600;
    }

    .team-name.strikethrough {
        text-decoration: line-through;
        color: #64748b;
    }

    .status-indicator {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        color: #64748b;
    }

    .status-indicator.winner {
        color: #10b981;
    }

    .master-override {
        background: rgba(239, 68, 68, 0.15);
        color: #f87171;
        font-size: 0.75rem;
        text-align: center;
        padding: 6px;
        font-weight: 600;
    }
</style>
