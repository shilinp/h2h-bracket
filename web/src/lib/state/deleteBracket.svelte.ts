import { api } from "../api/client"; 
import { DeleteBracketRequest } from "../proto/bracket";
import type { createBracketState } from "./bracket.svelte";

export function createDeleteState(bracketState: ReturnType<typeof createBracketState>) {
  let isInProgress = $state(false);

  async function deleteBracket(username: string) {
    isInProgress = true;
    try {
      const request = DeleteBracketRequest.create({ username });
      const response = await api.deleteBracket(request);
      if (response.updatedBracket) {
        bracketState.applyResponse(response.updatedBracket);
      }
    } catch (err) {
      console.error("Network sync execution failed during delete payload:", err);
      throw err;
    } finally {
      isInProgress = false;
    }
  }

  return {
    get isInProgress() { return isInProgress; },
    deleteBracket
  };
}