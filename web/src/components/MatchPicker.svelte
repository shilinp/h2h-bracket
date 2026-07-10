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
        <h2 class="picker-heading">Which sandwich is Miranda choosing? 🤔</h2>

        <div class="choice-row">
            <button
                class="choice-pane"
                onclick={() => chooseWinner(currentMatch.team1Id)}
                disabled={isSubmitting}
            >
                <!-- Burger Icon -->
                <svg class="choice-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M3 11c0-2.2 1.8-4 4-4h10c2.2 0 4 1.8 4 4v2H3v-2z" />
                    <path d="M3 17c0 2.2 1.8 4 4 4h10c2.2 0 4-1.8 4-4v-2H3v2z" />
                    <path d="M2 14h20" />
                    <circle cx="7" cy="10" r="0.5" fill="currentColor"/>
                    <circle cx="12" cy="9" r="0.5" fill="currentColor"/>
                    <circle cx="16" cy="10" r="0.5" fill="currentColor"/>
                </svg>
                <span class="team-name">{currentMatch.team1Id != null ? (teamNames[currentMatch.team1Id] ?? 'TBD') : 'TBD'}</span>
            </button>

            <button
                class="choice-pane"
                onclick={() => chooseWinner(currentMatch.team2Id)}
                disabled={isSubmitting}
            >
                <!-- Sandwich Icon -->
                <svg class="choice-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M3 19h14l4-10H7L3 19z" opacity="0.15"/>
                    <path d="M19.3 8.3L15 4.1c-.4-.4-1-.4-1.4 0l-10 10c-.2.2-.3.5-.3.8V19c0 .6.4 1 1 1h4.1c.3 0 .5-.1.7-.3l10.2-10.2c.4-.4.4-1 0-1.2zm-5 1.3l-1.4-1.4 1.4-1.4 1.4 1.4-1.4 1.4z" fill="currentColor"/>
                </svg>
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
        background: #ededf0;
        border: 1px solid #e1e4e8;
        border-radius: 36px;
        padding: 40px 32px;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        width: 100%;
        max-width: 600px;
        aspect-ratio: 1.5 / 1;
        box-sizing: border-box;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }

    .picker-heading {
        color: #0f172a;
        font-size: 1.4rem;
        font-weight: 600;
        text-align: center;
        margin: 0;
    }

    .choice-row {
        display: flex;
        gap: 20px;
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
        gap: 16px;
        border-radius: 28px;
        border: none;
        background: #ffffff;
        color: #0f172a;
        cursor: pointer;
        padding: 20px;
        box-sizing: border-box;
        transition: transform 0.15s ease, box-shadow 0.15s ease;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -1px rgba(0, 0, 0, 0.02);
    }

    .choice-pane:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    }

    .choice-pane:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .choice-icon {
        width: 56px;
        height: 56px;
        color: #0f172a;
        flex-shrink: 0;
    }

    .team-name {
        font-size: 1.15rem;
        font-weight: 600;
        text-align: center;
        line-height: 1.3;
        width: 100%;
        word-wrap: break-word;
        white-space: normal;
        color: #0f172a;
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

    .btn-primary:hover:not(:disabled) {
        background: #1e293b;
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