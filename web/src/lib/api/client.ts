import {
    FetchBracketResponse,
    SubmitBracketRequest,
    SubmitBracketResponse,
    DeleteBracketRequest,
    DeleteBracketResponse,
    SubmitTeamsRequest,
    SubmitTeamsResponse,
} from "../proto/bracket";

export const api = {
    async fetchBracket(username: string, isSpecialUser = false): Promise<FetchBracketResponse> {
        let queryString = `is_special_user=${isSpecialUser ? "true" : "false"}`;
        if (username.trim()) {
            queryString += `&username=${encodeURIComponent(username.trim())}`;
        }

        const res = await fetch(`/api/bracket?${queryString}`, {
            headers: { Accept: "application/json" },
        });
        if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);

        return FetchBracketResponse.fromJSON(await res.json());
    },

    async submitBracket(request: SubmitBracketRequest): Promise<SubmitBracketResponse> {
        const res = await fetch("/api/bracket", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Accept: "application/json",
            },
            body: JSON.stringify(SubmitBracketRequest.toJSON(request)),
        });
        if (!res.ok) throw new Error(`Submission failed: ${res.status}`);

        return SubmitBracketResponse.fromJSON(await res.json());
    },

    async deleteBracket(request: DeleteBracketRequest): Promise<DeleteBracketResponse> {
        const res = await fetch("/api/bracket", {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json",
                Accept: "application/json",
            },
            body: JSON.stringify(DeleteBracketRequest.toJSON(request)),
        });
        if (!res.ok) throw new Error(`Delete failed: ${res.status}`);

        return DeleteBracketResponse.fromJSON(await res.json());
    },

    async submitTeams(request: SubmitTeamsRequest): Promise<SubmitTeamsResponse> {
        const res = await fetch("/api/teams", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Accept: "application/json",
            },
            body: JSON.stringify(SubmitTeamsRequest.toJSON(request)),
        });
        if (!res.ok) throw new Error(`Delete failed: ${res.status}`);

        return SubmitTeamsResponse.fromJSON(await res.json());
    }
};