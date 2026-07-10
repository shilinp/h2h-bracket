<!-- BracketMatchCard.svelte -->
<script lang="ts">
    import type { Match } from "../lib/proto/bracket";

    interface Props {
        match: Match;
        selected: number | undefined;
        masterWinner: number | undefined;
        teamNames: Record<number, string>;
        isLocked: boolean;
        isActiveMatch: boolean;
    }

    let { match, selected, masterWinner, teamNames, isLocked, isActiveMatch }: Props = $props();
</script>

<div class="match-card" class:locked={isLocked} class:active-match={isActiveMatch}>
    <!-- Team 1 Row -->
    <div class="team-row" class:winner={match.team1Id != null && selected === match.team1Id}>
        <div class="team-identity">
            <span class="team-name" class:strikethrough={isLocked && masterWinner != null && selected === match.team1Id && selected !== masterWinner}>
                {match.team1Id != null ? (teamNames[match.team1Id] ?? "TBD") : "TBD"}
            </span>
        </div>
        <div class="status-indicator" class:winner={match.team1Id != null && selected === match.team1Id}>
            {#if match.team1Id != null && selected != null}
                {#if selected === match.team1Id}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" width="16" height="16">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                {:else}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" width="14" height="14">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                {/if}
            {/if}
        </div>
    </div>

    <!-- Team 2 Row -->
    <div class="team-row" class:winner={match.team2Id != null && selected === match.team2Id}>
        <div class="team-identity">
            <span class="team-name" class:strikethrough={isLocked && masterWinner != null && selected === match.team2Id && selected !== masterWinner}>
                {match.team2Id != null ? (teamNames[match.team2Id] ?? "TBD") : "TBD"}
            </span>
        </div>
        <div class="status-indicator" class:winner={match.team2Id != null && selected === match.team2Id}>
            {#if match.team2Id != null && selected != null}
                {#if selected === match.team2Id}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" width="16" height="16">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                {:else}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" width="14" height="14">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                {/if}
            {/if}
        </div>
    </div>

    {#if isLocked && masterWinner != null && selected !== masterWinner}
        <div class="master-override">
            Correct: {teamNames[masterWinner] ?? "TBD"}
        </div>
    {/if}
</div>

<style>
    .match-card {
        background: #2a3249;
        border-radius: 12px;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        border: 2px solid transparent;
        transition: border-color 0.2s ease;
    }
    .match-card.active-match {
        border-color: #6366f1;
        box-shadow: 0 0 12px rgba(99, 102, 241, 0.25);
    }
    .team-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: #232a3f;
        transition: background 0.2s;
    }
    .team-row:first-child {
        background: #181b28;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    }
    .team-row.winner {
        background: #1e2436;
    }
    .team-identity {
        display: flex;
        align-items: center;
        gap: 12px;
    }
    .team-name {
        color: #f8fafc;
        font-size: 0.95rem;
        font-weight: 600;
    }
    .team-name.strikethrough {
        text-decoration: line-through;
        color: #64748b;
    }
    .status-indicator {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        color: #64748b;
    }
    .status-indicator.winner {
        color: #10b981;
    }
    .master-override {
        background: rgba(239, 68, 68, 0.15);
        color: #f87171;
        font-size: 0.75rem;
        text-align: center;
        padding: 6px;
        font-weight: 600;
    }
</style>