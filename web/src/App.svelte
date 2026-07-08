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
    Match,
    MatchPosition,
    DeleteBracketResponse,
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
    const autoWinners: Record<number, number> = {};

    const sortedMatches = [...state.bracket.matches].sort((a, b) => {
      const posA = state.bracket.matchPositions[a.matchId]?.roundNumber ?? 0;
      const posB = state.bracket.matchPositions[b.matchId]?.roundNumber ?? 0;
      return posA - posB;
    });

    const resolvedMatches = sortedMatches.map((match) => {
      const position = state.bracket.matchPositions[match.matchId];
      const roundNumber = position?.roundNumber ?? 0;
      const visualPosition = position?.visualPosition ?? 0;

      let team1Id = match.team1Id;
      if (match.team1PrevMatchId) {
        team1Id =
          state.bracket.predictions[match.team1PrevMatchId] ??
          autoWinners[match.team1PrevMatchId];
      }

      let team2Id = match.team2Id;
      if (match.team2PrevMatchId) {
        team2Id =
          state.bracket.predictions[match.team2PrevMatchId] ??
          autoWinners[match.team2PrevMatchId];
      }

      const t1Name = team1Id != null ? state.bracket.teamNames[team1Id] : null;
      const t2Name = team2Id != null ? state.bracket.teamNames[team2Id] : null;

      const t1IsBye = t1Name?.toUpperCase() === "BYE";
      const t2IsBye = t2Name?.toUpperCase() === "BYE";

      if ((t1IsBye || t2IsBye) && team1Id != null && team2Id != null) {
        autoWinners[match.matchId] = t1IsBye ? team2Id : team1Id;
      }

      return {
        ...match,
        team1Id,
        team2Id,
        roundNumber,
        visualPosition,
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
          autoWinners[m.matchId] == null &&
          state.bracket.predictions[m.matchId] == null,
      ),
    };
  });

  let groupedMatches = $derived.by(() => bracketGraph.presentationRounds);
  let playableMatches = $derived.by(() => bracketGraph.playable);
  let currentMatch = $derived.by(() => playableMatches[0] ?? null);

  function applyBracketResponse(response: FetchBracketResponse) {
    state.bracket.matches = response.matches ?? [];
    state.bracket.matchPositions = response.matchPositions ?? {};
    state.bracket.isLocked = response.isLocked;
    state.bracket.accuracy = response.accuracy ?? null;

    const parseMap = <T,>(
      protoMap: Record<string, T> | undefined,
    ): Record<number, T> => {
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
    state.hasPersistedBracket =
      Object.keys(state.bracket.predictions).length > 0;
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
      if (state.username.trim()) params.set("username", state.username);
      params.set("is_special_user", "false");

      const res = await fetch(`/api/bracket?${params.toString()}`, {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) throw new Error("Network error fetching bracket data");
      ``;
      const response = FetchBracketResponse.fromJSON(await res.json());
      applyBracketResponse(response);
    } catch (err) {
      console.error("Unable to load bracket payload", err);
    } finally {
      state.isLoading = false;
    }
  }

  function selectWinner(matchId: number, winnerId: number) {
    if (state.bracket.isLocked) return;

    // Create a clean shallow copy to force Svelte 5 to trigger its reactivity graph
    state.bracket.predictions = {
      ...state.bracket.predictions,
      [matchId]: winnerId,
    };
  }

  async function finalizeAndSubmit() {
    state.isSubmitting = true;
    const request = SubmitBracketRequest.create({
      username: state.username,
      predictions: state.bracket.predictions,
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
        applyBracketResponse(response.updatedBracket);
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
          predictions={state.bracket.predictions}
          teamNames={state.bracket.teamNames}
          masterPredictions={state.bracket.masterPredictions}
          matchPositions={state.bracket.matchPositions}
          isLocked={state.bracket.isLocked}
          activeMatchId={currentMatch?.matchId ?? null}
        />
      </div>

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
