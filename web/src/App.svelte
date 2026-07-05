<script lang="ts">
  import { onMount } from "svelte";
  import SessionGate from "./components/SessionGate.svelte";
  import BracketPreview from "./components/BracketPreview.svelte";
  import MatchPicker from "./components/MatchPicker.svelte";
  import AccuracyPanel from "./components/AccuracyPanel.svelte";
  import {
    FetchBracketResponse,
    SubmitBracketRequest,
    SubmitBracketResponse,
    DeleteBracketRequest,
    DeleteBracketResponse,
    Match,
    MatchPosition,
  } from "./lib/proto/bracket";

  interface BracketState {
    matches: Match[];
    matchPositions: Record<number, MatchPosition>;
    predictions: Record<number, number>;
    teamNames: Record<number, string>;
    masterPredictions: Record<number, number>;
    isLocked: boolean;
    accuracy: number | null;
  }

  interface AppState {
    username: string;
    isLoggedIn: boolean;
    isLoading: boolean;
    isSubmitting: boolean;
    hasPersistedBracket: boolean;
    statusMessage: string | null;
    bracket: BracketState;
  }

  let state = $state<AppState>({
    username: "",
    isLoggedIn: false,
    isLoading: true,
    isSubmitting: false,
    hasPersistedBracket: false,
    statusMessage: null,
    bracket: {
      matches: [],
      matchPositions: {},
      predictions: {},
      teamNames: {},
      masterPredictions: {},
      isLocked: false,
      accuracy: null,
    },
  });

  const bracketGraph = $derived.by(() => {
    const resolvedMatches = state.bracket.matches.map((match) => {
      const team1Id = match.team1PrevMatchId
        ? state.bracket.predictions[match.team1PrevMatchId]
        : match.team1Id;
      
      const team2Id = match.team2PrevMatchId
        ? state.bracket.predictions[match.team2PrevMatchId]
        : match.team2Id;

      const position = state.bracket.matchPositions[match.matchId];

      return {
        ...match,
        team1Id,
        team2Id,
        roundNumber: position?.roundNumber ?? 1,
        visualPosition: position?.visualPosition ?? 0,
      };
    });

    const groupedRounds = new Map<number, typeof resolvedMatches>();
    for (const match of resolvedMatches) {
      if (!groupedRounds.has(match.roundNumber)) {
        groupedRounds.set(match.roundNumber, []);
      }
      groupedRounds.get(match.roundNumber)!.push(match);
    }

    for (const [_, matches] of groupedRounds) {
      matches.sort((a, b) => a.visualPosition - b.visualPosition);
    }

    return {
      presentationRounds: Array.from(groupedRounds.entries())
        .map(([round, matches]) => ({ round, matches }))
        .sort((a, b) => a.round - b.round),

      playable: resolvedMatches.filter(
        (m) =>
          m.team1Id != null &&
          m.team2Id != null &&
          state.bracket.predictions[m.matchId] == null,
      ),
    };
  });

  let groupedMatches = $derived.by(() => bracketGraph.presentationRounds);
  let playableMatches = $derived.by(() => bracketGraph.playable);
  let currentMatch = $derived.by(() => playableMatches[0] ?? null);

  // Helper to hydrate bracket state from either initial fetch or submit response
  function applyBracketResponse(response: FetchBracketResponse) {
    state.bracket.matches = response.matches ?? [];
    state.bracket.matchPositions = response.matchPositions ?? {};
    state.bracket.isLocked = response.isLocked;
    state.bracket.accuracy = response.accuracy ?? null;

    // Helper to safely parse stringified proto maps into number keys
    const parseMap = <T>(protoMap: Record<string, T> | undefined): Record<number, T> => {
      const result: Record<number, T> = {};
      for (const [key, val] of Object.entries(protoMap || {})) {
        const numKey = Number(key);
        if (!isNaN(numKey)) result[numKey] = val;
      }
      return result;
    };

    state.bracket.predictions = parseMap(response.predictions);
    state.bracket.teamNames = parseMap(response.teamNames);
    state.bracket.masterPredictions = parseMap(response.masterPredictions);
    state.hasPersistedBracket = Object.keys(state.bracket.predictions).length > 0;
  }

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
      if (state.username.trim()) {
        params.set("username", state.username);
      }
      
      const res = await fetch(`/api/bracket?${params.toString()}`, {
        headers: {
          accept: "application/x-protobuf",
        },
      });
      if (!res.ok) throw new Error("Network error fetching bracket data");

      const responseBytes = new Uint8Array(await res.arrayBuffer());
      const response = FetchBracketResponse.decode(responseBytes);
      applyBracketResponse(response);
    } catch (err) {
      console.error("Unable to load bracket payload", err);
    } finally {
      state.isLoading = false;
    }
  }

  function selectWinner(matchId: number, winnerId: number) {
    if (state.bracket.isLocked) return;
    state.bracket.predictions[matchId] = winnerId;
  }

  async function finalizeAndSubmit() {
    state.isSubmitting = true;
    const request = SubmitBracketRequest.create({
      username: state.username,
      predictions: state.bracket.predictions,
    });
    const body = SubmitBracketRequest.encode(request).finish();

    try {
      const res = await fetch("/api/bracket", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-protobuf",
          accept: "application/x-protobuf",
        },
        body,
      });
      if (!res.ok) {
        throw new Error(`Submission failed: ${res.status}`);
      }
      const bytes = new Uint8Array(await res.arrayBuffer());
      const response = SubmitBracketResponse.decode(bytes);
      
      state.statusMessage = "Bracket submission complete. Waiting for master bracket results.";
      
      // Update local state dynamically without extra network trip
      if (response.updatedBracket) {
          applyBracketResponse(response.updatedBracket);
      }

      if (state.bracket.isLocked) {
        state.statusMessage =
          "Bracket is now locked and scored against the master bracket.";
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
      state.bracket.predictions = {};
      state.statusMessage = "Local draft cleared.";
    }
  }

  async function deletePersistedBracket() {
    state.isSubmitting = true;
    try {
      const request = DeleteBracketRequest.create({
        username: state.username,
      });
      const body = DeleteBracketRequest.encode(request).finish();
      const res = await fetch("/api/bracket", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/x-protobuf",
          accept: "application/x-protobuf",
        },
        body,
      });
      if (!res.ok) {
        throw new Error(`Delete failed: ${res.status}`);
      }
      
      state.bracket.predictions = {};
      state.bracket.isLocked = false;
      state.bracket.accuracy = null;
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
    <SessionGate username={state.username} onsubmit={handleLogin} />
  {:else if state.isLoading}
    <div class="center-flow text-muted">
      Parsing match matrix architecture...
    </div>
  {:else}
    <div class="bracket-page">
      {#if state.statusMessage}
        <div class="status-banner">{state.statusMessage}</div>
      {/if}

      <BracketPreview
        rounds={groupedMatches}
        predictions={state.bracket.predictions}
        teamNames={state.bracket.teamNames}
        masterPredictions={state.bracket.masterPredictions}
        isLocked={state.bracket.isLocked}
        activeMatchId={currentMatch?.matchId ?? null}
      />

      <div class="bottom-panel">
        {#if state.bracket.isLocked}
          <AccuracyPanel
            accuracy={state.bracket.accuracy}
            isSubmitting={state.isSubmitting}
            onreset={() => {
              location.reload();
            }}
          />
        {:else}
          <MatchPicker
            {currentMatch}
            teamNames={state.bracket.teamNames}
            remainingCount={playableMatches.length}
            isSubmitting={state.isSubmitting}
            onselect={(event) => selectWinner(event.matchId, event.winnerId)}
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
    max-width: 440px;
    margin: 0 auto;
    height: 100vh;
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
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
    gap: 16px;
    padding: 16px;
  }

  .status-banner {
    background: rgba(37, 99, 235, 0.12);
    border: 1px solid rgba(37, 99, 235, 0.24);
    color: #c7d2fe;
    border-radius: 18px;
    padding: 12px 16px;
    margin-bottom: 12px;
    text-align: center;
    font-size: 0.95rem;
  }

  .bottom-panel {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }
</style>