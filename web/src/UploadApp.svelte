<script lang="ts">
    import { onMount } from 'svelte';
    import {
        SubmitTeamsRequest,
        SubmitTeamsResponse,
        FetchBracketResponse,
    } from './lib/proto/bracket';

    type View = 'upload' | 'adjust';

    interface UploadState {
        view: View;
        // Upload view
        teamInput: string;
        pendingTeams: string[];
        // Adjust view
        adjustTeams: { id: number; name: string }[];
        // Shared
        isSubmitting: boolean;
        isLoadingExisting: boolean;
        statusMessage: string | null;
        statusIsError: boolean;
    }

    let state = $state<UploadState>({
        view: 'upload',
        teamInput: '',
        pendingTeams: [],
        adjustTeams: [],
        isSubmitting: false,
        isLoadingExisting: false,
        statusMessage: null,
        statusIsError: false,
    });

    // ── Helpers ──────────────────────────────────────────────────────────────

    function setStatus(msg: string, isError = false) {
        state.statusMessage = msg;
        state.statusIsError = isError;
    }

    function clearStatus() {
        state.statusMessage = null;
        state.statusIsError = false;
    }

    async function postTeams(teams: string[]): Promise<SubmitTeamsResponse> {
        const req = SubmitTeamsRequest.create({ teams });
        const body = SubmitTeamsRequest.encode(req).finish();

        const res = await fetch('/api/teams', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-protobuf',
                accept: 'application/x-protobuf',
            },
            body,
        });

        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || `HTTP ${res.status}`);
        }

        const bytes = new Uint8Array(await res.arrayBuffer());
        return SubmitTeamsResponse.decode(bytes);
    }

    function bracketToAdjustTeams(bracket: FetchBracketResponse): { id: number; name: string }[] {
        return Object.entries(bracket.teamNames ?? {})
            .map(([id, name]) => ({ id: Number(id), name }))
            .filter(t => t.name.toUpperCase() !== 'BYE')
            .sort((a, b) => a.id - b.id);
    }

    // ── Upload view ───────────────────────────────────────────────────────────

    function addTeam() {
        const input = state.teamInput.trim();
        if (!input) return;
        if (state.pendingTeams.some(t => t.toLowerCase() === input.toLowerCase())) {
            setStatus('That team is already in the list.', true);
            return;
        }
        state.pendingTeams.push(input);
        state.teamInput = '';
        clearStatus();
    }

    function removeTeam(index: number) {
        state.pendingTeams.splice(index, 1);
    }

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === 'Enter') addTeam();
    }

    async function submitUpload() {
        if (state.pendingTeams.length < 2) {
            setStatus('Add at least two teams to create a bracket.', true);
            return;
        }
        state.isSubmitting = true;
        clearStatus();
        try {
            const resp = await postTeams(state.pendingTeams);
            if (resp.updatedBracket) {
                state.adjustTeams = bracketToAdjustTeams(resp.updatedBracket);
            }
            state.pendingTeams = [];
            state.view = 'adjust';
            setStatus('Tournament created! You can now adjust team names below.');
        } catch (err: any) {
            setStatus('Upload failed: ' + err.message, true);
        } finally {
            state.isSubmitting = false;
        }
    }

    // ── Adjust view ───────────────────────────────────────────────────────────

    async function submitAdjust() {
        const names = state.adjustTeams.map(t => t.name.trim()).filter(Boolean);
        if (names.length < 2) {
            setStatus('Need at least two non-empty team names.', true);
            return;
        }
        state.isSubmitting = true;
        clearStatus();
        try {
            const resp = await postTeams(names);
            if (resp.updatedBracket) {
                state.adjustTeams = bracketToAdjustTeams(resp.updatedBracket);
            }
            setStatus('Teams updated and bracket regenerated.');
        } catch (err: any) {
            setStatus('Update failed: ' + err.message, true);
        } finally {
            state.isSubmitting = false;
        }
    }

    function goToUpload() {
        state.view = 'upload';
        state.pendingTeams = [];
        clearStatus();
    }

    // ── On mount: check for an existing tournament ────────────────────────────

    onMount(async () => {
        state.isLoadingExisting = true;
        try {
            const res = await fetch('/api/bracket?is_special_user=true', {
                headers: { accept: 'application/x-protobuf' },
            });
            if (res.ok) {
                const bytes = new Uint8Array(await res.arrayBuffer());
                const bracket = FetchBracketResponse.decode(bytes);
                if (bracket.matches && bracket.matches.length > 0) {
                    state.adjustTeams = bracketToAdjustTeams(bracket);
                    state.view = 'adjust';
                }
            }
        } catch (_) {
            // No existing tournament — stay on upload view
        } finally {
            state.isLoadingExisting = false;
        }
    });
</script>

