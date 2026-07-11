<script lang="ts">
    import { onMount } from "svelte";
    import { createBracketState } from "./lib/state/bracket.svelte";
    import { createFetchBracketState } from "./lib/state/fetchBracket.svelte";
    import { createSubmitTeamsState } from "./lib/state/submitTeams.svelte";

    import {
        SubmitTeamsRequest,
        SubmitTeamsResponse,
        FetchBracketResponse,
        type Match,
    } from "./lib/proto/bracket";

    import BracketPreview from "./components/BracketPreview.svelte";

    type View = "upload" | "preview";

    interface UploadState {
        view: View;
        teamInput: string;
        pendingTeams: string[];
        statusMessage: string | null;
        statusIsError: boolean;
    }

    let state = $state<UploadState>({
        view: "upload",
        teamInput: "",
        pendingTeams: [],
        statusMessage: null,
        statusIsError: false,
    });

    const bracketState = createBracketState();
    const fetchBracketState = createFetchBracketState(bracketState);
    const submitTeamsState = createSubmitTeamsState(bracketState);

    let groupedMatches = $derived.by(
        () => bracketState.graph.presentationRounds,
    );

    function setStatus(msg: string, isError = false) {
        state.statusMessage = msg;
        state.statusIsError = isError;
    }

    function clearStatus() {
        state.statusMessage = null;
        state.statusIsError = false;
    }

    function addTeam() {
        const input = state.teamInput.trim();
        if (!input) return;
        if (
            state.pendingTeams.some(
                (t) => t.toLowerCase() === input.toLowerCase(),
            )
        ) {
            setStatus("That team is already in the list.", true);
            return;
        }
        state.pendingTeams.push(input);
        state.teamInput = "";
        clearStatus();
    }

    function removeTeam(index: number) {
        state.pendingTeams.splice(index, 1);
    }

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter") addTeam();
    }

    async function submitUpload() {
        if (state.pendingTeams.length < 2) {
            setStatus("Add at least two teams to create a bracket.", true);
            return;
        }
        clearStatus();
        await submitTeamsState.submitTeams(state.pendingTeams);
        state.pendingTeams = [];
        state.view = "preview";
        setStatus("Tournament created!");
    }

    function goToUpload() {
        state.view = "upload";
        state.pendingTeams = [];
        clearStatus();
    }

    onMount(async () => {
        await fetchBracketState.fetchBracket("", true);
        state.view = "preview";
    });
</script>

