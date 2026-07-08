<script lang="ts">
    import type { Match, MatchPosition } from '../lib/proto/bracket';

    interface Props {
        rounds: { round: number; matches: Match[] }[];
        predictions: Record<number, number>;
        teamNames: Record<number, string>;
        masterPredictions: Record<number, number>;
        matchPositions?: Record<number, MatchPosition>;
        isLocked: boolean;
        activeMatchId: number | null;
    }

    let { 
        rounds = [],
        predictions = {},
        teamNames = {},
        masterPredictions = {},
        matchPositions = {},
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
                <!-- Visual conversion from 0-indexed round to 1-indexed view -->
                <div class="round-label">Round {group.round + 1}</div>
                <div class="column-matches">
                    {#each group.matches as match}
                        {@const selected = predictions[match.matchId]}
                        {@const masterWinner = masterPredictions[match.matchId]}
                        <div
                            class="match-card"
                            class:selected={selected != null && !isLocked}
                            class:active={activeMatchId === match.matchId}
                            class:locked={isLocked}
                        >
                            <!-- Traditional Bracket Connective UI Structure -->
                            <div class="bracket-node-team team-top">
                                <span class="team-name" class:strikethrough={isLocked && masterWinner != null && selected === match.team1Id && selected !== masterWinner}>
                                    {match.team1Id != null ? (teamNames[match.team1Id] ?? 'TBD') : "TBD"}
                                </span>
                                {#if selected != null && selected === match.team1Id && !isLocked}
                                    <span class="indicator-dot"></span>
                                {/if}
                            </div>
                            
                            <div class="bracket-connector-rail">
                                <span class="vs-text">VS</span>
                            </div>

                            <div class="bracket-node-team team-bottom">
                                <span class="team-name" class:strikethrough={isLocked && masterWinner != null && selected === match.team2Id && selected !== masterWinner}>
                                    {match.team2Id != null ? (teamNames[match.team2Id] ?? 'TBD') : "TBD"}
                                </span>
                                {#if selected != null && selected === match.team2Id && !isLocked}
                                    <span class="indicator-dot"></span>
                                {/if}
                            </div>

                            {#if isLocked && masterWinner != null}
                                {@const isCorrect = selected === masterWinner}
                                <div class="master-override-pane" class:match-match={isCorrect} class:match-mismatch={!isCorrect}>
                                    <span class="override-label">Winner: {teamNames[masterWinner] ?? 'TBD'}</span>
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </div>
        {/each}
    </div>
</div>

<style>
    .preview-shell {
        background: #111827;
        border-radius: 20px;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 16px;
        width: 100%;
        box-sizing: border-box;
    }

    .preview-header {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .preview-header h2 {
        margin: 0;
        font-size: 1.15rem;
        font-weight: 700;
    }

    .preview-header p {
        margin: 0;
        color: #94a3b8;
        font-size: 0.85rem;
        line-height: 1.3;
    }

    .round-grid {
        display: flex;
        flex-direction: column;
        gap: 20px;
        width: 100%;
    }

    .round-column {
        display: flex;
        flex-direction: column;
        gap: 10px;
        width: 100%;
    }

    .round-label {
        font-weight: 800;
        font-size: 0.85rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: #3b82f6;
    }

    .column-matches {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .match-card {
        background: #0f172a;
        border: 1px solid #1e293b;
        border-radius: 12px;
        padding: 0;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        transition: border-color 0.2s ease, box-shadow 0.2s ease;
    }

    .match-card.active {
        border-color: #3b82f6;
        box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
    }

    .match-card.selected {
        border-color: #10b981;
    }

    .bracket-node-team {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 10px 14px;
        background: #1e293b;
        height: 40px;
        box-sizing: border-box;
    }

    .team-top {
        border-bottom: 1px solid #0f172a;
    }

    .team-name {
        font-size: 0.9rem;
        font-weight: 600;
        color: #f8fafc;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        flex: 1;
        padding-right: 8px;
    }

    .team-name.strikethrough {
        text-decoration: line-through;
        color: #ef4444;
        opacity: 0.8;
    }

    .indicator-dot {
        width: 6px;
        height: 6px;
        border-radius: 555px;
        background-color: #10b981;
        flex-shrink: 0;
    }

    .bracket-connector-rail {
        display: flex;
        align-items: center;
        background: #0f172a;
        height: 20px;
        padding: 0 14px;
    }

    .vs-text {
        font-size: 0.7rem;
        font-weight: 700;
        color: #475569;
        letter-spacing: 0.05em;
    }

    .master-override-pane {
        padding: 8px 14px;
        font-size: 0.8rem;
        font-weight: 700;
        display: flex;
        align-items: center;
    }

    .master-override-pane.match-match {
        background: rgba(16, 185, 129, 0.15);
        color: #34d399;
        border-top: 1px solid rgba(16, 185, 129, 0.2);
    }

    .master-override-pane.match-mismatch {
        background: rgba(239, 68, 68, 0.1);
        color: #f87171;
        border-top: 1px solid rgba(239, 68, 68, 0.15);
    }

    .override-label {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
</style>