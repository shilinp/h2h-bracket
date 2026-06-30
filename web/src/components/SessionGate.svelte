<script lang="ts">
    import { createEventDispatcher } from 'svelte';

    interface Props {
        username: string;
    
        onsubmit: () => void;
    }

    let { 
        username = '',
        onsubmit
    }: Props = $props();

    const dispatch = createEventDispatcher<{
        login: void;
        specialToggle: void;
    }>();
</script>

<div class="card center-flow">
    <h1 class="title">🏆 Bracket Pro</h1>
    <p class="subtitle">
        {'Enter your unique handle to configure your predictions.'}
    </p>

    <input
        type="text"
        bind:value={username}
        placeholder="Username"
        class="mobile-input"
    />

    <button class="btn-primary" onclick={onsubmit}>
        Load Tournament
    </button>
</div>

<style>
    .center-flow {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        text-align: center;
        height: 100%;
        padding: 24px;
        box-sizing: border-box;
    }

    .card {
        background: #1e293b;
        border-radius: 28px;
        margin: auto 20px;
        padding: 40px 24px;
        box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.4);
        width: 100%;
        max-width: 420px;
    }

    .title {
        font-size: 2rem;
        font-weight: 800;
        margin-bottom: 8px;
    }

    .subtitle {
        color: #94a3b8;
        margin-bottom: 32px;
        font-size: 0.95rem;
    }

    .mobile-input {
        width: 100%;
        padding: 18px;
        border-radius: 14px;
        border: 2px solid #334155;
        background: #0f172a;
        color: white;
        font-size: 1.1rem;
        box-sizing: border-box;
        margin-bottom: 20px;
        text-align: center;
    }

    .mobile-input:focus {
        outline: none;
        border-color: #3b82f6;
    }

    .btn-primary {
        width: 100%;
        background: #2563eb;
        color: white;
        font-weight: 600;
        padding: 18px;
        border-radius: 14px;
        border: none;
        font-size: 1.1rem;
        cursor: pointer;
        transition: transform 0.1s ease;
    }

    .btn-primary:active {
        transform: scale(0.98);
    }
</style>
