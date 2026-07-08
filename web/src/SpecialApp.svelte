<script lang="ts">
    import { onMount } from "svelte";
    import { createBracketState } from "./lib/state/bracket.svelte";
    import { createFetchBracketState } from "./lib/state/fetchBracket.svelte";
    import { createSubmitBracketState } from "./lib/state/submitBracket.svelte";
    import { createDeleteState } from "./lib/state/deleteBracket.svelte";

    import BracketPreview from "./components/BracketPreview.svelte";
    import MatchPicker from "./components/MatchPicker.svelte";
    import AccuracyPanel from "./components/AccuracyPanel.svelte";
    import {
        FetchBracketResponse,
        SubmitBracketRequest,
        SubmitBracketResponse,
        DeleteBracketRequest,
        Match,
        MatchPosition,
        DeleteBracketResponse,
    } from "./lib/proto/bracket";

    interface AppState {
        statusMessage: string | null;
    }

    let state = $state<AppState>({
        statusMessage: null,
    });

    const bracketState = createBracketState();
    const fetchBracketState = createFetchBracketState(bracketState);
    const submitBracketState = createSubmitBracketState(bracketState);
    const deleteBracketState = createDeleteState(bracketState);

    let groupedMatches = $derived.by(
        () => bracketState.graph.presentationRounds,
    );
    let playableMatches = $derived.by(() => bracketState.graph.playable);
    let currentMatch = $derived.by(() => playableMatches[0] ?? null);

    async function finalizeAndSubmit() {
        await submitBracketState.submitBracket("special", true)
        state.statusMessage = "Bracket persisted to server successfully.";
    }

    async function handleReset() {
        if (bracketState.hasPersistedBracket) {
            await deleteBracketState.deleteBracket("special")
            state.statusMessage = "Server-persisted bracket has been deleted.";
        } else {
            bracketState.clearPredictions();
            state.statusMessage = "Local draft cleared.";
        }
    }

    onMount(async () => {
        fetchBracketState.fetchBracket("special", true)
    });
</script>

<main class="mobile-viewport">
    {#if fetchBracketState.isInProgress}
        <div class="center-flow text-muted">
            Parsing match matrix architecture...
        </div>
    {:else}
        <div class="bracket-page">
            {#if state.statusMessage}
                <div class="status-banner">{state.statusMessage}</div>
            {/if}

            <div class="preview-scroll-container">
                <BracketPreview
                    rounds={groupedMatches}
                    predictions={bracketState.predictions}
                    teamNames={bracketState.teamNames}
                    masterPredictions={bracketState.masterPredictions}
                    matchPositions={bracketState.matchPositions}
                    isLocked={bracketState.isLocked}
                    activeMatchId={currentMatch?.matchId ?? null}
                />
            </div>

            <div class="bottom-panel">
                {#if bracketState.isLocked}
                    <AccuracyPanel
                        accuracy={bracketState.accuracy}
                        isSubmitting={submitBracketState.isInProgress || deleteBracketState.isInProgress}
                        onreset={() => {
                            location.reload();
                        }}
                    />
                {:else}
                    <MatchPicker
                        {currentMatch}
                        teamNames={bracketState.teamNames}
                        remainingCount={playableMatches.length}
                        isSubmitting={submitBracketState.isInProgress || deleteBracketState.isInProgress}
                        onselect={(event) =>
                            bracketState.selectWinner(
                                event.matchId,
                                event.winnerId,
                            )}
                        onsubmit={finalizeAndSubmit}
                        onreset={handleReset}
                    />
                {/if}
            </div>
        </div>
    {/if}
</main>

<style>
    :global(body) {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            sans-serif;
        background-color: #0b0f19;
        color: #f1f5f9;
        overflow: hidden;
    }

    .mobile-viewport {
        width: 100vw;
        max-width: 440px;
        margin: 0 auto;
        height: 100dvh;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
        overflow: hidden;
    }

    .center-flow {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        text-align: center;
        height: 100%;
        padding: 24px;
        box-sizing: border-box;
    }

    .text-muted {
        color: #94a3b8;
    }

    .bracket-page {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
        gap: 12px;
        padding: 12px;
        box-sizing: border-box;
    }

    .preview-scroll-container {
        flex: 1;
        overflow-y: auto;
        overflow-x: hidden;
        -webkit-overflow-scrolling: touch;
    }

    .status-banner {
        background: rgba(37, 99, 235, 0.12);
        border: 1px solid rgba(37, 99, 235, 0.24);
        color: #c7d2fe;
        border-radius: 12px;
        padding: 10px 14px;
        text-align: center;
        font-size: 0.9rem;
        flex-shrink: 0;
    }

    .bottom-panel {
        display: flex;
        flex-direction: column;
        gap: 12px;
        flex-shrink: 0;
    }
</style>
