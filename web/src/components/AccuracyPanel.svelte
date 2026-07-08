<script lang="ts">
    interface Props {
        accuracy: number | null;
        isSubmitting: boolean;
        onreset: () => void;
    }

    let { accuracy = null, isSubmitting = false, onreset }: Props = $props();
</script>

<div class="accuracy-panel">
    <div class="score-card">
        <span class="lock-icon">🔒</span>
        <h2>Submissions Locked</h2>
        <p>Matches are already underway. Modifications are suspended.</p>
        {#if accuracy !== null}
            <p class="accuracy-score">
                Prediction Score: {accuracy.toFixed(0)}
            </p>
        {/if}
    </div>
</div>

<style>
    .accuracy-panel {
        background: #1e293b;
        border-radius: 28px;
        width: 100%;
        display: flex;
        justify-content: center;
        align-items: center;
        min-height: 220px;
    }

    .score-card {
        display: flex;
        padding: 24px;
        flex-direction: column;
        align-items: center;
        gap: 16px;
        width: 100%;
        max-width: 420px;
        text-align: center;
    }

    .lock-icon {
        font-size: 2rem;
    }

    h2 {
        margin: 0;
        font-size: 1.4rem;
    }

    p {
        color: #cbd5e1;
        margin: 0;
    }

    .accuracy-score {
        margin-top: 8px;
        color: #d1fae5;
        font-size: 1.05rem;
        font-weight: 700;
    }

    .btn-secondary {
        margin-top: 16px;
        background: #475569;
        color: #f8fafc;
        border: none;
        padding: 14px 28px;
        border-radius: 12px;
        font-weight: 500;
        cursor: pointer;
    }

    .btn-secondary:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
</style>
