<script lang="ts">
  import { onMount } from "svelte";
  import { createBracketState } from "./lib/state/bracket.svelte";
  import { createFetchBracketState } from "./lib/state/fetchBracket.svelte";
  import { createSubmitBracketState } from "./lib/state/submitBracket.svelte";
  import { createDeleteState } from "./lib/state/deleteBracket.svelte";
  import SessionGate from "./components/SessionGate.svelte";
  import BracketPreview from "./components/BracketPreview.svelte";
  import MatchPicker from "./components/MatchPicker.svelte";
  import AccuracyPanel from "./components/AccuracyPanel.svelte";
  import BreadLoader from "./components/BreadLoader.svelte"; 
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
    username: string;
    isLoggedIn: boolean;
    statusMessage: string | null;
  }

  let state = $state<AppState>({
    username: "",
    isLoggedIn: false,
    statusMessage: null,
  });

  const bracketState = createBracketState();
  const fetchBracketState = createFetchBracketState(bracketState);
  const submitBracketState = createSubmitBracketState(bracketState);
  const deleteBracketState = createDeleteState(bracketState);

  let groupedMatches = $derived.by(() => bracketState.graph.presentationRounds);
  let playableMatches = $derived.by(() => bracketState.graph.playable);
  let currentMatch = $derived.by(() => playableMatches[0] ?? null);

  async function handleLogin() {
    if (!state.username.trim()) return;
    localStorage.setItem("bracket_username", state.username);
    state.isLoggedIn = true;
    await fetchBracketState.fetchBracket(state.username, false);
  }

  async function finalizeAndSubmit() {
    await submitBracketState.submitBracket(state.username, false);
    state.statusMessage =
      "Bracket submission complete. Waiting for master bracket results.";
  }

  async function handleReset() {
    if (bracketState.hasPersistedBracket) {
      await deleteBracketState.deleteBracket(state.username);
      state.statusMessage = "Server-persisted bracket has been deleted.";
    } else {
      bracketState.clearPredictions();
      state.statusMessage = "Local draft cleared.";
    }
  }

  onMount(() => {
    const savedUsername = localStorage.getItem("bracket_username");
    if (savedUsername) {
      state.username = savedUsername;
    }

    if (state.username) {
      state.isLoggedIn = true;
      fetchBracketState.fetchBracket(state.username, false);
    }
  });
</script>

<main class="mobile-viewport">
  {#if !state.isLoggedIn}
    <SessionGate bind:username={state.username} onsubmit={handleLogin} />
  {:else if fetchBracketState.isInProgress}
    <div class="center-flow">
      <BreadLoader showText={true} size={180} />
    </div>
  {:else}
    <div class="bracket-page">
      <div class="bracket-container">
        <BracketPreview
          rounds={groupedMatches}
          predictions={bracketState.predictions}
          teamNames={bracketState.teamNames}
          matchPositions={bracketState.matchPositions}
          isLocked={bracketState.isLocked}
          activeMatchId={currentMatch?.matchId ?? null}
        />
      </div>

      <div class="bottom-panel">
        {#if bracketState.isLocked}
          <AccuracyPanel
            accuracy={bracketState.accuracy}
          />
        {:else}
          <MatchPicker
            {currentMatch}
            teamNames={bracketState.teamNames}
            isSubmitting={submitBracketState.isInProgress || deleteBracketState.isInProgress}
            hasPersistedBracket={bracketState.hasPersistedBracket}
            onselect={(event) =>
              bracketState.selectWinner(event.matchId, event.winnerId)}
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
    background-color: var(--bracket-background);
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

  .bracket-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    gap: 12px;
    padding: 1rem;
    box-sizing: border-box;
  }

  .bracket-container {
    flex: 1;
    overflow-y: hidden;
    overflow-x: hidden;
    display: flex;
    flex-direction: column;
  }

  .bottom-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex-shrink: 0;
  }
</style>