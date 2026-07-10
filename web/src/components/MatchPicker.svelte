<script lang="ts">
    import type { Match } from '../lib/proto/bracket';
    import MatchIcon from "./MatchIcon.svelte"; // 1. Import the icon component

    interface Props {
        currentMatch?: (Match & { roundNumber?: number }) | null;
        teamNames?: Record<number, string>;
        remainingCount?: number;
        isSubmitting?: boolean;
        onselect: (event: { matchId: number; winnerId: number }) => void;
        onsubmit: () => void;
        onreset: () => void;
    }

    let { 
        currentMatch = null, 
        teamNames = {},
        remainingCount = 0, 
        isSubmitting = false,
        onselect,
        onsubmit,
        onreset
    }: Props = $props();

    function chooseWinner(teamId: number | undefined | null) {
        if (!currentMatch || teamId == null) return;
        onselect({ matchId: currentMatch.matchId, winnerId: teamId });
    }
</script>

<div class="picker-panel">
    {#if currentMatch}
        <h2 class="picker-heading">Which sandwich is Miranda choosing? 🤔</h2>

        <div class="choice-row">
            <button
                class="choice-pane"
                onclick={() => chooseWinner(currentMatch.team1Id)}
                disabled={isSubmitting}
            >
                <MatchIcon teamId={currentMatch.team1Id} sizePx={48}/>
                <span class="team-name">{currentMatch.team1Id != null ? (teamNames[currentMatch.team1Id] ?? 'TBD') : 'TBD'}</span>
            </button>

            <button
                class="choice-pane"
                onclick={() => chooseWinner(currentMatch.team2Id)}
                disabled={isSubmitting}
            >
                <MatchIcon teamId={currentMatch.team2Id} sizePx={48} />
                <span class="team-name">{currentMatch.team2Id != null ? (teamNames[currentMatch.team2Id] ?? 'TBD') : 'TBD'}</span>
            </button>
        </div>
    {:else}
        <div class="picker-empty">
            <span class="success-icon">🎉</span>
            <h2>Bracket Generated!</h2>
            <p>Your comparative evaluations are complete. Lock it into the database?</p>
            <button class="btn-primary" onclick={onsubmit} disabled={isSubmitting}>
                {isSubmitting ? 'Transmitting to Go Core...' : 'Lock Submission'}
            </button>
            <button class="btn-text" onclick={onreset} disabled={isSubmitting}>
                Wipe Local Progress & Restart
            </button>
        </div>
    {/if}
</div>

<style>
    .picker-panel {
        background: var(--match-picker-container-background);
        border: 1px solid var(--match-picker-container-border);
        border-radius: 20px;
        padding: 1.5rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        width: 100%;
        max-width: 600px;
        aspect-ratio: 1.5 / 1;
        box-sizing: border-box;
    }

    .picker-heading {
        color: var(--dark-navy-blue);
        font-size: 1rem;
        font-weight: 500;
        text-align: center;
        margin: 0;
    }

    .choice-row {
        display: flex;
        gap: 1.5rem;
        width: 100%;
        flex: 1;
        margin-top: 16px;
        box-sizing: border-box;
    }

    .choice-pane {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        border-radius: 20px;
        border: 1px solid var(--border);
        background: #ffffff;
        color: var(--dark-navy-blue);
        font-size: 1rem;
        font-weight: 500;
        padding: 1rem;
        box-sizing: border-box;
        transition: transform 0.15s ease, box-shadow 0.15s ease;
    }

    .team-name {
        font-size: 1rem;
        text-align: center;
        width: 100%;
        overflow-wrap: break-word;
        color: var(--dark-navy-blue);
    }

    .picker-empty {
        background: #ffffff;
        border-radius: 28px;
        padding: 24px;
        text-align: center;
        color: #0f172a;
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 12px;
        align-items: center;
        box-sizing: border-box;
        width: 100%;
        height: 100%;
    }

    .success-icon {
        font-size: 2.2rem;
    }

    .picker-empty h2 {
        margin: 0;
        font-size: 1.25rem;
        font-weight: 600;
    }

    .picker-empty p {
        color: #475569;
        margin: 0;
        font-size: 0.9rem;
        line-height: 1.4;
    }

    .btn-primary {
        width: 100%;
        max-width: 280px;
        background: #0f172a;
        color: white;
        border: none;
        padding: 12px;
        border-radius: 14px;
        font-weight: 600;
        font-size: 0.95rem;
        cursor: pointer;
        transition: background 0.15s;
    }

    .btn-text {
        background: transparent;
        border: none;
        color: #ef4444;
        font-weight: 500;
        font-size: 0.85rem;
        cursor: pointer;
        text-decoration: underline;
    }

    .btn-primary:disabled,
    .btn-text:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>