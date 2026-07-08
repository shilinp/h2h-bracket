import { FetchBracketResponse, Match, MatchPosition } from "./lib/proto/bracket";

export function createBracketState() {
  let matches = $state<Match[]>([]);
  let matchPositions = $state<Record<number, MatchPosition>>({});
  let predictions = $state<Record<number, number>>({});
  let teamNames = $state<Record<number, string>>({});
  let masterPredictions = $state<Record<number, number>>({});
  let isLocked = $state(false);
  let accuracy = $state<number | null>(null);
  let hasPersistedBracket = $state(false)

  const graph = $derived.by(() => {
    const autoWinners: Record<number, number> = {};

    const sortedMatches = [...matches].sort((a, b) => {
      const posA = matchPositions[a.matchId]?.roundNumber ?? 0;
      const posB = matchPositions[b.matchId]?.roundNumber ?? 0;
      return posA - posB;
    });

    const resolvedMatches = sortedMatches.map((match) => {
      const position = matchPositions[match.matchId];
      const roundNumber = position?.roundNumber ?? 0;
      const visualPosition = position?.visualPosition ?? 0;

      let team1Id = match.team1Id;
      if (match.team1PrevMatchId) {
        team1Id = predictions[match.team1PrevMatchId] ?? autoWinners[match.team1PrevMatchId] ?? match.team1Id;
      }

      let team2Id = match.team2Id;
      if (match.team2PrevMatchId) {
        team2Id = predictions[match.team2PrevMatchId] ?? autoWinners[match.team2PrevMatchId] ?? match.team2Id;
      }

      const t1Name = team1Id != null ? teamNames[team1Id] : null;
      const t2Name = team2Id != null ? teamNames[team2Id] : null;

      const t1IsBye = t1Name?.toUpperCase() === "BYE";
      const t2IsBye = t2Name?.toUpperCase() === "BYE";

      if ((t1IsBye || t2IsBye) && team1Id != null && team2Id != null) {
        autoWinners[match.matchId] = t1IsBye ? team2Id : team1Id;
      }

      return { ...match, team1Id, team2Id, roundNumber, visualPosition };
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
          predictions[m.matchId] == null,
      ),
    };
  });

  function applyResponse(response: FetchBracketResponse) {
    matches = response.matches ?? [];
    matchPositions = response.matchPositions ?? {};
    isLocked = response.isLocked;
    accuracy = response.accuracy ?? null;

    const parseMap = <T>(protoMap: Record<string, T> | undefined): Record<number, T> => {
      const result: Record<number, T> = {};
      for (const [key, val] of Object.entries(protoMap || {})) {
        const numKey = Number(key);
        if (!isNaN(numKey)) result[numKey] = val;
      }
      return result;
    };

    predictions = parseMap(response.predictions);
    teamNames = parseMap(response.teamNames);
    masterPredictions = parseMap(response.masterPredictions);
    hasPersistedBracket = Object.keys(predictions).length > 0;
  }

  function selectWinner(matchId: number, winnerId: number) {
    if (isLocked) return;
    predictions = { ...predictions, [matchId]: winnerId };
  }

  function clearPredictions() {
    predictions = {};
  }

  return {
    get graph() { return graph; },
    get predictions() { return predictions; },
    get teamNames() { return teamNames; },
    get isLocked() { return isLocked; },
    get accuracy() { return accuracy; },
    get masterPredictions() { return masterPredictions; },
    get matchPositions() { return matchPositions; },
    get hasPersistedBracket() { return hasPersistedBracket },
    applyResponse,
    selectWinner,
    clearPredictions,
  };
}