<main class="mobile-viewport">
    <div class="upload-page">
        <h1 class="title">🏆 Tournament Setup</h1>

        {#if fetchBracketState.isInProgress}
            <div class="center-flow text-muted">
                Loading existing tournament...
            </div>
        {:else if state.view === "upload"}
            <div class="section-card">
                <h2 class="section-title">Add Teams</h2>
                <p class="section-hint">
                    Enter one team name at a time. BYEs are inserted
                    automatically if needed.
                </p>

                <div class="input-row">
                    <input
                        type="text"
                        bind:value={state.teamInput}
                        onkeypress={handleKeyPress}
                        placeholder="Team name..."
                        class="form-input"
                        disabled={submitTeamsState.isInProgress}
                    />
                    <button
                        onclick={addTeam}
                        disabled={submitTeamsState.isInProgress}
                        class="btn-add"
                    >
                        Add
                    </button>
                </div>

                {#if state.pendingTeams.length > 0}
                    <div class="team-list">
                        {#each state.pendingTeams as team, i}
                            <div class="team-row">
                                <span class="team-name">{team}</span>
                                <button
                                    onclick={() => removeTeam(i)}
                                    class="btn-remove"
                                    disabled={submitTeamsState.isInProgress}
                                    >✕</button
                                >
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="empty-hint">
                        No teams yet. Add at least 2 to create a bracket.
                    </p>
                {/if}
            </div>

            <div class="warning-box">
                ⚠️ Uploading will <strong>delete all existing brackets</strong> and
                regenerate the tournament.
            </div>

            {#if state.statusMessage}
                <div class="status-msg" class:error={state.statusIsError}>
                    {state.statusMessage}
                </div>
            {/if}

            <button
                onclick={submitUpload}
                disabled={submitTeamsState.isInProgress ||
                    state.pendingTeams.length < 2}
                class="btn-primary"
            >
                {submitTeamsState.isInProgress
                    ? "Uploading..."
                    : `Upload Tournament (${state.pendingTeams.length} teams)`}
            </button>
        {:else}
            <div class="bracket-container preview-card">
                <BracketPreview
                    rounds={groupedMatches}
                    teamNames={bracketState.teamNames}
                    predictions={{}}
                    masterPredictions={{}}
                    isLocked={false}
                    activeMatchId={null}
                />
            </div>

            <div class="warning-box">
                ⚠️ Starting fresh will allow you to generate a new tournament,
                but uploading it will <strong>delete the current bracket</strong
                >.
            </div>

            {#if state.statusMessage}
                <div class="status-msg" class:error={state.statusIsError}>
                    {state.statusMessage}
                </div>
            {/if}

            <div class="action-row">
                <button
                    onclick={goToUpload}
                    disabled={submitTeamsState.isInProgress}
                    class="btn-secondary flex-1"
                >
                    Start Fresh
                </button>
            </div>
        {/if}
    </div>
</main>

<style>
    :global(body) {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            sans-serif;
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

    .upload-page {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
        gap: 12px;
        padding: 1rem;
        box-sizing: border-box;
    }

    .title {
        font-size: 1.5rem;
        font-weight: 800;
        margin: 0;
        text-align: center;
        flex-shrink: 0;
        color: #1e293b;
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

    .text-muted {
        color: #94a3b8;
    }

    .section-card {
        background: #1e293b;
        border-radius: 20px;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 12px;
        flex: 1;
        overflow: hidden;
    }

    .bracket-container {
        flex: 1;
        overflow-y: hidden;
        overflow-x: hidden;
        display: flex;
        flex-direction: column;
    }

    .preview-card {
        background: transparent;
    }

    .section-title {
        font-size: 1.1rem;
        font-weight: 700;
        margin: 0;
        flex-shrink: 0;
    }

    .section-hint {
        font-size: 0.85rem;
        color: #94a3b8;
        margin: 0;
        flex-shrink: 0;
    }

    .form-input {
        width: 100%;
        padding: 11px 14px;
        border-radius: 10px;
        border: 2px solid #334155;
        background: #0f172a;
        color: white;
        font-size: 0.95rem;
        box-sizing: border-box;
        transition: border-color 0.2s;
    }

    .form-input:focus {
        outline: none;
        border-color: #2563eb;
    }

    .form-input:disabled {
        opacity: 0.5;
    }

    .input-row {
        display: flex;
        gap: 8px;
        align-items: center;
        flex-shrink: 0;
    }

    .team-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
        flex: 1;
        overflow-y: auto;
    }

    .team-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        background: #0f172a;
        border: 1px solid #334155;
        border-radius: 10px;
        padding: 10px 12px;
    }

    .team-name {
        font-size: 0.95rem;
        font-weight: 500;
    }

    .empty-hint {
        color: #475569;
        font-size: 0.85rem;
        text-align: center;
        margin: auto 0;
    }

    .warning-box {
        background: rgba(245, 158, 11, 0.08);
        border: 1px solid rgba(245, 158, 11, 0.25);
        border-radius: 12px;
        padding: 10px 14px;
        font-size: 0.8rem;
        color: #fcd34d;
        line-height: 1.4;
        flex-shrink: 0;
    }

    .status-msg {
        padding: 10px 14px;
        border-radius: 12px;
        font-size: 0.85rem;
        text-align: center;
        background: rgba(34, 197, 94, 0.1);
        border: 1px solid rgba(34, 197, 94, 0.25);
        color: #86efac;
        flex-shrink: 0;
    }

    .status-msg.error {
        background: rgba(239, 68, 68, 0.1);
        border-color: rgba(239, 68, 68, 0.25);
        color: #fca5a5;
    }

    .btn-primary {
        padding: 14px 20px;
        border-radius: 12px;
        background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
        color: white;
        border: none;
        font-weight: 600;
        font-size: 0.95rem;
        cursor: pointer;
        transition: all 0.2s;
        width: 100%;
        flex-shrink: 0;
    }

    .btn-primary:hover:not(:disabled) {
        background: linear-gradient(135deg, #1d4ed8 0%, #1e40af 100%);
        transform: translateY(-1px);
        box-shadow: 0 8px 20px rgba(37, 99, 235, 0.3);
    }

    .btn-primary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .btn-secondary {
        padding: 14px 18px;
        border-radius: 12px;
        background: #334155;
        color: #cbd5e1;
        border: none;
        font-weight: 500;
        font-size: 0.9rem;
        cursor: pointer;
        transition: background 0.2s;
        white-space: nowrap;
        text-align: center;
    }

    .btn-secondary:hover:not(:disabled) {
        background: #475569;
    }

    .btn-secondary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .btn-add {
        padding: 11px 18px;
        border-radius: 10px;
        background: #2563eb;
        color: white;
        border: none;
        font-weight: 600;
        font-size: 0.9rem;
        cursor: pointer;
        white-space: nowrap;
        transition: background 0.2s;
        flex-shrink: 0;
    }

    .btn-add:hover:not(:disabled) {
        background: #1d4ed8;
    }

    .btn-add:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .btn-remove {
        background: #dc2626;
        border: none;
        color: white;
        width: 28px;
        height: 28px;
        border-radius: 7px;
        font-size: 1rem;
        cursor: pointer;
        transition: background 0.2s;
        padding: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }

    .btn-remove:hover:not(:disabled) {
        background: #b91c1c;
    }

    .btn-remove:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .action-row {
        display: flex;
        gap: 10px;
        align-items: stretch;
        flex-shrink: 0;
    }

    .flex-1 {
        flex: 1;
    }
</style>