<main class="mobile-viewport">
    <div class="upload-page">
        <h1 class="title">🏆 Tournament Setup</h1>

        {#if state.isLoadingExisting}
            <div class="loading-msg">Loading existing tournament...</div>

        {:else if state.view === 'upload'}
            <!-- ── Upload view ── -->
            <div class="section-card">
                <h2 class="section-title">Add Teams</h2>
                <p class="section-hint">Enter one team name at a time. BYEs are inserted automatically if needed.</p>

                <div class="input-row">
                    <input
                        type="text"
                        bind:value={state.teamInput}
                        onkeypress={handleKeyPress}
                        placeholder="Team name..."
                        class="form-input"
                        disabled={state.isSubmitting}
                    />
                    <button onclick={addTeam} disabled={state.isSubmitting} class="btn-add">
                        Add
                    </button>
                </div>

                {#if state.pendingTeams.length > 0}
                    <div class="team-list">
                        {#each state.pendingTeams as team, i}
                            <div class="team-row">
                                <span class="team-name">{team}</span>
                                <button onclick={() => removeTeam(i)} class="btn-remove" disabled={state.isSubmitting}>✕</button>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <p class="empty-hint">No teams yet. Add at least 2 to create a bracket.</p>
                {/if}
            </div>

            <div class="warning-box">
                ⚠️ Uploading will <strong>delete all existing brackets</strong> and regenerate the tournament.
            </div>

            {#if state.statusMessage}
                <div class="status-msg" class:error={state.statusIsError}>{state.statusMessage}</div>
            {/if}

            <button
                onclick={submitUpload}
                disabled={state.isSubmitting || state.pendingTeams.length < 2}
                class="btn-primary"
            >
                {state.isSubmitting ? 'Uploading...' : `Upload Tournament (${state.pendingTeams.length} teams)`}
            </button>

        {:else}
            <!-- ── Adjust view ── -->
            <div class="section-card">
                <h2 class="section-title">Adjust Teams</h2>
                <p class="section-hint">Edit team names below. Saving regenerates the bracket and deletes all submitted brackets.</p>

                <div class="adjust-list">
                    {#each state.adjustTeams as team, i}
                        <div class="adjust-row">
                            <span class="adjust-number">{i + 1}</span>
                            <input
                                type="text"
                                bind:value={team.name}
                                class="form-input adjust-input"
                                disabled={state.isSubmitting}
                                placeholder="Team name"
                            />
                        </div>
                    {/each}
                </div>
            </div>

            <div class="warning-box">
                ⚠️ Saving changes will <strong>delete all existing brackets</strong> and regenerate the tournament.
            </div>

            {#if state.statusMessage}
                <div class="status-msg" class:error={state.statusIsError}>{state.statusMessage}</div>
            {/if}

            <div class="action-row">
                <button onclick={submitAdjust} disabled={state.isSubmitting} class="btn-primary flex-1">
                    {state.isSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
                <button onclick={goToUpload} disabled={state.isSubmitting} class="btn-secondary">
                    Start Fresh
                </button>
            </div>
        {/if}
    </div>
</main>

<style>
    :global(body) {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        background-color: #0b0f19;
        color: #f1f5f9;
        overflow-y: auto;
    }

    .mobile-viewport {
        max-width: 440px;
        margin: 0 auto;
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
    }

    .upload-page {
        display: flex;
        flex-direction: column;
        gap: 20px;
        padding: 24px 16px 40px;
    }

    .title {
        font-size: 1.8rem;
        font-weight: 800;
        margin: 0;
        text-align: center;
    }

    .loading-msg {
        color: #94a3b8;
        text-align: center;
        padding: 40px 0;
    }

    /* ── Section card ── */
    .section-card {
        background: #1e293b;
        border-radius: 20px;
        padding: 20px;
        display: flex;
        flex-direction: column;
        gap: 14px;
    }

    .section-title {
        font-size: 1.1rem;
        font-weight: 700;
        margin: 0;
    }

    .section-hint {
        font-size: 0.85rem;
        color: #94a3b8;
        margin: 0;
    }

    /* ── Inputs ── */
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
    }

    /* ── Team list (upload view) ── */
    .team-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
        max-height: 280px;
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
        margin: 8px 0 0;
    }

    /* ── Adjust list ── */
    .adjust-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
        max-height: 340px;
        overflow-y: auto;
    }

    .adjust-row {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .adjust-number {
        font-size: 0.8rem;
        color: #64748b;
        min-width: 20px;
        text-align: right;
    }

    .adjust-input {
        flex: 1;
    }

    /* ── Warning ── */
    .warning-box {
        background: rgba(245, 158, 11, 0.08);
        border: 1px solid rgba(245, 158, 11, 0.25);
        border-radius: 12px;
        padding: 12px 16px;
        font-size: 0.85rem;
        color: #fcd34d;
        line-height: 1.5;
    }

    /* ── Status ── */
    .status-msg {
        padding: 11px 16px;
        border-radius: 12px;
        font-size: 0.875rem;
        text-align: center;
        background: rgba(34, 197, 94, 0.1);
        border: 1px solid rgba(34, 197, 94, 0.25);
        color: #86efac;
    }

    .status-msg.error {
        background: rgba(239, 68, 68, 0.1);
        border-color: rgba(239, 68, 68, 0.25);
        color: #fca5a5;
    }

    /* ── Buttons ── */
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
    }

    .flex-1 {
        flex: 1;
    }
</style>