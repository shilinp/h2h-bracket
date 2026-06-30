<script lang="ts">
    import type { Match } from '../lib/proto/bracket';

    interface Props {
        rounds: { round: number; matches: Match[] }[];
        predictions: Record<number, number>;
        teamNames: Record<number, string>;
        masterPredictions: Record<number, number>;
        isLocked: boolean;
        activeMatchId: number | null;
    }

    let { 
        rounds = [],
        predictions = {},
        teamNames = {},
        masterPredictions = {},
        isLocked = false,
        activeMatchId = null
    }: Props = $props();
</script>

<div class="preview-shell">
    <div class="preview-header">
        <h2>Bracket Preview</h2>
        <p>
            {isLocked
                ? "Locked bracket view showing your selected winners and locked results."
                : "Watch your bracket fill out as you choose winners."}
        </p>
    </div>

    <div class="round-grid">
        {#each rounds as group}
            <div class="round-column">
                <div class="round-label">Round {group.round}</div>
                {#each group.matches as match}
                    {@const selected = predictions[match.matchId]}
                    {@const masterWinner = masterPredictions[match.matchId]}
                    <div
                        class="match-card"
                        class:selected={selected != null && !isLocked}
                        class:active={activeMatchId === match.matchId}
                        class:locked={isLocked}
                    >
                        <div class="match-teams">
                            <span class="team-name">{match.team1Id != null ? (teamNames[match.team1Id] ?? match.team1Id) : "TBD"}</span>
                            <span class="vs-text">vs</span>
                            <span class="team-name">{match.team2Id != null ? (teamNames[match.team2Id] ?? match.team2Id) : "TBD"}</span>
                        </div>
                        <div class="match-result">
                            {#if isLocked && masterWinner != null}
                                {@const isCorrect = selected === masterWinner}
                                <div class="locked-results">
                                    <span class="master-pick">Result: {teamNames[masterWinner] ?? masterWinner}</span>
                                    {#if selected != null}
                                        <span class="user-pick" class:correct={isCorrect} class:incorrect={!isCorrect}>
                                            Pick: {teamNames[selected] ?? selected}
                                        </span>
                                    {/if}
                                </div>
                            {:else if selected != null}
                                <span class="result-tag">Picked: {teamNames[selected] ?? selected}</span>
                            {:else}
                                <span class="unresolved">Pending</span>
                            {/if}
                        </div>
                    </div>
                {/each}
            </div>
        {/each}
    </div>
</div>

<style>
    .preview-shell {
        background: #111827;
        border-radius: 28px;
        padding: 20px;
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 16px;
        overflow: hidden;
    }

    .preview-header {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .preview-header h2 {
        margin: 0;
        font-size: 1.3rem;
    }

    .preview-header p {
        margin: 0;
        color: #94a3b8;
        font-size: 0.95rem;
    }

    .round-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 14px;
        width: 100%;
    }

    .round-column {
        background: #0f172a;
        border: 1px solid #1e293b;
        border-radius: 22px;
        padding: 12px;
        min-width: 120px;
    }

    .round-label {
        font-weight: 700;
        margin-bottom: 10px;
        color: #e2e8f0;
    }

    .match-card {
        background: #111827;
        border: 1px solid #334155;
        border-radius: 16px;
        padding: 12px;
        margin-bottom: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
        transition: border-color 0.2s ease, transform 0.2s ease, background 0.2s ease;
    }

    .match-card.selected {
        border-color: #22c55e;
        background: rgba(34, 197, 94, 0.08);
    }

    .match-card.active {
        transform: translateY(-1px);
        box-shadow: 0 14px 30px -18px rgba(34, 197, 94, 0.7);
    }

    .match-card.locked {
        background: rgba(15, 23, 42, 0.4);
    }

    .match-teams {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-weight: 700;
        color: #f8fafc;
        font-size: 0.95rem;
    }

    .team-name {
        text-align: center;
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .vs-text {
        color: #94a3b8;
        margin: 0 8px;
        font-size: 0.8rem;
    }

    .match-result {
        display: flex;
        justify-content: center;
        align-items: center;
        font-size: 0.85rem;
        color: #cbd5e1;
    }

    .result-tag,
    .unresolved {
        display: inline-flex;
        align-items: center;
        padding: 6px 10px;
        border-radius: 999px;
        border: 1px solid transparent;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .result-tag {
        color: #d1fae5;
        background: rgba(16, 185, 129, 0.16);
        border-color: rgba(16, 185, 129, 0.3);
    }

    .unresolved {
        color: #fbbf24;
        background: rgba(245, 158, 11, 0.14);
    }

    .locked-results {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 4px;
        width: 100%;
        background: rgba(0, 0, 0, 0.2);
        padding: 8px;
        border-radius: 8px;
    }

    .master-pick {
        font-weight: 700;
        color: #e2e8f0;
    }

    .user-pick {
        font-size: 0.8rem;
        font-weight: 600;
    }

    .user-pick.correct {
        color: #4ade80;
    }

    .user-pick.incorrect {
        color: #f87171;
        text-decoration: line-through;
    }
</style>