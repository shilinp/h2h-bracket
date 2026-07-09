import { api } from "../api/client"; 
import { SubmitTeamsRequest } from "../proto/bracket";
import type { createBracketState } from "./bracket.svelte";

export function createSubmitTeamsState(bracketState: ReturnType<typeof createBracketState>) {
  let isInProgress = $state(false);

  async function submitTeams(teams: string[]) {
    isInProgress = true;
    try {
      const request = SubmitTeamsRequest.create({ teams });
      const response = await api.submitTeams(request);
      if (response.updatedBracket) {
        bracketState.applyResponse(response.updatedBracket);
      }
    } catch (err) {
      console.error("Failed to upload team matrix architecture:", err);
      throw err;
    } finally {
      isInProgress = false;
    }
  }

  return {
    get isInProgress() { return isInProgress; },
    submitTeams
  };
}