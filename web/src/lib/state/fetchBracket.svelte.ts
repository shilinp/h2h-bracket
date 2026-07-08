import { api } from "../api/client"; 
import type { createBracketState } from "./bracket.svelte";

export function createFetchBracketState(bracketState: ReturnType<typeof createBracketState>) {
  let isInProgress = $state(false);

  async function fetchBracket(username: string, isSpecialUser = false) {
    isInProgress = true;
    try {
      const response = await api.fetchBracket(username, isSpecialUser);
      bracketState.applyResponse(response);
    } catch (err) {
      console.error("Network sync execution failed during fetch:", err);
      throw err;
    } finally {
      isInProgress = false;
    }
  }
  return {
    get isInProgress() { return isInProgress; },
    fetchBracket
  };
}