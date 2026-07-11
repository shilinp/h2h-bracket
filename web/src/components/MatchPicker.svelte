<script lang="ts">
    import type { Match } from "../lib/proto/bracket";
    import BreadLoader from "./BreadLoader.svelte";
    import MatchIcon from "./MatchIcon.svelte";

    interface Props {
        currentMatch?: (Match & { roundNumber?: number }) | null;
        teamNames?: Record<number, string>;
        isSubmitting?: boolean;
        hasPersistedBracket: boolean;
        onselect: (event: { matchId: number; winnerId: number }) => void;
        onsubmit: () => void;
        onreset: () => void;
    }

    let {
        currentMatch = null,
        teamNames = {},
        isSubmitting = false,
        hasPersistedBracket,
        onselect,
        onsubmit,
        onreset,
    }: Props = $props();

    let activeTeamId: number | null = $state(null);
    let submittingDelay: boolean = $state(false);

    function handleSelect(teamId: number | undefined | null) {
        if (!currentMatch || teamId == null || isSubmitting) return;

        activeTeamId = teamId;

        setTimeout(() => {
            activeTeamId = null;
            chooseWinner(teamId);
        }, 100);
    }

    function handleSubmit() {
        submittingDelay = true;
        setTimeout(() => {
            submittingDelay = false;
            onsubmit();
        }, 100);
    }

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
                class="choice-pane {activeTeamId === currentMatch.team1Id
                    ? 'active'
                    : ''}"
                onclick={() => handleSelect(currentMatch.team1Id)}
                disabled={isSubmitting}
            >
                <MatchIcon teamId={currentMatch.team1Id} sizePx={48} />
                <span class="team-name"
                    >{currentMatch.team1Id != null
                        ? (teamNames[currentMatch.team1Id] ?? "TBD")
                        : "TBD"}</span
                >
            </button>

            <button
                class="choice-pane {activeTeamId === currentMatch.team2Id
                    ? 'active'
                    : ''}"
                onclick={() => handleSelect(currentMatch.team2Id)}
                disabled={isSubmitting}
            >
                <MatchIcon teamId={currentMatch.team2Id} sizePx={48} />
                <span class="team-name"
                    >{currentMatch.team2Id != null
                        ? (teamNames[currentMatch.team2Id] ?? "TBD")
                        : "TBD"}</span
                >
            </button>
        </div>
    {:else if isSubmitting}
        <div class="loader-container">
            <BreadLoader showText={false} size={180} />
        </div>
    {:else}
        <div class="picker-empty">
            {#if hasPersistedBracket}
                <span class="picker-empty-text"
                    >We got your submission big dawg, good luck &lt;3</span
                >
                <img src="/cool-guy-sandwich.svg" alt="" class="cool-guy"/> 
                <button class="btn-reset" onclick={onreset}> Reset </button>
            {:else}
                <span class="picker-empty-text">Locked in?</span>
                <button
                    class="btn-submit {submittingDelay ? 'clicked' : ''}"
                    onclick={handleSubmit}
                    >Submit Your Bracket
                </button>
                <button class="btn-reset" onclick={onreset}> Reset </button>
            {/if}
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
        transition:
            transform 0.15s ease,
            box-shadow 0.15s ease;
    }

    .choice-pane.active {
        background: var(--neon-yellow-lime);
        border-color: var(--dark-navy-blue);
        transform: scale(0.98);
    }

    .team-name {
        font-size: 1rem;
        text-align: center;
        width: 100%;
        overflow-wrap: break-word;
        color: var(--dark-navy-blue);
    }

    .picker-empty {
        text-align: center;
        color: var(--dark-navy-blue);
        font-size: 1rem;
        display: flex;
        flex-direction: column;
        align-items: center;
        box-sizing: border-box;
        width: 100%;
        height: 100%;
        padding-left: 2rem;
        padding-right: 2rem;
    }

    .picker-empty-text {
        margin-top: auto;
        margin-bottom: 1rem;
    }

    .btn-submit {
        background: var(--dark-navy-blue);
        color: white;
        font-weight: 500;
        padding: 0.75rem 1.5rem;
        border-radius: 999px;
        border: none;
        font-size: 1rem;
        transition: transform 0.1s ease;
        font-family: var(--sans);
    }

    .btn-submit.clicked {
        background: var(--neon-yellow-lime);
        color: var(--dark-navy-blue);
        transform: scale(0.98);
    }

    .btn-reset {
        background: transparent;
        border: none;
        color: var(--dark-navy-blue);
        font-weight: 500;
        font-size: 1rem;
        margin-top: auto;
        margin-bottom: 0;
    }

    .loader-container {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        height: 100%;
    }

    .cool-guy {
        padding-bottom: 1rem
    }
</style>
