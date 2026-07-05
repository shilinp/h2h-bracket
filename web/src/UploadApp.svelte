<script lang="ts">
    import type { SubmitTeamsRequest } from "./lib/proto/bracket";

    interface UploadState {
        teams: string[];
        teamInput: string;
        isSubmitting: boolean;
        statusMessage: string | null;
    }

    let state = $state<UploadState>({
        teams: [],
        teamInput: '',
        isSubmitting: false,
        statusMessage: null,
    });

    function addTeam() {
        const input = state.teamInput.trim();
        if (!input) return;

        if (state.teams.includes(input)) {
            state.statusMessage = 'Team already added';
            setTimeout(() => { state.statusMessage = null; }, 3000);
            return;
        }

        state.teams.push(input);
        state.teamInput = '';
        state.statusMessage = null;
    }

    function removeTeam(index: number) {
        state.teams.splice(index, 1);
    }

    async function submitTournament() {
        if (state.teams.length < 2) {
            state.statusMessage = 'Add at least two teams to create a bracket';
            return;
        }

        state.isSubmitting = true;

        try {
            const payload: SubmitTeamsRequest = {
                teams: state.teams,
            };

            const res = await fetch('/api/tournament?is_special_user=true', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!res.ok) {
                throw new Error(`Upload failed: ${res.status}`);
            }

            state.statusMessage = 'Tournament uploaded successfully!';
            state.teams = [];
            
            setTimeout(() => {
                state.statusMessage = null;
            }, 3000);

        } catch (err) {
            console.error('Tournament upload failed', err);
            state.statusMessage = 'Failed to upload tournament. Please try again.';
        } finally {
            state.isSubmitting = false;
        }
    }

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            addTeam();
        }
    }
</script>

<main class="mobile-viewport">
    <div class="upload-page">
        <h1 class="title">🏆 Tournament Upload</h1>

        <div class="matches-section">
            <div class="section-header">
                <h2 class="section-title">Participating Teams ({state.teams.length})</h2>
            </div>

            <div class="match-input-group">
                <input
                    type="text"
                    bind:value={state.teamInput}
                    onkeypress={handleKeyPress}
                    placeholder="Enter team name..."
                    class="form-input"
                />
                <button
                    onclick={addTeam}
                    disabled={state.isSubmitting}
                    class="btn-primary btn-add"
                >
                    Add Team
                </button>
            </div>

            <div class="matches-list">
                {#each state.teams as team, i}
                    <div class="match-card">
                        <div class="match-content">
                            <div class="team">{team}</div>
                        </div>
                        <button
                            onclick={() => removeTeam(i)}
                            class="btn-remove"
                        >
                            ✕
                        </button>
                    </div>
                {/each}
            </div>
        </div>

        {#if state.statusMessage}
            <div class={`status-message ${state.statusMessage.includes('successfully') ? 'success' : 'error'}`}>
                {state.statusMessage}
            </div>
        {/if}

        <button
            onclick={submitTournament}
            disabled={state.isSubmitting || state.teams.length < 2}
            class="btn-primary btn-submit"
        >
            {state.isSubmitting ? 'Uploading...' : 'Upload Tournament'}
        </button>
    </div>
</main>

<style>
    /* Retain all your existing CSS styles from the original file here */
    :global(body) {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        background-color: #0b0f19;
        color: #f1f5f9;
        overflow-y: auto;
    }
    /* ... rest of existing styles ... */
</style>