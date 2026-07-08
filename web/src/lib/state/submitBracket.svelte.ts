import { api } from "../api/client"; 
import { SubmitBracketRequest } from "../proto/bracket";
import type { createBracketState } from "./bracket.svelte";

export function createSubmitBracketState(bracketState: ReturnType<typeof createBracketState>) {
  let isInProgress = $state(false);

  async function submitBracket(username: string, isSpecialUser = false) {
    isInProgress = true;
    try {
      const request = SubmitBracketRequest.create({ 
        username, 
        predictions: bracketState.predictions,
        isSpecialUser
      });
      const response = await api.submitBracket(request);
      if (response.updatedBracket) {
        bracketState.applyResponse(response.updatedBracket);
      }
    } catch (err) {
      console.error("Network sync execution failed during submission:", err);
      throw err;
    } finally {
      isInProgress = false;
    }
  }

  return {
    get isInProgress() { return isInProgress; },
    submitBracket
  };
}