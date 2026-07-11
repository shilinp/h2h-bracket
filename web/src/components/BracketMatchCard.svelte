<!-- BracketMatchCard.svelte -->
<script lang="ts">
    import type { Match } from "../lib/proto/bracket";
    import MatchIcon from "./MatchIcon.svelte";

    interface Props {
        match: Match;
        selected: number | undefined;
        masterWinner: number | undefined;
        teamNames: Record<number, string>;
        isLocked: boolean;
        isActiveMatch: boolean;
    }

    let {
        match,
        selected,
        masterWinner,
        teamNames,
        isLocked,
        isActiveMatch,
    }: Props = $props();
</script>

<div
    class="match-card"
    class:locked={isLocked}
    class:active-match={isActiveMatch}
>
    <!-- Team 1 Row -->
    <div
        class="team-row"
        class:winner={match.team1Id != null && selected === match.team1Id}
        class:loser={match.team1Id != null &&
            selected != null &&
            selected !== match.team1Id}
    >
        <div class="team-identity">
            <MatchIcon teamId={match.team1Id} sizePx={24} />
            <span
                class="team-name"
                class:strikethrough={isLocked &&
                    masterWinner != null &&
                    selected === match.team1Id &&
                    selected !== masterWinner}
            >
                {match.team1Id != null
                    ? (teamNames[match.team1Id] ?? "TBD")
                    : "TBD"}
            </span>
        </div>
        <div class="status-indicator">
            {#if match.team1Id != null && selected != null}
                {#if selected === match.team1Id}
                    <img
                        src="/trophy-symbol.svg"
                        alt=""
                        width="24"
                        height="24"
                    />
                {:else}
                    <img src="/x-symbol.svg" alt="" width="24" height="24" />
                {/if}
            {/if}
        </div>
    </div>

    <!-- Team 2 Row -->
    <div
        class="team-row"
        class:winner={match.team2Id != null && selected === match.team2Id}
        class:loser={match.team2Id != null &&
            selected != null &&
            selected !== match.team2Id}
    >
        <div class="team-identity">
            <MatchIcon teamId={match.team2Id} sizePx={24} />
            <span
                class="team-name"
                class:strikethrough={isLocked &&
                    masterWinner != null &&
                    selected === match.team2Id &&
                    selected !== masterWinner}
            >
                {match.team2Id != null
                    ? (teamNames[match.team2Id] ?? "TBD")
                    : "TBD"}
            </span>
        </div>
        <div class="status-indicator">
            {#if match.team2Id != null && selected != null}
                {#if selected === match.team2Id}
                    <img
                        src="/trophy-symbol.svg"
                        alt=""
                        width="24"
                        height="24"
                    />
                {:else}
                    <img src="/x-symbol.svg" alt="" width="24" height="24" />
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
        background: var(--bracket-background);
        border-radius: 12px;
        overflow: hidden;
        display: flex;
        flex-direction: column;

        border: 1px solid var(--bracket-container-border);

        transition: all 0.2s ease;
    }

    .match-card.active-match {
        border: 2px solid var(--dark-navy-blue);

        box-shadow: 0 8px 24px rgba(14, 23, 43, 0.12);
    }
    .team-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem;
        background: #ffffff;
        color: var(--dark-navy-blue);
        transition:
            background-color 0.2s ease,
            color 0.2s ease;
    }
    .team-row:first-child {
        border-bottom: 0.1px solid var(--border);
    }

    .team-row.winner {
        background: var(--neon-yellow-lime);
        color: var(--dark-navy-blue);
    }

    .team-row.loser {
        color: var(--disabled-content);
    }
    .team-identity {
        display: flex;
        align-items: center;
        gap: 14px;
    }
    .team-name {
        font-size: 0.9375rem;
        font-weight: 500;
    }
    .team-name.strikethrough {
        text-decoration: line-through;
        opacity: 0.5;
    }
    .status-indicator {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
    }
    .master-override {
        background: rgba(239, 68, 68, 0.08);
        color: #ef4444;
        font-size: 0.8rem;
        text-align: center;
        padding: 8px;
        font-weight: 600;
        border-top: 1px solid rgba(239, 68, 68, 0.1);
    }
</style>
