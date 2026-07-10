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
            <!-- Universal icon placeholder matching layout -->
            <svg
                class="team-icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z"
                />
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 8v4l3 3"
                />
            </svg>
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
                    <!-- Trophy Icon for Winner -->
                    <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        width="18"
                        height="18"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M6 9H4.5a2.5 2.5 0 010-5H6M18 9h1.5a2.5 2.5 0 000-5H18M4 22h16M10 14.66V17c0 .55-.45 1-1 1H4v2h16v-2h-5c-.55 0-1-.45-1-1v-2.34M12 2a7 7 0 00-7 7c0 2.58 1.4 4.83 3.5 6h7c2.1-1.17 3.5-3.42 3.5-7a7 7 0 00-7-7z"
                        />
                    </svg>
                {:else}
                    <!-- X Icon for Loser -->
                    <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        width="16"
                        height="16"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M6 18L18 6M6 6l12 12"
                        />
                    </svg>
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
            <svg
                class="team-icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z"
                />
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 8v4l3 3"
                />
            </svg>
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
                    <!-- Trophy Icon for Winner -->
                    <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        width="18"
                        height="18"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M6 9H4.5a2.5 2.5 0 010-5H6M18 9h1.5a2.5 2.5 0 000-5H18M4 22h16M10 14.66V17c0 .55-.45 1-1 1H4v2h16v-2h-5c-.55 0-1-.45-1-1v-2.34M12 2a7 7 0 00-7 7c0 2.58 1.4 4.83 3.5 6h7c2.1-1.17 3.5-3.42 3.5-7a7 7 0 00-7-7z"
                        />
                    </svg>
                {:else}
                    <!-- X Icon for Loser -->
                    <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        width="16"
                        height="16"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M6 18L18 6M6 6l12 12"
                        />
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
    /* BracketMatchCard.svelte */

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
        padding: .75rem;
        background: #ffffff;
        color: var(--dark-navy-blue);
        transition:
            background-color 0.2s ease,
            color 0.2s ease;
    }
    .team-row:first-child {
        border-bottom: .1px solid var(--border);
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
    .team-icon {
        width: 20px;
        height: 20px;
        opacity: 0.7;
    }
    .team-row.loser .team-icon {
        opacity: 0.3;
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
