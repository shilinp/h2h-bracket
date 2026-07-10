<script lang="ts">
    interface Props {
        username: string;
        onsubmit: () => void;
    }

    let { 
        username = $bindable(''),
        onsubmit
    }: Props = $props();
</script>

<div class="login-wrapper">
    <div class="content">
        <h1 class="title">
            Welcome to Miranda's<br />
            Sandwich Tournament
        </h1>

        <img 
            src="/login-illustration.svg" 
            alt="Sandwich Tournament" 
            class="illustration" 
        />

        <div class="input-container">
            <input
                type="text"
                bind:value={username}
                placeholder="Enter a username"
                class="username-input"
            />

            <button class="btn-primary" onclick={onsubmit}>
                Create New User
            </button>
        </div>
    </div>
</div>

<style>
    .login-wrapper {
        display: flex;
        flex-direction: column;
        flex-grow: 1;
        width: 100%;
        padding: 2rem;
        box-sizing: border-box;
        overflow-y: auto;
        min-height: 100%;
    }

    .content {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        margin: auto;
        width: 100%;
        max-width: 25rem;
    }

    .title {
        font-size: 1.5rem;
        font-weight: 500;
        text-align: center;
        color: var(--dark-navy-blue);
        margin-top: 0;
        margin-bottom: 2rem;
        line-height: 1.2;
    }

    .illustration {
        width: 100%;
        height: auto;
        max-width: 20rem;
        margin-bottom: 2.5rem;
        display: block;
    }

    .input-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
        gap: 1rem;
    }

    .username-input {
        width: 100%;
        padding: 1rem 1.5rem;
        border-radius: 999px;
        border: 1px solid var(--border);
        background: #FFFFFF;
        color: var(--dark-navy-blue);
        font-size: 1rem;
        font-family: var(--sans);
        box-sizing: border-box;
    }

    .username-input::placeholder {
        color: var(--disabled-content);
    }

    .username-input:focus {
        outline: none;
        border-color: var(--dark-navy-blue);
    }

    .btn-primary {
        background: var(--dark-navy-blue);
        color: white;
        font-weight: 500;
        padding: .75rem 1.5rem;
        border-radius: 999px;
        border: none;
        font-size: 1rem;
        cursor: pointer;
        transition: transform 0.1s ease;
        font-family: var(--sans);
    }

    .btn-primary:active {
        transform: scale(0.98);
    }
</style>