<script lang="ts">
    import type { Match } from '../lib/proto/bracket';

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
        <div class="header-meta">
            <!-- Interpret 0-indexed roundNumber safely as 1-indexed visual label -->
            <span class="badge">Round {(currentMatch.roundNumber ?? 0) + 1}</span>
            <span class="counter">Unresolved picks: {remainingCount}</span>
        </div>

        <div class="choice-row">
            <button
                class="choice-pane team-a"
                onclick={() => chooseWinner(currentMatch.team1Id)}
                disabled={isSubmitting}
            >
                <span class="team-name">{currentMatch.team1Id != null ? (teamNames[currentMatch.team1Id] ?? 'TBD') : 'TBD'}</span>
                <div class="tap-indicator">Tap to choose</div>
            </button>
            <div class="divider-vs">
                <span>VS</span>
            </div>
            <button
                class="choice-pane team-b"
                onclick={() => chooseWinner(currentMatch.team2Id)}
                disabled={isSubmitting}
            >
                <span class="team-name">{currentMatch.team2Id != null ? (teamNames[currentMatch.team2Id] ?? 'TBD') : 'TBD'}</span>
                <div class="tap-indicator">Tap to choose</div>
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
        background: #1e293b;
        border-radius: 20px;
        padding: 14px;
        display: flex;
        flex-direction: column;
        gap: 12px;
        width: 100%;
        box-sizing: border-box;
    }

    .header-meta {
        display: flex;
        justify-content: space-between;
        gap: 12px;
        align-items: center;
        color: #cbd5e1;
        font-size: 0.85rem;
    }

    .badge {
        background: #0f172a;
        padding: 6px 12px;
        border-radius: 999px;
        font-weight: 700;
    }

    .counter {
        color: #94a3b8;
    }

    .choice-row {
        display: flex;
        gap: 8px;
        width: 100%;
        box-sizing: border-box;
    }

    .choice-pane {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        min-height: 100px;
        border-radius: 12px;
        border: 1px solid #334155;
        background: #111827;
        color: white;
        cursor: pointer;
        padding: 12px;
        box-sizing: border-box;
        overflow: hidden;
        transition: border-color 0.15s ease, background 0.15s ease;
    }

    .choice-pane:hover:not(:disabled) {
        border-color: #2563eb;
    }

    .choice-pane:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .team-a {
        background: linear-gradient(180deg, rgba(37, 99, 235, 0.12), #111827);
    }

    .team-b {
        background: linear-gradient(180deg, rgba(16, 185, 129, 0.12), #111827);
    }

    .team-name {
        font-size: 1.05rem;
        font-weight: 800;
        text-align: center;
        line-height: 1.2;
        width: 100%;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .tap-indicator {
        margin-top: 6px;
        color: #94a3b8;
        font-size: 0.75rem;
    }

    .divider-vs {
        display: flex;
        align-items: center;
        justify-content: center;
        min-width: 32px;
        color: #64748b;
        font-weight: 800;
        font-size: 0.75rem;
    }

    .picker-empty {
        background: #0f172a;
        border-radius: 14px;
        padding: 16px;
        text-align: center;
        color: #cbd5e1;
        display: flex;
        flex-direction: column;
        gap: 12px;
        align-items: center;
        box-sizing: border-box;
        width: 100%;
    }

    .success-icon {
        font-size: 1.75rem;
    }

    .picker-empty h2 {
        margin: 0;
        font-size: 1.1rem;
    }

    .picker-empty p {
        color: #94a3b8;
        margin: 0;
        font-size: 0.85rem;
        line-height: 1.3;
    }

    .btn-primary {
        width: 100%;
        background: #059669;
        color: white;
        border: none;
        padding: 14px;
        border-radius: 10px;
        font-weight: 600;
        font-size: 0.95rem;
        cursor: pointer;
    }

    .btn-text {
        background: transparent;
        border: none;
        color: #f87171;
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