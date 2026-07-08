<script lang="ts">
  import { onMount } from "svelte";
  import { createBracketState } from "./bracket.svelte";

  import SessionGate from "./components/SessionGate.svelte";
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
    username: string;
    isLoggedIn: boolean;
    isLoading: boolean;
    isSubmitting: boolean;
    hasPersistedBracket: boolean;
    statusMessage: string | null;
  }

  let state = $state<AppState>({
    username: "",
    isLoggedIn: false,
    isLoading: true,
    isSubmitting: false,
    hasPersistedBracket: false,
    statusMessage: null
  });

  const bracketState = createBracketState();

  let groupedMatches = $derived.by(() => bracketState.graph.presentationRounds);
  let playableMatches = $derived.by(() => bracketState.graph.playable);
  let currentMatch = $derived.by(() => playableMatches[0] ?? null);

  async function handleLogin() {
    if (!state.username.trim()) return;
    localStorage.setItem("bracket_username", state.username);
    state.isLoggedIn = true;
    await fetchBracketData();
  }

  async function fetchBracketData() {
    state.isLoading = true;
    try {
      const params = new URLSearchParams();
      if (state.username.trim()) params.set("username", state.username);
      params.set("is_special_user", "false");

      const res = await fetch(`/api/bracket?${params.toString()}`, {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) throw new Error("Network error fetching bracket data");
      ``;
      const response = FetchBracketResponse.fromJSON(await res.json());
      bracketState.applyResponse(response);
    } catch (err) {
      console.error("Unable to load bracket payload", err);
    } finally {
      state.isLoading = false;
    }
  }

  async function finalizeAndSubmit() {
    state.isSubmitting = true;
    const request = SubmitBracketRequest.create({
      username: state.username,
      predictions: bracketState.predictions,
    });

    try {
      const res = await fetch("/api/bracket", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(SubmitBracketRequest.toJSON(request)),
      });
      if (!res.ok) throw new Error(`Submission failed: ${res.status}`);

      const response = SubmitBracketResponse.fromJSON(await res.json());

      state.statusMessage =
        "Bracket submission complete. Waiting for master bracket results.";

      if (response.updatedBracket) {
        bracketState.applyResponse(response.updatedBracket);
      }
    } catch (err) {
      console.error("Bracket submission failed", err);
      alert("Could not submit your bracket. Please try again.");
    } finally {
      state.isSubmitting = false;
    }
  }

  async function handleReset() {
    if (state.hasPersistedBracket) {
      await deletePersistedBracket();
      state.statusMessage = "Server-persisted bracket has been deleted.";
    } else {
      bracketState.clearPredictions();
      state.statusMessage = "Local draft cleared.";
    }
  }

  async function deletePersistedBracket() {
    state.isSubmitting = true;
    try {
      const request = DeleteBracketRequest.create({
        username: state.username,
      });
      const body = JSON.stringify(DeleteBracketRequest.toJSON(request));
      const res = await fetch("/api/bracket", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body,
      });
      if (!res.ok) throw new Error(`Delete failed: ${res.status}`);

      bracketState.clearPredictions();
      state.hasPersistedBracket = false;
      await fetchBracketData();
    } catch (err) {
      console.error("Bracket delete failed", err);
      alert("Could not reset your server-saved bracket.");
    } finally {
      state.isSubmitting = false;
    }
  }

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    const savedUsername = localStorage.getItem("bracket_username");
    const urlUsername = params.get("username");

    if (urlUsername) {
      state.username = urlUsername;
    } else if (savedUsername) {
      state.username = savedUsername;
    }

    if (state.username) {
      state.isLoggedIn = true;
      fetchBracketData();
    } else {
      state.isLoading = false;
    }
  });
</script>

<main class="mobile-viewport">
  {#if !state.isLoggedIn}
    <SessionGate bind:username={state.username} onsubmit={handleLogin} />
  {:else if state.isLoading}
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
            isSubmitting={state.isSubmitting}
            onreset={() => {
              location.reload();
            }}
          />
        {:else}
          <MatchPicker
            {currentMatch}
            teamNames={bracketState.teamNames}
            remainingCount={playableMatches.length}
            isSubmitting={state.isSubmitting}
            onselect={(event) => bracketState.selectWinner(event.matchId, event.winnerId)}
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
